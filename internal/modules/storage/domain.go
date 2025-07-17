package storage

import (
	"context"
	"io"
)

// UploadOptions contains options for uploading files
type UploadOptions struct {
	ContentType string
	Metadata    map[string]string
}

// ObjectStorage defines the interface for object storage operations
type ObjectStorage interface {
	Upload(ctx context.Context, key string, data io.Reader, options *UploadOptions) error
	Delete(ctx context.Context, key string) error
	GetPublicURL(key string) string
}

// StorageService provides high-level storage operations
type StorageService interface {
	UploadFile(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
}