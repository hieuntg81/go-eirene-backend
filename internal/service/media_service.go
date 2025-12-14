package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/repository"
	"bamboo-rescue/pkg/storage"
	"go.uber.org/zap"
)

// MediaService defines the interface for media operations
type MediaService interface {
	Upload(ctx context.Context, file *multipart.FileHeader, caseID uuid.UUID) (*entity.MediaUploadResult, error)
	UploadMultiple(ctx context.Context, files []*multipart.FileHeader, caseID uuid.UUID) ([]entity.MediaUploadResult, error)
	Delete(ctx context.Context, mediaID uuid.UUID) error
	GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseMedia, error)
}

type mediaService struct {
	mediaRepo     repository.MediaRepository
	storageClient storage.Client
	log           *zap.Logger
}

// NewMediaService creates a new MediaService
func NewMediaService(mediaRepo repository.MediaRepository, storageClient storage.Client, log *zap.Logger) MediaService {
	return &mediaService{
		mediaRepo:     mediaRepo,
		storageClient: storageClient,
		log:           log,
	}
}

var allowedImageTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

var allowedVideoTypes = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".avi":  true,
	".webm": true,
}

const maxFileSize = 50 * 1024 * 1024 // 50MB

func (s *mediaService) Upload(ctx context.Context, file *multipart.FileHeader, caseID uuid.UUID) (*entity.MediaUploadResult, error) {
	// Validate file size
	if file.Size > maxFileSize {
		return nil, middleware.NewAppError("FILE_TOO_LARGE", "File size exceeds 50MB limit", 400)
	}

	// Determine media type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	var mediaType enum.MediaType
	if allowedImageTypes[ext] {
		mediaType = enum.MediaTypeImage
	} else if allowedVideoTypes[ext] {
		mediaType = enum.MediaTypeVideo
	} else {
		return nil, middleware.NewAppError("INVALID_FILE_TYPE", "File type not allowed", 400)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		s.log.Error("Failed to open uploaded file", zap.Error(err))
		return nil, err
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		s.log.Error("Failed to read file content", zap.Error(err))
		return nil, err
	}

	// Generate unique filename
	mediaID := uuid.New()
	filename := fmt.Sprintf("cases/%s/%s%s", caseID.String(), mediaID.String(), ext)

	// Upload to storage
	fileURL, err := s.storageClient.Upload(ctx, filename, content, file.Header.Get("Content-Type"))
	if err != nil {
		s.log.Error("Failed to upload file to storage", zap.Error(err))
		return nil, err
	}

	// Generate thumbnail for images
	var thumbnailURL *string
	if mediaType == enum.MediaTypeImage {
		thumbFilename := fmt.Sprintf("cases/%s/%s_thumb%s", caseID.String(), mediaID.String(), ext)
		// TODO: Implement actual thumbnail generation
		// For now, use the same URL
		thumbURL := fileURL
		thumbnailURL = &thumbURL
		_ = thumbFilename
	}

	// Save to database
	media := &entity.CaseMedia{
		ID:           mediaID,
		CaseID:       caseID,
		MediaType:    mediaType,
		URL:          fileURL,
		ThumbnailURL: thumbnailURL,
		FileName:     file.Filename,
		FileSize:     file.Size,
		CreatedAt:    time.Now(),
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		s.log.Error("Failed to save media record", zap.Error(err))
		// Try to delete uploaded file
		_ = s.storageClient.Delete(ctx, filename)
		return nil, err
	}

	return &entity.MediaUploadResult{
		ID:           mediaID,
		URL:          fileURL,
		ThumbnailURL: thumbnailURL,
		MediaType:    mediaType,
		FileSize:     file.Size,
	}, nil
}

func (s *mediaService) UploadMultiple(ctx context.Context, files []*multipart.FileHeader, caseID uuid.UUID) ([]entity.MediaUploadResult, error) {
	results := make([]entity.MediaUploadResult, 0, len(files))

	for _, file := range files {
		result, err := s.Upload(ctx, file, caseID)
		if err != nil {
			s.log.Warn("Failed to upload file", zap.String("filename", file.Filename), zap.Error(err))
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

func (s *mediaService) Delete(ctx context.Context, mediaID uuid.UUID) error {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return err
	}
	if media == nil {
		return middleware.NewAppError("MEDIA_NOT_FOUND", "Media not found", 404)
	}

	// Delete from storage
	if err := s.storageClient.Delete(ctx, media.URL); err != nil {
		s.log.Warn("Failed to delete file from storage", zap.Error(err))
	}

	// Delete thumbnail if exists
	if media.ThumbnailURL != nil {
		if err := s.storageClient.Delete(ctx, *media.ThumbnailURL); err != nil {
			s.log.Warn("Failed to delete thumbnail from storage", zap.Error(err))
		}
	}

	// Delete from database
	if err := s.mediaRepo.Delete(ctx, mediaID); err != nil {
		s.log.Error("Failed to delete media record", zap.Error(err))
		return err
	}

	return nil
}

func (s *mediaService) GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseMedia, error) {
	return s.mediaRepo.GetByCaseID(ctx, caseID)
}
