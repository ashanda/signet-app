package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
)

// Local stores files under Root (default "storage", relative to the
// process working directory — /opt/signet/backend in production).
type Local struct {
	Root string
}

func (l *Local) Name() string { return "local disk" }

func (l *Local) path(key string) (string, error) {
	key, err := CleanKey(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.Root, filepath.FromSlash(key)), nil
}

func (l *Local) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	dst, err := os.Create(p)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, body); err != nil {
		dst.Close()
		os.Remove(p)
		return err
	}
	return dst.Close()
}

func (l *Local) Open(ctx context.Context, key string) (io.ReadCloser, Info, error) {
	p, err := l.path(key)
	if err != nil {
		return nil, Info{}, err
	}
	f, err := os.Open(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, Info{}, ErrNotFound
	}
	if err != nil {
		return nil, Info{}, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, Info{}, err
	}
	if st.IsDir() {
		f.Close()
		return nil, Info{}, ErrNotFound
	}
	return f, Info{Size: st.Size(), ContentType: mime.TypeByExtension(filepath.Ext(p)), ModTime: st.ModTime()}, nil
}
