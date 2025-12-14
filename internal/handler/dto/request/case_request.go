package request

import "bamboo-rescue/internal/domain/enum"

// CreateCaseRequest represents case creation request
type CreateCaseRequest struct {
	CaseType      enum.CaseType     `json:"case_type" validate:"required,oneof=animal flood accident"`
	Urgency       enum.UrgencyLevel `json:"urgency" validate:"required,oneof=low medium high critical"`
	Latitude      float64           `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude     float64           `json:"longitude" validate:"required,min=-180,max=180"`
	Address       *string           `json:"address"`
	LocationNote  *string           `json:"location_note" validate:"omitempty,max=500"`
	Title         string            `json:"title" validate:"required,min=5,max=200"`
	Description   *string           `json:"description"`
	ReporterName  *string           `json:"reporter_name" validate:"omitempty,max=100"`
	ReporterPhone string            `json:"reporter_phone" validate:"required,min=10,max=20"`
	IsAnonymous   bool              `json:"is_anonymous"`

	// Animal details
	AnimalType           *enum.AnimalType      `json:"animal_type" validate:"omitempty,oneof=dog cat bird other"`
	AnimalTypeOther      *string               `json:"animal_type_other" validate:"omitempty,max=100"`
	AnimalCondition      *enum.AnimalCondition `json:"animal_condition" validate:"omitempty,oneof=injured trapped sick abandoned other"`
	ConditionDescription *string               `json:"condition_description"`
	EstimatedCount       *int                  `json:"estimated_count" validate:"omitempty,min=1"`

	// Flood details
	PeopleCount  *int    `json:"people_count" validate:"omitempty,min=1"`
	HasChildren  *bool   `json:"has_children"`
	HasElderly   *bool   `json:"has_elderly"`
	HasDisabled  *bool   `json:"has_disabled"`
	WaterLevelCm *int    `json:"water_level_cm" validate:"omitempty,min=0"`
	FloorLevel   *int    `json:"floor_level" validate:"omitempty,min=0"`
	HasPower     *bool   `json:"has_power"`
	HasFoodWater *bool   `json:"has_food_water"`
	MedicalNeeds *string `json:"medical_needs"`

	// Accident details
	AccidentType      *enum.AccidentType `json:"accident_type" validate:"omitempty,oneof=traffic fall fire drowning electric other"`
	VictimCount       *int               `json:"victim_count" validate:"omitempty,min=1"`
	HasUnconscious    *bool              `json:"has_unconscious"`
	HasBleeding       *bool              `json:"has_bleeding"`
	HasFracture       *bool              `json:"has_fracture"`
	IsTrapped         *bool              `json:"is_trapped"`
	HazardPresent     *bool              `json:"hazard_present"`
	HazardDescription *string            `json:"hazard_description"`

	// Media
	MediaURLs []string `json:"media_urls"`
}

// GetNearbyCasesRequest represents nearby cases query request
type GetNearbyCasesRequest struct {
	Latitude  float64           `form:"lat" validate:"required,min=-90,max=90"`
	Longitude float64           `form:"lng" validate:"required,min=-180,max=180"`
	RadiusKm  int               `form:"radius" validate:"omitempty,min=1,max=100"`
	Types     []enum.CaseType   `form:"types"`
	Limit     int               `form:"limit" validate:"omitempty,min=1,max=100"`
}

// GetCasesRequest represents cases list query with search and pagination
type GetCasesRequest struct {
	Query   string          `form:"q"`
	Type    *enum.CaseType  `form:"type"`
	Status  *enum.CaseStatus `form:"status"`
	Urgency *enum.UrgencyLevel `form:"urgency"`
	Page    int             `form:"page" validate:"omitempty,min=1"`
	Limit   int             `form:"limit" validate:"omitempty,min=1,max=100"`
}

// GetDefaultPage returns the page number or default
func (r *GetCasesRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultLimit returns the limit or default
func (r *GetCasesRequest) GetDefaultLimit() int {
	if r.Limit <= 0 {
		return 20
	}
	return r.Limit
}

// GetOffset calculates the offset for pagination
func (r *GetCasesRequest) GetOffset() int {
	return (r.GetDefaultPage() - 1) * r.GetDefaultLimit()
}

// UpdateCaseRequest represents case update request
type UpdateCaseRequest struct {
	Title        *string            `json:"title" validate:"omitempty,min=5,max=200"`
	Description  *string            `json:"description"`
	Urgency      *enum.UrgencyLevel `json:"urgency" validate:"omitempty,oneof=low medium high critical"`
	Status       *enum.CaseStatus   `json:"status" validate:"omitempty,oneof=pending accepted in_progress resolved cancelled"`
	Address      *string            `json:"address"`
	LocationNote *string            `json:"location_note" validate:"omitempty,max=500"`
}

// AcceptCaseRequest represents case acceptance request
type AcceptCaseRequest struct {
	Latitude  *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
}

// UpdateVolunteerStatusRequest represents volunteer status update request
type UpdateVolunteerStatusRequest struct {
	Status enum.VolunteerStatus `json:"status" validate:"required,oneof=accepted en_route on_site handling completed withdrawn"`
	Note   *string              `json:"note"`
}

// CreateCaseUpdateRequest represents case update/timeline entry request
type CreateCaseUpdateRequest struct {
	Content   string   `json:"content" validate:"required,min=1"`
	MediaURLs []string `json:"media_urls"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `form:"page" validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

// GetDefaultPage returns the page number or default
func (p *PaginationRequest) GetDefaultPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// GetDefaultLimit returns the limit or default
func (p *PaginationRequest) GetDefaultLimit() int {
	if p.Limit <= 0 {
		return 20
	}
	return p.Limit
}

// GetOffset calculates the offset for pagination
func (p *PaginationRequest) GetOffset() int {
	return (p.GetDefaultPage() - 1) * p.GetDefaultLimit()
}

// GetLimit returns the limit for pagination
func (p *PaginationRequest) GetLimit() int {
	return p.GetDefaultLimit()
}

// CreateCommentRequest represents comment creation request
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}
