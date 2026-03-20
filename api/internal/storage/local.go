package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalBackend stores objects on the local filesystem.
type LocalBackend struct {
	basePath string
}

// NewLocalBackend creates a LocalBackend rooted at basePath.
// The directory is created if it does not exist.
func NewLocalBackend(basePath string) (*LocalBackend, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("storage: create base path: %w", err)
	}
	return &LocalBackend{basePath: basePath}, nil
}

func (b *LocalBackend) Upload(_ context.Context, key string, r io.Reader, _ int64, contentType string) error {
	fullPath := b.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (b *LocalBackend) Delete(_ context.Context, key string) error {
	err := os.Remove(b.fullPath(key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// PresignURL returns empty — local backend serves objects directly via Open.
func (b *LocalBackend) PresignURL(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (b *LocalBackend) Open(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(b.fullPath(key))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (b *LocalBackend) Type() string { return "local" }

// fullPath resolves key to an absolute path under basePath.
// filepath.Clean prevents directory traversal.
func (b *LocalBackend) fullPath(key string) string {
	return filepath.Join(b.basePath, filepath.Clean("/"+key))
}
