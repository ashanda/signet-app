package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"signet-backend/internal/config"
)

func TestCleanKey(t *testing.T) {
	for _, ok := range []string{"kyc/nic_front/ab12.jpg", "kyc/passport/x"} {
		if _, err := CleanKey(ok); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "/kyc/a.jpg", "kyc/../go.mod", "kyc//a", "kyc/./a", "..", "kyc\\..\\a", "kyc/a\x00"} {
		if _, err := CleanKey(bad); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestLocal(t *testing.T) {
	ctx := context.Background()
	l := &Local{Root: t.TempDir()}
	if err := l.Put(ctx, "kyc/nic_front/a.jpg", strings.NewReader("img"), 3, "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	rc, info, err := l.Open(ctx, "kyc/nic_front/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(rc)
	rc.Close()
	if string(b) != "img" || info.Size != 3 || info.ContentType != "image/jpeg" {
		t.Fatalf("got %q %+v", b, info)
	}
	if _, _, err := l.Open(ctx, "kyc/nic_front/missing.jpg"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if _, _, err := l.Open(ctx, "kyc/nic_front"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("directory: %v", err)
	}
	if _, _, err := l.Open(ctx, "kyc/../../etc/passwd"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("traversal: %v", err)
	}
}

// fakeS3 is a minimal path-style S3: PUT/GET/HEAD objects in one bucket.
type fakeS3 struct {
	mu      sync.Mutex
	bucket  string
	objects map[string][]byte
	types   map[string]string
	sse     map[string]string
}

func (f *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	if parts[0] != f.bucket {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `<Error><Code>NoSuchBucket</Code><Message>no bucket</Message></Error>`)
		return
	}
	if len(parts) == 1 || parts[1] == "" { // HeadBucket
		w.WriteHeader(http.StatusOK)
		return
	}
	key := parts[1]
	switch r.Method {
	case http.MethodPut:
		b, _ := io.ReadAll(r.Body)
		f.objects[key] = b
		f.types[key] = r.Header.Get("Content-Type")
		f.sse[key] = r.Header.Get("X-Amz-Server-Side-Encryption")
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		b, ok := f.objects[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `<Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`)
			return
		}
		w.Header().Set("Content-Type", f.types[key])
		w.Write(b)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func newFakeS3(t *testing.T, bucket string) (*fakeS3, *S3) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	fake := &fakeS3{bucket: "signet-test", objects: map[string][]byte{}, types: map[string]string{}, sse: map[string]string{}}
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	s, err := NewS3(context.Background(), &config.Config{
		AWSBucket: bucket, AWSRegion: "us-east-1", AWSEndpoint: srv.URL, AWSUsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return fake, s
}

func TestS3(t *testing.T) {
	ctx := context.Background()
	fake, s := newFakeS3(t, "signet-test")

	if err := s.Check(ctx); err != nil {
		t.Fatalf("check: %v", err)
	}
	data := []byte("\xff\xd8\xffjpeg-bytes")
	if err := s.Put(ctx, "kyc/nic_back/b.jpg", bytes.NewReader(data), int64(len(data)), "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fake.objects["kyc/nic_back/b.jpg"], data) || fake.sse["kyc/nic_back/b.jpg"] != "AES256" {
		t.Fatalf("stored %q sse=%q", fake.objects["kyc/nic_back/b.jpg"], fake.sse["kyc/nic_back/b.jpg"])
	}

	rc, info, err := s.Open(ctx, "kyc/nic_back/b.jpg")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(rc)
	rc.Close()
	if !bytes.Equal(got, data) || info.ContentType != "image/jpeg" {
		t.Fatalf("read %q %+v", got, info)
	}

	if _, _, err := s.Open(ctx, "kyc/nic_back/missing.jpg"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing key: %v", err)
	}
	if err := s.Put(ctx, "../escape", bytes.NewReader(nil), 0, ""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("bad key: %v", err)
	}
}

// A wrong bucket must surface as an error, never as "not found" (which
// would silently fall back to local disk).
func TestS3WrongBucket(t *testing.T) {
	ctx := context.Background()
	_, s := newFakeS3(t, "wrong-bucket")
	if _, _, err := s.Open(ctx, "kyc/a.jpg"); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("want a real error, got %v", err)
	}
	if err := Check(ctx, &Fallback{Primary: s, Secondary: &Local{Root: t.TempDir()}}); err == nil {
		t.Fatal("Check passed for a missing bucket")
	}
}

func TestFallback(t *testing.T) {
	ctx := context.Background()
	fake, s := newFakeS3(t, "signet-test")
	local := &Local{Root: t.TempDir()}
	local.Put(ctx, "kyc/passport/old.png", strings.NewReader("old"), 3, "")
	f := &Fallback{Primary: s, Secondary: local}

	rc, _, err := f.Open(ctx, "kyc/passport/old.png")
	if err != nil {
		t.Fatalf("old local file not served: %v", err)
	}
	b, _ := io.ReadAll(rc)
	rc.Close()
	if string(b) != "old" {
		t.Fatalf("got %q", b)
	}

	if err := f.Put(ctx, "kyc/passport/new.png", strings.NewReader("new"), 3, "image/png"); err != nil {
		t.Fatal(err)
	}
	if _, ok := fake.objects["kyc/passport/new.png"]; !ok {
		t.Fatal("new upload did not go to S3")
	}
	if _, _, err := local.Open(ctx, "kyc/passport/new.png"); !errors.Is(err, ErrNotFound) {
		t.Fatal("new upload was written to local disk")
	}
	if _, _, err := f.Open(ctx, "kyc/passport/none.png"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing everywhere: %v", err)
	}
}
