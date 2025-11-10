package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorageClient implements storage using local filesystem (for development).
type LocalStorageClient struct {
	baseDir string
	baseURL string // Base URL for serving files (e.g., "http://localhost:8000/files")
}

// NewLocalStorageClient creates a new local file storage client.
func NewLocalStorageClient(baseDir string, baseURL string) (*LocalStorageClient, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageClient{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

// Upload uploads a file to local storage and returns the file path.
func (s *LocalStorageClient) Upload(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	// Create full path
	fullPath := filepath.Join(s.baseDir, key)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(fullPath) // Clean up on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return key, nil
}

// GetSignedURL generates a URL for accessing a file (for local dev, just returns the URL).
func (s *LocalStorageClient) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// For local development, just return the base URL + key
	// In production, this would generate a presigned URL
	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}

// Delete deletes a file from local storage.
func (s *LocalStorageClient) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, consider it deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

