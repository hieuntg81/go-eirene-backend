package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"gorm.io/gorm"
)

// CaseRepository defines the interface for case data access
type CaseRepository interface {
	Create(ctx context.Context, c *entity.Case) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Case, error)
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Case, error)
	GetNearby(ctx context.Context, lat, lng float64, radiusKm int, types []enum.CaseType, limit int) ([]entity.CaseNearby, error)
	GetCases(ctx context.Context, query string, caseType *enum.CaseType, status *enum.CaseStatus, urgency *enum.UrgencyLevel, limit, offset int) ([]entity.Case, int64, error)
	Update(ctx context.Context, c *entity.Case) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.CaseStatus) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Volunteers
	AddVolunteer(ctx context.Context, cv *entity.CaseVolunteer) error
	GetVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID) (*entity.CaseVolunteer, error)
	UpdateVolunteerStatus(ctx context.Context, caseID, volunteerID uuid.UUID, status enum.VolunteerStatus) error
	ReactivateVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID, latitude, longitude *float64, distanceKm *float64) error
	RemoveVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID) error
	GetVolunteersByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseVolunteer, error)

	// Updates/Timeline
	CreateUpdate(ctx context.Context, update *entity.CaseUpdate) error
	GetUpdates(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]entity.CaseUpdate, int64, error)

	// Comments
	CreateComment(ctx context.Context, comment *entity.CaseComment) error
	GetCommentsByCaseID(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]entity.CaseComment, int64, error)
	DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error

	// User cases
	GetUserReportedCases(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Case, int64, error)
	GetUserAcceptedCases(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Case, int64, error)
}

type caseRepository struct {
	db *gorm.DB
}

// NewCaseRepository creates a new CaseRepository
func NewCaseRepository(db interface{}) CaseRepository {
	return &caseRepository{db: db.(*gorm.DB)}
}

func (r *caseRepository) Create(ctx context.Context, c *entity.Case) error {
	// Generate UUID if not set
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the main case (omit associations to handle them manually)
		if err := tx.Omit("AnimalDetails", "FloodDetails", "AccidentDetails", "Media", "Volunteers", "Updates").Create(c).Error; err != nil {
			return err
		}

		// Create type-specific details
		switch c.CaseType {
		case enum.CaseTypeAnimal:
			if c.AnimalDetails != nil {
				if c.AnimalDetails.ID == uuid.Nil {
					c.AnimalDetails.ID = uuid.New()
				}
				c.AnimalDetails.CaseID = c.ID
				if err := tx.Create(c.AnimalDetails).Error; err != nil {
					return err
				}
			}
		case enum.CaseTypeFlood:
			if c.FloodDetails != nil {
				if c.FloodDetails.ID == uuid.Nil {
					c.FloodDetails.ID = uuid.New()
				}
				c.FloodDetails.CaseID = c.ID
				if err := tx.Create(c.FloodDetails).Error; err != nil {
					return err
				}
			}
		case enum.CaseTypeAccident:
			if c.AccidentDetails != nil {
				if c.AccidentDetails.ID == uuid.Nil {
					c.AccidentDetails.ID = uuid.New()
				}
				c.AccidentDetails.CaseID = c.ID
				if err := tx.Create(c.AccidentDetails).Error; err != nil {
					return err
				}
			}
		}

		// Create initial update
		update := &entity.CaseUpdate{
			ID:         uuid.New(),
			CaseID:     c.ID,
			UpdateType: enum.UpdateTypeSystem,
			Content:    stringPtr("Case created"),
			NewStatus:  &c.Status,
		}
		return tx.Create(update).Error
	})
}

func (r *caseRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Case, error) {
	var c entity.Case
	err := r.db.WithContext(ctx).
		First(&c, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *caseRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Case, error) {
	var c entity.Case
	err := r.db.WithContext(ctx).
		Preload("AnimalDetails").
		Preload("FloodDetails").
		Preload("AccidentDetails").
		Preload("Media").
		Preload("Volunteers").
		Preload("Volunteers.Volunteer").
		First(&c, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *caseRepository) GetNearby(ctx context.Context, lat, lng float64, radiusKm int, types []enum.CaseType, limit int) ([]entity.CaseNearby, error) {
	// Create bounding box for initial filtering (performance optimization)
	bbox := entity.NewBoundingBox(lat, lng, float64(radiusKm))

	// Build query with bounding box filter
	query := r.db.WithContext(ctx).
		Model(&entity.Case{}).
		Select("id, case_type, title, urgency, status, volunteer_count, created_at, latitude, longitude").
		Where("status IN ?", []string{"pending", "accepted", "in_progress"}).
		Where("latitude BETWEEN ? AND ?", bbox.MinLat, bbox.MaxLat).
		Where("longitude BETWEEN ? AND ?", bbox.MinLng, bbox.MaxLng)

	if len(types) > 0 {
		typeStrings := make([]string, len(types))
		for i, t := range types {
			typeStrings[i] = string(t)
		}
		query = query.Where("case_type IN ?", typeStrings)
	}

	var cases []entity.Case
	if err := query.Find(&cases).Error; err != nil {
		return nil, err
	}

	// Calculate exact distances using Haversine and filter
	centerPoint := entity.NewGeoPoint(lat, lng)
	var nearby []entity.CaseNearby

	for _, c := range cases {
		casePoint := entity.NewGeoPoint(c.Latitude, c.Longitude)
		distance := centerPoint.DistanceKm(casePoint)

		// Filter by exact radius
		if distance <= float64(radiusKm) {
			nearby = append(nearby, entity.CaseNearby{
				ID:             c.ID,
				CaseType:       c.CaseType,
				Title:          c.Title,
				Urgency:        c.Urgency,
				Status:         c.Status,
				DistanceKm:     distance,
				VolunteerCount: c.VolunteerCount,
				CreatedAt:      c.CreatedAt,
				Latitude:       c.Latitude,
				Longitude:      c.Longitude,
			})
		}
	}

	// Sort by urgency then distance
	sort.Slice(nearby, func(i, j int) bool {
		// Priority order: critical=1, high=2, medium=3, low=4
		urgencyOrder := map[enum.UrgencyLevel]int{
			enum.UrgencyCritical: 1,
			enum.UrgencyHigh:     2,
			enum.UrgencyMedium:   3,
			enum.UrgencyLow:      4,
		}
		if urgencyOrder[nearby[i].Urgency] != urgencyOrder[nearby[j].Urgency] {
			return urgencyOrder[nearby[i].Urgency] < urgencyOrder[nearby[j].Urgency]
		}
		return nearby[i].DistanceKm < nearby[j].DistanceKm
	})

	// Apply limit
	if limit > 0 && len(nearby) > limit {
		nearby = nearby[:limit]
	}

	return nearby, nil
}

func (r *caseRepository) GetCases(ctx context.Context, query string, caseType *enum.CaseType, status *enum.CaseStatus, urgency *enum.UrgencyLevel, limit, offset int) ([]entity.Case, int64, error) {
	var cases []entity.Case
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Case{})

	// Apply search query - search in title and address
	if query != "" {
		searchPattern := "%" + query + "%"
		db = db.Where("title ILIKE ? OR address ILIKE ?", searchPattern, searchPattern)
	}

	// Apply type filter
	if caseType != nil {
		db = db.Where("case_type = ?", *caseType)
	}

	// Apply status filter
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	// Apply urgency filter
	if urgency != nil {
		db = db.Where("urgency = ?", *urgency)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&cases).Error; err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

func (r *caseRepository) Update(ctx context.Context, c *entity.Case) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *caseRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.CaseStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == enum.CaseStatusAccepted {
		updates["accepted_at"] = time.Now()
	} else if status == enum.CaseStatusResolved {
		updates["resolved_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&entity.Case{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *caseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entity.Case{}).
		Where("id = ?", id).
		Update("status", enum.CaseStatusCancelled).Error
}

func (r *caseRepository) AddVolunteer(ctx context.Context, cv *entity.CaseVolunteer) error {
	// Generate UUID if not set
	if cv.ID == uuid.Nil {
		cv.ID = uuid.New()
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Add volunteer
		if err := tx.Create(cv).Error; err != nil {
			return err
		}

		// Increment volunteer count
		if err := tx.Model(&entity.Case{}).
			Where("id = ?", cv.CaseID).
			UpdateColumn("volunteer_count", gorm.Expr("volunteer_count + 1")).Error; err != nil {
			return err
		}

		// Update case status if this is the first volunteer
		return tx.Model(&entity.Case{}).
			Where("id = ? AND status = ?", cv.CaseID, enum.CaseStatusPending).
			Updates(map[string]interface{}{
				"status":      enum.CaseStatusAccepted,
				"accepted_at": time.Now(),
			}).Error
	})
}

func (r *caseRepository) GetVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID) (*entity.CaseVolunteer, error) {
	var cv entity.CaseVolunteer
	err := r.db.WithContext(ctx).
		Preload("Volunteer").
		First(&cv, "case_id = ? AND volunteer_id = ?", caseID, volunteerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cv, nil
}

func (r *caseRepository) UpdateVolunteerStatus(ctx context.Context, caseID, volunteerID uuid.UUID, status enum.VolunteerStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case enum.VolunteerStatusOnSite:
		updates["arrived_at"] = time.Now()
	case enum.VolunteerStatusCompleted:
		updates["completed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&entity.CaseVolunteer{}).
		Where("case_id = ? AND volunteer_id = ?", caseID, volunteerID).
		Updates(updates).Error
}

func (r *caseRepository) ReactivateVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID, latitude, longitude *float64, distanceKm *float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update volunteer status back to accepted and update location info
		updates := map[string]interface{}{
			"status":             enum.VolunteerStatusAccepted,
			"accepted_at":        time.Now(),
			"accepted_latitude":  latitude,
			"accepted_longitude": longitude,
			"distance_km":        distanceKm,
			"arrived_at":         nil,
			"completed_at":       nil,
		}

		if err := tx.Model(&entity.CaseVolunteer{}).
			Where("case_id = ? AND volunteer_id = ?", caseID, volunteerID).
			Updates(updates).Error; err != nil {
			return err
		}

		// Increment volunteer count
		return tx.Model(&entity.Case{}).
			Where("id = ?", caseID).
			UpdateColumn("volunteer_count", gorm.Expr("volunteer_count + 1")).Error
	})
}

func (r *caseRepository) RemoveVolunteer(ctx context.Context, caseID, volunteerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update volunteer status to withdrawn
		if err := tx.Model(&entity.CaseVolunteer{}).
			Where("case_id = ? AND volunteer_id = ?", caseID, volunteerID).
			Update("status", enum.VolunteerStatusWithdrawn).Error; err != nil {
			return err
		}

		// Decrement volunteer count
		return tx.Model(&entity.Case{}).
			Where("id = ?", caseID).
			UpdateColumn("volunteer_count", gorm.Expr("GREATEST(0, volunteer_count - 1)")).Error
	})
}

func (r *caseRepository) GetVolunteersByCaseID(ctx context.Context, caseID uuid.UUID) ([]entity.CaseVolunteer, error) {
	var volunteers []entity.CaseVolunteer
	err := r.db.WithContext(ctx).
		Preload("Volunteer").
		Where("case_id = ?", caseID).
		Order("accepted_at ASC").
		Find(&volunteers).Error
	return volunteers, err
}

func (r *caseRepository) CreateUpdate(ctx context.Context, update *entity.CaseUpdate) error {
	if update.ID == uuid.Nil {
		update.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(update).Error
}

func (r *caseRepository) GetUpdates(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]entity.CaseUpdate, int64, error) {
	var updates []entity.CaseUpdate
	var total int64

	err := r.db.WithContext(ctx).
		Model(&entity.CaseUpdate{}).
		Where("case_id = ?", caseID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Preload("User").
		Where("case_id = ?", caseID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&updates).Error

	return updates, total, err
}

func (r *caseRepository) GetUserReportedCases(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Case, int64, error) {
	var cases []entity.Case
	var total int64

	err := r.db.WithContext(ctx).
		Model(&entity.Case{}).
		Where("reporter_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Where("reporter_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&cases).Error

	return cases, total, err
}

func (r *caseRepository) GetUserAcceptedCases(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Case, int64, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&entity.CaseVolunteer{}).
		Where("volunteer_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var caseIDs []uuid.UUID
	err = r.db.WithContext(ctx).
		Model(&entity.CaseVolunteer{}).
		Select("case_id").
		Where("volunteer_id = ?", userID).
		Order("accepted_at DESC").
		Limit(limit).
		Offset(offset).
		Pluck("case_id", &caseIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(caseIDs) == 0 {
		return []entity.Case{}, 0, nil
	}

	var cases []entity.Case
	err = r.db.WithContext(ctx).
		Where("id IN ?", caseIDs).
		Find(&cases).Error

	return cases, total, err
}

func (r *caseRepository) CreateComment(ctx context.Context, comment *entity.CaseComment) error {
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *caseRepository) GetCommentsByCaseID(ctx context.Context, caseID uuid.UUID, limit, offset int) ([]entity.CaseComment, int64, error) {
	var comments []entity.CaseComment
	var total int64

	err := r.db.WithContext(ctx).
		Model(&entity.CaseComment{}).
		Where("case_id = ?", caseID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Preload("User").
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error

	return comments, total, err
}

func (r *caseRepository) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", commentID, userID).
		Delete(&entity.CaseComment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("comment not found or not authorized")
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}
