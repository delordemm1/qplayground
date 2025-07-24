package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type storageService struct {
	storage ObjectStorage
}

func NewStorageService(storage ObjectStorage) StorageService {
	return &storageService{
		storage: storage,
	}
}

func (s *storageService) UploadFile(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	options := &UploadOptions{
		ContentType: contentType,
	}

	err := s.storage.Upload(ctx, key, data, options)
	if err != nil {
		slog.Error("Failed to upload file", "error", err, "key", key)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	publicURL := s.storage.GetPublicURL(key)
	slog.Info("File uploaded successfully", "key", key, "url", publicURL)
	return publicURL, nil
}

func (s *storageService) DeleteFile(ctx context.Context, key string) error {
	err := s.storage.Delete(ctx, key)
	if err != nil {
		slog.Error("Failed to delete file", "error", err, "key", key)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	slog.Info("File deleted successfully", "key", key)
	return nil
}

func (s *storageService) GetPublicURL(key string) string {
	return s.storage.GetPublicURL(key)
}