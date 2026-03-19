package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Backend stores objects in an S3-compatible service (AWS S3, MinIO, etc.).
type S3Backend struct {
	client *minio.Client
	bucket string
}

// NewS3Backend creates an S3Backend connected to endpoint.
// endpoint should be the host:port (e.g. "s3.amazonaws.com" or "localhost:9000").
func NewS3Backend(endpoint, bucket, accessKey, secretKey string, useSSL bool) (*S3Backend, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3Backend{client: client, bucket: bucket}, nil
}

func (b *S3Backend) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := b.client.PutObject(ctx, b.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (b *S3Backend) Delete(ctx context.Context, key string) error {
	return b.client.RemoveObject(ctx, b.bucket, key, minio.RemoveObjectOptions{})
}

// PresignURL generates a presigned GET URL valid for 15 minutes.
func (b *S3Backend) PresignURL(ctx context.Context, key string) (string, error) {
	u, err := b.client.PresignedGetObject(ctx, b.bucket, key, 15*time.Minute, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Open streams the object directly from S3 (fallback — prefer PresignURL + redirect).
func (b *S3Backend) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return b.client.GetObject(ctx, b.bucket, key, minio.GetObjectOptions{})
}

func (b *S3Backend) Type() string { return "s3" }
