package response

import (
	"time"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
)

// CaseResponse represents a case in response
type CaseResponse struct {
	ID              uuid.UUID                 `json:"id"`
	CaseType        enum.CaseType             `json:"caseType"`
	Status          enum.CaseStatus           `json:"status"`
	Urgency         enum.UrgencyLevel         `json:"urgency"`
	Location        GeoPointResponse          `json:"location"`
	Address         *string                   `json:"address,omitempty"`
	LocationNote    *string                   `json:"locationNote,omitempty"`
	Title           string                    `json:"title"`
	Description     *string                   `json:"description,omitempty"`
	ReporterID      *uuid.UUID                `json:"reporterId,omitempty"`
	ReporterName    *string                   `json:"reporterName,omitempty"`
	ReporterPhone   string                    `json:"reporterPhone"`
	IsAnonymous     bool                      `json:"isAnonymous"`
	VolunteerCount  int                       `json:"volunteerCount"`
	MaxVolunteers   int                       `json:"maxVolunteers"`
	CreatedAt       time.Time                 `json:"createdAt"`
	UpdatedAt       time.Time                 `json:"updatedAt"`
	AcceptedAt      *time.Time                `json:"acceptedAt,omitempty"`
	ResolvedAt      *time.Time                `json:"resolvedAt,omitempty"`
	AnimalDetails   *AnimalDetailsResponse    `json:"animalDetails,omitempty"`
	FloodDetails    *FloodDetailsResponse     `json:"floodDetails,omitempty"`
	AccidentDetails *AccidentDetailsResponse  `json:"accidentDetails,omitempty"`
	Media           []MediaResponse           `json:"media,omitempty"`
	Volunteers      []VolunteerResponse       `json:"volunteers,omitempty"`
}

// AnimalDetailsResponse represents animal details in response
type AnimalDetailsResponse struct {
	ID                   uuid.UUID            `json:"id"`
	AnimalType           enum.AnimalType      `json:"animalType"`
	AnimalTypeOther      *string              `json:"animalTypeOther,omitempty"`
	Condition            enum.AnimalCondition `json:"condition"`
	ConditionDescription *string              `json:"conditionDescription,omitempty"`
	EstimatedCount       int                  `json:"estimatedCount"`
}

// FloodDetailsResponse represents flood details in response
type FloodDetailsResponse struct {
	ID           uuid.UUID `json:"id"`
	PeopleCount  *int      `json:"peopleCount,omitempty"`
	HasChildren  bool      `json:"hasChildren"`
	HasElderly   bool      `json:"hasElderly"`
	HasDisabled  bool      `json:"hasDisabled"`
	WaterLevelCm *int      `json:"waterLevelCm,omitempty"`
	FloorLevel   *int      `json:"floorLevel,omitempty"`
	HasPower     *bool     `json:"hasPower,omitempty"`
	HasFoodWater *bool     `json:"hasFoodWater,omitempty"`
	MedicalNeeds *string   `json:"medicalNeeds,omitempty"`
}

// AccidentDetailsResponse represents accident details in response
type AccidentDetailsResponse struct {
	ID                uuid.UUID         `json:"id"`
	AccidentType      enum.AccidentType `json:"accidentType"`
	VictimCount       int               `json:"victimCount"`
	HasUnconscious    bool              `json:"hasUnconscious"`
	HasBleeding       bool              `json:"hasBleeding"`
	HasFracture       bool              `json:"hasFracture"`
	IsTrapped         bool              `json:"isTrapped"`
	HazardPresent     bool              `json:"hazardPresent"`
	HazardDescription *string           `json:"hazardDescription,omitempty"`
}

// MediaResponse represents media in response
type MediaResponse struct {
	ID           uuid.UUID      `json:"id"`
	MediaType    enum.MediaType `json:"mediaType"`
	URL          string         `json:"url"`
	ThumbnailURL *string        `json:"thumbnailUrl,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

// VolunteerResponse represents volunteer in response
type VolunteerResponse struct {
	ID              uuid.UUID            `json:"id"`
	VolunteerID     uuid.UUID            `json:"volunteerId"`
	VolunteerName   string               `json:"volunteerName"`
	VolunteerAvatar *string              `json:"volunteerAvatar,omitempty"`
	Status          enum.VolunteerStatus `json:"status"`
	DistanceKm      *float64             `json:"distanceKm,omitempty"`
	AcceptedAt      time.Time            `json:"acceptedAt"`
	ArrivedAt       *time.Time           `json:"arrivedAt,omitempty"`
	CompletedAt     *time.Time           `json:"completedAt,omitempty"`
	Note            *string              `json:"note,omitempty"`
}

// CaseNearbyResponse represents a nearby case in response
type CaseNearbyResponse struct {
	ID             uuid.UUID         `json:"id"`
	CaseType       enum.CaseType     `json:"caseType"`
	Title          string            `json:"title"`
	Urgency        enum.UrgencyLevel `json:"urgency"`
	Status         enum.CaseStatus   `json:"status"`
	DistanceKm     float64           `json:"distanceKm"`
	VolunteerCount int               `json:"volunteerCount"`
	CreatedAt      time.Time         `json:"createdAt"`
	Location       GeoPointResponse  `json:"location"`
}

// CaseUpdateResponse represents a case update in response
type CaseUpdateResponse struct {
	ID         uuid.UUID        `json:"id"`
	CaseID     uuid.UUID        `json:"caseId"`
	UpdateType enum.UpdateType  `json:"updateType"`
	UserID     *uuid.UUID       `json:"userId,omitempty"`
	UserName   *string          `json:"userName,omitempty"`
	Content    *string          `json:"content,omitempty"`
	OldStatus  *enum.CaseStatus `json:"oldStatus,omitempty"`
	NewStatus  *enum.CaseStatus `json:"newStatus,omitempty"`
	MediaURLs  []string         `json:"mediaUrls,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

// ToCaseResponse converts entity to response
func ToCaseResponse(c *entity.Case) *CaseResponse {
	if c == nil {
		return nil
	}

	resp := &CaseResponse{
		ID:       c.ID,
		CaseType: c.CaseType,
		Status:   c.Status,
		Urgency:  c.Urgency,
		Location: GeoPointResponse{
			Latitude:  c.Latitude,
			Longitude: c.Longitude,
		},
		Address:        c.Address,
		LocationNote:   c.LocationNote,
		Title:          c.Title,
		Description:    c.Description,
		ReporterID:     c.ReporterID,
		ReporterName:   c.ReporterName,
		ReporterPhone:  c.ReporterPhone,
		IsAnonymous:    c.IsAnonymous,
		VolunteerCount: c.VolunteerCount,
		MaxVolunteers:  c.MaxVolunteers,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		AcceptedAt:     c.AcceptedAt,
		ResolvedAt:     c.ResolvedAt,
	}

	// Convert animal details
	if c.AnimalDetails != nil {
		resp.AnimalDetails = &AnimalDetailsResponse{
			ID:                   c.AnimalDetails.ID,
			AnimalType:           c.AnimalDetails.AnimalType,
			AnimalTypeOther:      c.AnimalDetails.AnimalTypeOther,
			Condition:            c.AnimalDetails.Condition,
			ConditionDescription: c.AnimalDetails.ConditionDescription,
			EstimatedCount:       c.AnimalDetails.EstimatedCount,
		}
	}

	// Convert flood details
	if c.FloodDetails != nil {
		resp.FloodDetails = &FloodDetailsResponse{
			ID:           c.FloodDetails.ID,
			PeopleCount:  c.FloodDetails.PeopleCount,
			HasChildren:  c.FloodDetails.HasChildren,
			HasElderly:   c.FloodDetails.HasElderly,
			HasDisabled:  c.FloodDetails.HasDisabled,
			WaterLevelCm: c.FloodDetails.WaterLevelCm,
			FloorLevel:   c.FloodDetails.FloorLevel,
			HasPower:     c.FloodDetails.HasPower,
			HasFoodWater: c.FloodDetails.HasFoodWater,
			MedicalNeeds: c.FloodDetails.MedicalNeeds,
		}
	}

	// Convert accident details
	if c.AccidentDetails != nil {
		resp.AccidentDetails = &AccidentDetailsResponse{
			ID:                c.AccidentDetails.ID,
			AccidentType:      c.AccidentDetails.AccidentType,
			VictimCount:       c.AccidentDetails.VictimCount,
			HasUnconscious:    c.AccidentDetails.HasUnconscious,
			HasBleeding:       c.AccidentDetails.HasBleeding,
			HasFracture:       c.AccidentDetails.HasFracture,
			IsTrapped:         c.AccidentDetails.IsTrapped,
			HazardPresent:     c.AccidentDetails.HazardPresent,
			HazardDescription: c.AccidentDetails.HazardDescription,
		}
	}

	// Convert media
	if len(c.Media) > 0 {
		resp.Media = make([]MediaResponse, len(c.Media))
		for i, m := range c.Media {
			resp.Media[i] = MediaResponse{
				ID:           m.ID,
				MediaType:    m.MediaType,
				URL:          m.URL,
				ThumbnailURL: m.ThumbnailURL,
				CreatedAt:    m.CreatedAt,
			}
		}
	}

	// Convert volunteers
	if len(c.Volunteers) > 0 {
		resp.Volunteers = make([]VolunteerResponse, len(c.Volunteers))
		for i, v := range c.Volunteers {
			vr := VolunteerResponse{
				ID:          v.ID,
				VolunteerID: v.VolunteerID,
				Status:      v.Status,
				DistanceKm:  v.DistanceKm,
				AcceptedAt:  v.AcceptedAt,
				ArrivedAt:   v.ArrivedAt,
				CompletedAt: v.CompletedAt,
				Note:        v.Note,
			}
			if v.Volunteer != nil {
				vr.VolunteerName = v.Volunteer.DisplayName
				vr.VolunteerAvatar = v.Volunteer.AvatarURL
			}
			resp.Volunteers[i] = vr
		}
	}

	return resp
}

// ToCaseNearbyResponse converts entity to response
func ToCaseNearbyResponse(c *entity.CaseNearby) *CaseNearbyResponse {
	if c == nil {
		return nil
	}

	return &CaseNearbyResponse{
		ID:             c.ID,
		CaseType:       c.CaseType,
		Title:          c.Title,
		Urgency:        c.Urgency,
		Status:         c.Status,
		DistanceKm:     c.DistanceKm,
		VolunteerCount: c.VolunteerCount,
		CreatedAt:      c.CreatedAt,
		Location: GeoPointResponse{
			Latitude:  c.Latitude,
			Longitude: c.Longitude,
		},
	}
}

// ToCaseUpdateResponse converts entity to response
func ToCaseUpdateResponse(u *entity.CaseUpdate) *CaseUpdateResponse {
	if u == nil {
		return nil
	}

	resp := &CaseUpdateResponse{
		ID:         u.ID,
		CaseID:     u.CaseID,
		UpdateType: u.UpdateType,
		UserID:     u.UserID,
		Content:    u.Content,
		OldStatus:  u.OldStatus,
		NewStatus:  u.NewStatus,
		CreatedAt:  u.CreatedAt,
	}

	if u.User != nil {
		resp.UserName = &u.User.DisplayName
	}

	if len(u.MediaURLs) > 0 {
		resp.MediaURLs = u.MediaURLs
	}

	return resp
}

// ToCaseListResponse converts a slice of cases to response
func ToCaseListResponse(cases []entity.Case) []CaseResponse {
	result := make([]CaseResponse, len(cases))
	for i, c := range cases {
		result[i] = *ToCaseResponse(&c)
	}
	return result
}

// ToCaseNearbyListResponse converts a slice of nearby cases to response
func ToCaseNearbyListResponse(cases []entity.CaseNearby) []CaseNearbyResponse {
	result := make([]CaseNearbyResponse, len(cases))
	for i, c := range cases {
		result[i] = *ToCaseNearbyResponse(&c)
	}
	return result
}

// ToVolunteerResponse converts a single volunteer entity to response
func ToVolunteerResponse(v *entity.CaseVolunteer) *VolunteerResponse {
	if v == nil {
		return nil
	}

	vr := &VolunteerResponse{
		ID:          v.ID,
		VolunteerID: v.VolunteerID,
		Status:      v.Status,
		DistanceKm:  v.DistanceKm,
		AcceptedAt:  v.AcceptedAt,
		ArrivedAt:   v.ArrivedAt,
		CompletedAt: v.CompletedAt,
		Note:        v.Note,
	}

	if v.Volunteer != nil {
		vr.VolunteerName = v.Volunteer.DisplayName
		vr.VolunteerAvatar = v.Volunteer.AvatarURL
	}

	return vr
}

// ToVolunteerListResponse converts a slice of volunteers to response
func ToVolunteerListResponse(volunteers []entity.CaseVolunteer) []VolunteerResponse {
	result := make([]VolunteerResponse, len(volunteers))
	for i, v := range volunteers {
		result[i] = *ToVolunteerResponse(&v)
	}
	return result
}

// CommentResponse represents a comment in response
type CommentResponse struct {
	ID        uuid.UUID      `json:"id"`
	CaseID    uuid.UUID      `json:"caseId"`
	Author    CommentAuthor  `json:"author"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
}

// CommentAuthor represents the author of a comment
type CommentAuthor struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Role      string    `json:"role"`
}

// ToCommentResponse converts entity to response
func ToCommentResponse(c *entity.CaseComment, caseReporterID *uuid.UUID) *CommentResponse {
	if c == nil {
		return nil
	}

	resp := &CommentResponse{
		ID:        c.ID,
		CaseID:    c.CaseID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}

	if c.User != nil {
		// Determine role based on user role and case reporter
		role := string(c.User.Role)
		if caseReporterID != nil && c.UserID == *caseReporterID {
			role = "reporter"
		} else if c.User.Role == enum.UserRoleVolunteer || c.User.Role == enum.UserRoleBoth {
			role = "volunteer"
		}

		resp.Author = CommentAuthor{
			ID:        c.User.ID,
			Name:      c.User.DisplayName,
			AvatarURL: c.User.AvatarURL,
			Role:      role,
		}
	}

	return resp
}

// ToCommentListResponse converts a slice of comments to response
func ToCommentListResponse(comments []entity.CaseComment, caseReporterID *uuid.UUID) []CommentResponse {
	result := make([]CommentResponse, len(comments))
	for i, c := range comments {
		result[i] = *ToCommentResponse(&c, caseReporterID)
	}
	return result
}
