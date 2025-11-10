// Package storage provides storage interfaces and implementations.
// For local development, use LocalStorageClient.
// For production, use S3Client (requires AWS SDK).
package storage

import (
	"context"
	"io"
	"time"
)

// StorageClientInterface defines the contract for storage operations.
type StorageClientInterface interface {
	Upload(ctx context.Context, file io.Reader, key string, contentType string) (string, error)
	GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// Note: S3Client implementation requires AWS SDK dependencies.
// For local development, use LocalStorageClient instead.
// Uncomment and add AWS SDK dependencies when ready for production:
/*
import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// StorageClient implements S3/MinIO storage operations.
type StorageClient struct {
	client     *s3.Client
	bucket     string
	endpoint   string // For MinIO compatibility
	useSSL     bool
	region     string
}

// StorageConfig holds configuration for storage client.
type StorageConfig struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Optional: for MinIO (e.g., "http://localhost:9000")
	UseSSL          bool   // Default: true for S3, false for MinIO
}

// NewStorageClient creates a new storage client (S3 or MinIO).
func NewStorageClient(cfg StorageConfig) (*StorageClient, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Configure S3 client with optional MinIO endpoint
	s3Options := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Options...)

	return &StorageClient{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
		useSSL:   cfg.UseSSL,
		region:   cfg.Region,
	}, nil
}

// Upload uploads a file to S3/MinIO and returns the object key.
func (s *StorageClient) Upload(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return key, nil
}

// GetSignedURL generates a presigned URL for accessing an object.
func (s *StorageClient) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return request.URL, nil
}

// Delete deletes an object from storage.
func (s *StorageClient) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
*/

