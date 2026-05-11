package system

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"amos-backend/pkg/storage"
)

// Service defines the interface for file operations (upload, download, delete).
type Service interface {
	UploadFile(ctx context.Context, entityType string, entityID uint, category string, fileHeader *multipart.FileHeader) (*File, error)
	GetFilesByEntity(entityType string, entityID uint) ([]File, error)
	GetFilesByEntityAndCategory(entityType string, entityID uint, category string) ([]File, error)
	GetFileDownloadURL(ctx context.Context, fileID uint) (string, error)
	DeleteFile(ctx context.Context, fileID uint) error
}

type service struct {
	repo          Repository
	storageClient storage.Client
}

// NewService creates a new file service with the given repository and storage client.
func NewService(repo Repository, storageClient storage.Client) Service {
	return &service{
		repo:          repo,
		storageClient: storageClient,
	}
}

// allowedMimeTypes defines the permitted file types for upload.
var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// maxFileSize is the maximum allowed file size (10 MB).
const maxFileSize = 10 * 1024 * 1024

func (s *service) UploadFile(ctx context.Context, entityType string, entityID uint, category string, fileHeader *multipart.FileHeader) (*File, error) {
	// 1. Validate file size
	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds the maximum limit of 10MB")
	}

	// 2. Validate MIME type
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		return nil, fmt.Errorf("file type '%s' is not allowed, accepted: jpg, png, webp, pdf", contentType)
	}

	// 3. Open the multipart file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// 4. Generate unique object name for MinIO
	objectName := storage.GenerateObjectName(entityType, entityID, category, fileHeader.Filename)

	// 5. Upload to MinIO
	publicURL, err := s.storageClient.Upload(ctx, objectName, file, fileHeader.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 6. Save metadata to database
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileRecord := &File{
		EntityType: entityType,
		EntityID:   entityID,
		Category:   category,
		FilePath:   objectName,
		FileName:   fileHeader.Filename,
		FileType:   ext,
		MimeType:   contentType,
		Size:       fileHeader.Size,
		PublicURL:  publicURL,
	}

	if err := s.repo.CreateFile(fileRecord); err != nil {
		// Best effort: try to clean up the uploaded file on DB failure
		_ = s.storageClient.Delete(ctx, objectName)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return fileRecord, nil
}

func (s *service) GetFilesByEntity(entityType string, entityID uint) ([]File, error) {
	return s.repo.FindFilesByEntity(entityType, entityID)
}

func (s *service) GetFilesByEntityAndCategory(entityType string, entityID uint, category string) ([]File, error) {
	return s.repo.FindFilesByEntityAndCategory(entityType, entityID, category)
}

// GetFileDownloadURL generates a temporary presigned URL (valid 15 minutes) for secure file access.
func (s *service) GetFileDownloadURL(ctx context.Context, fileID uint) (string, error) {
	fileRecord, err := s.repo.FindFileByID(fileID)
	if err != nil {
		return "", fmt.Errorf("file not found")
	}

	url, err := s.storageClient.GetPresignedURL(ctx, fileRecord.FilePath, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate download link: %w", err)
	}
	return url, nil
}

func (s *service) DeleteFile(ctx context.Context, fileID uint) error {
	fileRecord, err := s.repo.FindFileByID(fileID)
	if err != nil {
		return fmt.Errorf("file not found")
	}

	// Delete from MinIO storage first
	if err := s.storageClient.Delete(ctx, fileRecord.FilePath); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Then delete metadata from database
	return s.repo.DeleteFile(fileID)
}
