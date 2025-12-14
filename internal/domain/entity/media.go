package entity

import (
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/enum"
)

// CaseMedia represents a media file (image/video) attached to a case
type CaseMedia struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CaseID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"case_id"`
	MediaType    enum.MediaType `gorm:"type:varchar(20);not null" json:"media_type"`
	URL          string         `gorm:"type:varchar(500);not null" json:"url"`
	ThumbnailURL *string        `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	FileName     string         `gorm:"type:varchar(255)" json:"file_name,omitempty"`
	FileSize     int64          `json:"file_size,omitempty"`
	UploadedBy   *uuid.UUID     `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for CaseMedia
func (CaseMedia) TableName() string {
	return "case_media"
}

// IsImage returns true if the media is an image
func (m *CaseMedia) IsImage() bool {
	return m.MediaType == enum.MediaTypeImage
}

// IsVideo returns true if the media is a video
func (m *CaseMedia) IsVideo() bool {
	return m.MediaType == enum.MediaTypeVideo
}

// MediaUploadResult represents the result of a media upload
type MediaUploadResult struct {
	ID           uuid.UUID      `json:"id"`
	URL          string         `json:"url"`
	ThumbnailURL *string        `json:"thumbnail_url,omitempty"`
	MediaType    enum.MediaType `json:"media_type"`
	FileSize     int64          `json:"file_size"`
}

// AllowedImageTypes contains allowed image MIME types
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AllowedVideoTypes contains allowed video MIME types
var AllowedVideoTypes = map[string]bool{
	"video/mp4":       true,
	"video/quicktime": true,
	"video/x-msvideo": true,
	"video/webm":      true,
}

// MaxImageSize is the maximum allowed image size (10MB)
const MaxImageSize = 10 * 1024 * 1024

// MaxVideoSize is the maximum allowed video size (100MB)
const MaxVideoSize = 100 * 1024 * 1024

// IsAllowedImageType checks if the MIME type is an allowed image type
func IsAllowedImageType(mimeType string) bool {
	return AllowedImageTypes[mimeType]
}

// IsAllowedVideoType checks if the MIME type is an allowed video type
func IsAllowedVideoType(mimeType string) bool {
	return AllowedVideoTypes[mimeType]
}

// IsAllowedMediaType checks if the MIME type is allowed
func IsAllowedMediaType(mimeType string) bool {
	return IsAllowedImageType(mimeType) || IsAllowedVideoType(mimeType)
}

// GetMediaTypeFromMIME returns the MediaType based on MIME type
func GetMediaTypeFromMIME(mimeType string) enum.MediaType {
	if IsAllowedImageType(mimeType) {
		return enum.MediaTypeImage
	}
	if IsAllowedVideoType(mimeType) {
		return enum.MediaTypeVideo
	}
	return enum.MediaTypeImage // default
}

// GetMaxSizeForType returns the maximum file size for a media type
func GetMaxSizeForType(mediaType enum.MediaType) int64 {
	if mediaType == enum.MediaTypeVideo {
		return MaxVideoSize
	}
	return MaxImageSize
}
