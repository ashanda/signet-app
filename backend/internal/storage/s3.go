package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"signet-backend/internal/config"
)

// S3 stores files as private objects in one bucket, keyed exactly like the
// local layout ("kyc/nic_front/<hex>.jpg"). Objects are never public: the
// API streams them back through /storage/kyc/*. Credentials come from the
// standard AWS chain (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY env vars,
// ~/.aws/credentials, or an instance role).
type S3 struct {
	Client *s3.Client
	Bucket string
}

func NewS3(ctx context.Context, cfg *config.Config) (*S3, error) {
	if cfg.AWSBucket == "" {
		return nil, errors.New("STORAGE_DRIVER=s3 needs AWS_BUCKET")
	}
	if cfg.AWSRegion == "" {
		return nil, errors.New("STORAGE_DRIVER=s3 needs AWS_DEFAULT_REGION")
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.AWSUsePathStyle
		if cfg.AWSEndpoint != "" {
			// S3-compatible stores (MinIO, R2, ...) often reject the
			// SDK's default flexible checksums.
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		}
	})
	return &S3{Client: client, Bucket: cfg.AWSBucket}, nil
}

func (s *S3) Name() string { return "s3://" + s.Bucket }

// Check confirms the bucket is reachable with the configured credentials.
func (s *S3) Check(ctx context.Context) error {
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.Bucket)})
	return err
}

func (s *S3) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	key, err := CleanKey(key)
	if err != nil {
		return err
	}
	in := &s3.PutObjectInput{
		Bucket:               aws.String(s.Bucket),
		Key:                  aws.String(key),
		Body:                 body,
		ContentLength:        aws.Int64(size),
		ServerSideEncryption: types.ServerSideEncryptionAes256,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}
	if _, err := s.Client.PutObject(ctx, in); err != nil {
		return fmt.Errorf("s3 put %s: %w", key, err)
	}
	return nil
}

func (s *S3) Open(ctx context.Context, key string) (io.ReadCloser, Info, error) {
	key, err := CleanKey(key)
	if err != nil {
		return nil, Info{}, err
	}
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key)})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound") {
			return nil, Info{}, ErrNotFound
		}
		return nil, Info{}, fmt.Errorf("s3 get %s: %w", key, err)
	}
	info := Info{Size: aws.ToInt64(out.ContentLength), ContentType: aws.ToString(out.ContentType)}
	if out.LastModified != nil {
		info.ModTime = *out.LastModified
	}
	return out.Body, info, nil
}
