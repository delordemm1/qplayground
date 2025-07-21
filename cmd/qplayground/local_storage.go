package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorageService implements a simple file-based storage service
type LocalStorageService struct {
	baseDir string
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService(baseDir string) LocalStorageService {
	return LocalStorageService{
		baseDir: baseDir,
	}
}

// UploadFile saves a file to the local filesystem
func (s LocalStorageService) UploadFile(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	filePath := filepath.Join(s.baseDir, key)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create and write file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// DeleteFile removes a file from the local filesystem
func (s LocalStorageService) DeleteFile(ctx context.Context, key string) error {
	filePath := filepath.Join(s.baseDir, key)
	return os.Remove(filePath)
}

// GetPublicURL returns the local file path
func (s LocalStorageService) GetPublicURL(key string) string {
	return filepath.Join(s.baseDir, key)
}