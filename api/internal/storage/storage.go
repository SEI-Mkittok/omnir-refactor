package storage

import (
	"context"
	"io"
)

// Backend abstracts S3-compatible and local file storage.
type Backend interface {
	// Upload stores r under key. size may be -1 if unknown.
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// Delete removes the object at key. Returns nil if not found.
	Delete(ctx context.Context, key string) error
	// PresignURL returns a temporary download URL for the object.
	// Returns an empty string for the local backend (use Open to stream).
	PresignURL(ctx context.Context, key string) (string, error)
	// Open opens the object for reading (used by the local backend).
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Type returns the backend identifier ("s3" or "local").
	Type() string
}
