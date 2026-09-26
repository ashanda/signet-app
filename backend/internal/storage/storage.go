// Package storage stores uploaded files (today: KYC images) under
// slash-separated keys such as "kyc/nic_front/<hex>.jpg" — the same
// relative path the database already records. The key never depends on
// the backend, so switching STORAGE_DRIVER needs no data migration beyond
// copying the files.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"signet-backend/internal/config"
)

// ErrNotFound is returned by Open when the key does not exist.
var ErrNotFound = errors.New("storage: not found")

// ErrInvalidKey is returned for keys that could escape their root.
var ErrInvalidKey = errors.New("storage: invalid key")

type Info struct {
	Size        int64
	ContentType string
	ModTime     time.Time
}

type Store interface {
	Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, Info, error)
	Name() string
}

// New builds the store selected by STORAGE_DRIVER. In s3 mode, reads fall
// back to the local directory for files uploaded before the switch, so
// existing images keep working while (and after) they are copied over.
func New(ctx context.Context, cfg *config.Config) (Store, error) {
	local := &Local{Root: cfg.StorageLocalRoot}
	switch cfg.StorageDriver {
	case "", "local":
		return local, nil
	case "s3":
		s3, err := NewS3(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return &Fallback{Primary: s3, Secondary: local}, nil
	default:
		return nil, fmt.Errorf("unknown STORAGE_DRIVER %q (use local or s3)", cfg.StorageDriver)
	}
}

// Check confirms a remote store is reachable with the configured
// credentials; local storage always passes.
func Check(ctx context.Context, s Store) error {
	if f, ok := s.(*Fallback); ok {
		s = f.Primary
	}
	if s3, ok := s.(*S3); ok {
		if err := s3.Check(ctx); err != nil {
			return fmt.Errorf("cannot reach %s: %w", s3.Name(), err)
		}
	}
	return nil
}

// CleanKey validates a key taken from a URL or the database: relative,
// slash-separated, no empty/dot segments. It returns ErrInvalidKey rather
// than normalising, so "a/../b" is refused instead of silently mapped.
func CleanKey(key string) (string, error) {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") || strings.ContainsRune(key, 0) {
		return "", ErrInvalidKey
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", ErrInvalidKey
		}
	}
	return key, nil
}

// Fallback writes to Primary and reads from Primary, then Secondary when
// Primary does not have the key.
type Fallback struct {
	Primary, Secondary Store
}

func (f *Fallback) Name() string { return f.Primary.Name() + " (fallback: " + f.Secondary.Name() + ")" }

func (f *Fallback) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	return f.Primary.Put(ctx, key, body, size, contentType)
}

func (f *Fallback) Open(ctx context.Context, key string) (io.ReadCloser, Info, error) {
	rc, info, err := f.Primary.Open(ctx, key)
	if errors.Is(err, ErrNotFound) {
		rc, info, err = f.Secondary.Open(ctx, key)
		if err == nil {
			log.Printf("storage: served %s from %s (not in %s yet)", key, f.Secondary.Name(), f.Primary.Name())
		}
	}
	return rc, info, err
}
