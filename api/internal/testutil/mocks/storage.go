package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockStorageBackend is a testify mock implementing storage.Backend.
type MockStorageBackend struct {
	mock.Mock
}

func (m *MockStorageBackend) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	args := m.Called(ctx, key, r, size, contentType)
	return args.Error(0)
}

func (m *MockStorageBackend) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockStorageBackend) PresignURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockStorageBackend) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorageBackend) Type() string {
	args := m.Called()
	return args.String(0)
}

// NoopStorageBackend is a zero-value storage backend for tests that don't exercise file I/O.
type NoopStorageBackend struct{}

func (NoopStorageBackend) Upload(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
	return nil
}
func (NoopStorageBackend) Delete(_ context.Context, _ string) error       { return nil }
func (NoopStorageBackend) PresignURL(_ context.Context, _ string) (string, error) { return "", nil }
func (NoopStorageBackend) Open(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, nil
}
func (NoopStorageBackend) Type() string { return "noop" }
