package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client is the interface for all file storage operations.
// This abstraction allows swapping MinIO with any S3-compatible service.
type Client interface {
	Upload(ctx context.Context, objectName string, file multipart.File, fileSize int64, contentType string) (string, error)
	Delete(ctx context.Context, objectName string) error
	GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

type minioClient struct {
	client     *minio.Client
	bucketName string
	publicURL  string // Base URL for constructing public file links
}

// noopClient is a fallback when MinIO is not available (development mode).
type noopClient struct{}

func (n *noopClient) Upload(_ context.Context, _ string, _ multipart.File, _ int64, _ string) (string, error) {
	return "", fmt.Errorf("storage not available: MinIO is not connected")
}
func (n *noopClient) Delete(_ context.Context, _ string) error {
	return fmt.Errorf("storage not available: MinIO is not connected")
}
func (n *noopClient) GetPresignedURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", fmt.Errorf("storage not available: MinIO is not connected")
}

// NewMinIOClient initializes a MinIO client and ensures the bucket exists.
// Returns a no-op fallback client if MinIO is unreachable (non-fatal for development).
func NewMinIOClient() Client {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	publicURL := os.Getenv("MINIO_PUBLIC_URL")

	if endpoint == "" {
		endpoint = "127.0.0.1:9000"
	}
	if bucketName == "" {
		bucketName = "amos-files"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("⚠️  MinIO client init failed: %v (file storage disabled)", err)
		return &noopClient{}
	}

	// Ensure bucket exists on startup
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Printf("⚠️  MinIO unreachable: %v (file storage disabled)", err)
		return &noopClient{}
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			log.Printf("⚠️  Failed to create MinIO bucket '%s': %v (file storage disabled)", bucketName, err)
			return &noopClient{}
		}
		log.Printf("MinIO bucket '%s' created successfully", bucketName)
	}

	log.Printf("✅ MinIO storage connected: %s/%s", endpoint, bucketName)

	return &minioClient{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}
}

// Upload stores a file in the MinIO bucket and returns the object path.
func (m *minioClient) Upload(ctx context.Context, objectName string, file multipart.File, fileSize int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	// Return the public URL if configured, otherwise just the object name
	if m.publicURL != "" {
		return fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucketName, objectName), nil
	}
	return objectName, nil
}

// Delete removes a file from the MinIO bucket.
func (m *minioClient) Delete(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

// GetPresignedURL generates a temporary download link valid for the given duration.
func (m *minioClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GenerateObjectName creates a unique, organized object path for storage.
// Example: "employees/42/profile/1715443200_photo.jpg"
func GenerateObjectName(entityType string, entityID uint, category string, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%d/%s/%d%s", entityType, entityID, category, timestamp, ext)
}
