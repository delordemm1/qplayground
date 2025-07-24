package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	baseDir string
}

func NewLocalFileStorage(baseDir string) *LocalFileStorage {
	return &LocalFileStorage{
		baseDir: baseDir,
	}
}

func (l *LocalFileStorage) Upload(ctx context.Context, key string, data io.Reader, options *UploadOptions) error {
	// Ensure the directory exists
	fullPath := filepath.Join(l.baseDir, key)
	dir := filepath.Dir(fullPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fullPath, err)
	}
	defer file.Close()

	// Copy data to file
	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %w", fullPath, err)
	}

	slog.Info("Successfully uploaded file to local storage", "key", key, "path", fullPath)
	return nil
}

func (l *LocalFileStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(l.baseDir, key)
	
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", fullPath, err)
	}

	slog.Info("Successfully deleted file from local storage", "key", key, "path", fullPath)
	return nil
}

func (l *LocalFileStorage) GetPublicURL(key string) string {
	// Return relative path for use in reports
	return key
}