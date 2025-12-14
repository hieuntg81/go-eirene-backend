package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"gorm.io/gorm"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	Create(ctx context.Context, media *entity.CaseMedia) error
	CreateBatch(ctx context.Context, media []entity.CaseMedia) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CaseMedia, error)
	GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseMedia, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCaseID(ctx context.Context, caseID uuid.UUID) error
}

type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new MediaRepository
func NewMediaRepository(db interface{}) MediaRepository {
	return &mediaRepository{db: db.(*gorm.DB)}
}

func (r *mediaRepository) Create(ctx context.Context, media *entity.CaseMedia) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *mediaRepository) CreateBatch(ctx context.Context, media []entity.CaseMedia) error {
	if len(media) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&media).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CaseMedia, error) {
	var media entity.CaseMedia
	err := r.db.WithContext(ctx).
		First(&media, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseMedia, error) {
	var media []entity.CaseMedia
	err := r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&media).Error
	return media, err
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.CaseMedia{}, "id = ?", id).Error
}

func (r *mediaRepository) DeleteByCaseID(ctx context.Context, caseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.CaseMedia{}, "case_id = ?", caseID).Error
}
