package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"bamboo-rescue/internal/domain/enum"
)

// Case represents a rescue case
type Case struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	CaseType       enum.CaseType     `gorm:"type:varchar(20);not null" json:"case_type"`
	Status         enum.CaseStatus   `gorm:"type:varchar(30);not null;default:'pending'" json:"status"`
	Urgency        enum.UrgencyLevel `gorm:"type:varchar(20);not null;default:'medium'" json:"urgency"`
	Latitude       float64           `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude      float64           `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Address        *string           `gorm:"type:text" json:"address,omitempty"`
	LocationNote   *string           `gorm:"type:varchar(500)" json:"location_note,omitempty"`
	Title          string            `gorm:"type:varchar(200);not null" json:"title"`
	Description    *string           `gorm:"type:text" json:"description,omitempty"`
	ReporterID     *uuid.UUID        `gorm:"type:uuid" json:"reporter_id,omitempty"`
	ReporterName   *string           `gorm:"type:varchar(100)" json:"reporter_name,omitempty"`
	ReporterPhone  string            `gorm:"type:varchar(20);not null" json:"reporter_phone"`
	IsAnonymous    bool              `gorm:"default:false" json:"is_anonymous"`
	VolunteerCount int               `gorm:"default:0" json:"volunteer_count"`
	MaxVolunteers  int               `gorm:"default:5" json:"max_volunteers"`
	CreatedAt      time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	AcceptedAt     *time.Time        `json:"accepted_at,omitempty"`
	ResolvedAt     *time.Time        `json:"resolved_at,omitempty"`
	ExpiresAt      *time.Time        `gorm:"-" json:"expires_at,omitempty"`

	// Relations
	Reporter        *User                `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	AnimalDetails   *CaseAnimalDetails   `gorm:"foreignKey:CaseID" json:"animal_details,omitempty"`
	FloodDetails    *CaseFloodDetails    `gorm:"foreignKey:CaseID" json:"flood_details,omitempty"`
	AccidentDetails *CaseAccidentDetails `gorm:"foreignKey:CaseID" json:"accident_details,omitempty"`
	Media           []CaseMedia          `gorm:"foreignKey:CaseID" json:"media,omitempty"`
	Volunteers      []CaseVolunteer      `gorm:"foreignKey:CaseID" json:"volunteers,omitempty"`
	Updates         []CaseUpdate         `gorm:"foreignKey:CaseID" json:"updates,omitempty"`
}

// TableName returns the table name for Case
func (Case) TableName() string {
	return "cases"
}

// GetLocation returns a GeoPoint from the case's coordinates
func (c *Case) GetLocation() *GeoPoint {
	return NewGeoPoint(c.Latitude, c.Longitude)
}

// SetLocation sets the case's coordinates from a GeoPoint
func (c *Case) SetLocation(loc *GeoPoint) {
	if loc == nil {
		c.Latitude = 0
		c.Longitude = 0
		return
	}
	c.Latitude = loc.Latitude
	c.Longitude = loc.Longitude
}

// CaseAnimalDetails contains specific details for animal rescue cases
type CaseAnimalDetails struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID               uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex" json:"case_id"`
	AnimalType           enum.AnimalType      `gorm:"type:varchar(20);not null" json:"animal_type"`
	AnimalTypeOther      *string              `gorm:"type:varchar(100)" json:"animal_type_other,omitempty"`
	Condition            enum.AnimalCondition `gorm:"type:varchar(20);not null" json:"condition"`
	ConditionDescription *string              `gorm:"type:text" json:"condition_description,omitempty"`
	EstimatedCount       int                  `gorm:"default:1" json:"estimated_count"`
	CreatedAt            time.Time            `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for CaseAnimalDetails
func (CaseAnimalDetails) TableName() string {
	return "case_animal_details"
}

// CaseFloodDetails contains specific details for flood rescue cases
type CaseFloodDetails struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"case_id"`
	PeopleCount  *int      `json:"people_count,omitempty"`
	HasChildren  bool      `gorm:"default:false" json:"has_children"`
	HasElderly   bool      `gorm:"default:false" json:"has_elderly"`
	HasDisabled  bool      `gorm:"default:false" json:"has_disabled"`
	WaterLevelCm *int      `json:"water_level_cm,omitempty"`
	FloorLevel   *int      `json:"floor_level,omitempty"`
	HasPower     *bool     `json:"has_power,omitempty"`
	HasFoodWater *bool     `json:"has_food_water,omitempty"`
	MedicalNeeds *string   `gorm:"type:text" json:"medical_needs,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for CaseFloodDetails
func (CaseFloodDetails) TableName() string {
	return "case_flood_details"
}

// CaseAccidentDetails contains specific details for accident rescue cases
type CaseAccidentDetails struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID            uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex" json:"case_id"`
	AccidentType      enum.AccidentType `gorm:"type:varchar(20);not null" json:"accident_type"`
	VictimCount       int               `gorm:"default:1" json:"victim_count"`
	HasUnconscious    bool              `gorm:"default:false" json:"has_unconscious"`
	HasBleeding       bool              `gorm:"default:false" json:"has_bleeding"`
	HasFracture       bool              `gorm:"default:false" json:"has_fracture"`
	IsTrapped         bool              `gorm:"default:false" json:"is_trapped"`
	HazardPresent     bool              `gorm:"default:false" json:"hazard_present"`
	HazardDescription *string           `gorm:"type:text" json:"hazard_description,omitempty"`
	CreatedAt         time.Time         `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for CaseAccidentDetails
func (CaseAccidentDetails) TableName() string {
	return "case_accident_details"
}

// CaseVolunteer represents a volunteer who has accepted a case
type CaseVolunteer struct {
	ID                uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID            uuid.UUID            `gorm:"type:uuid;not null;index" json:"case_id"`
	VolunteerID       uuid.UUID            `gorm:"type:uuid;not null;index" json:"volunteer_id"`
	Status            enum.VolunteerStatus `gorm:"type:varchar(30);not null;default:'accepted'" json:"status"`
	AcceptedLatitude  *float64             `gorm:"type:decimal(10,8)" json:"accepted_latitude,omitempty"`
	AcceptedLongitude *float64             `gorm:"type:decimal(11,8)" json:"accepted_longitude,omitempty"`
	DistanceKm        *float64             `gorm:"type:decimal(10,2)" json:"distance_km,omitempty"`
	AcceptedAt        time.Time            `gorm:"autoCreateTime" json:"accepted_at"`
	ArrivedAt         *time.Time           `json:"arrived_at,omitempty"`
	CompletedAt       *time.Time           `json:"completed_at,omitempty"`
	Note              *string              `gorm:"type:text" json:"note,omitempty"`

	// Relations
	Volunteer *User `gorm:"foreignKey:VolunteerID" json:"volunteer,omitempty"`
}

// TableName returns the table name for CaseVolunteer
func (CaseVolunteer) TableName() string {
	return "case_volunteers"
}

// GetAcceptedLocation returns a GeoPoint from the volunteer's accepted coordinates
func (cv *CaseVolunteer) GetAcceptedLocation() *GeoPoint {
	if cv.AcceptedLatitude == nil || cv.AcceptedLongitude == nil {
		return nil
	}
	return NewGeoPoint(*cv.AcceptedLatitude, *cv.AcceptedLongitude)
}

// SetAcceptedLocation sets the volunteer's accepted coordinates from a GeoPoint
func (cv *CaseVolunteer) SetAcceptedLocation(loc *GeoPoint) {
	if loc == nil {
		cv.AcceptedLatitude = nil
		cv.AcceptedLongitude = nil
		return
	}
	cv.AcceptedLatitude = &loc.Latitude
	cv.AcceptedLongitude = &loc.Longitude
}

// CaseUpdate represents an update/timeline entry for a case
type CaseUpdate struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"case_id"`
	UpdateType enum.UpdateType  `gorm:"type:varchar(30);not null" json:"update_type"`
	UserID     *uuid.UUID       `gorm:"type:uuid" json:"user_id,omitempty"`
	UserName   *string          `gorm:"-" json:"user_name,omitempty"`
	Content    *string          `gorm:"type:text" json:"content,omitempty"`
	OldStatus  *enum.CaseStatus `gorm:"type:varchar(30)" json:"old_status,omitempty"`
	NewStatus  *enum.CaseStatus `gorm:"type:varchar(30)" json:"new_status,omitempty"`
	MediaURLs  pq.StringArray   `gorm:"type:text[]" json:"media_urls,omitempty"`
	CreatedAt  time.Time        `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for CaseUpdate
func (CaseUpdate) TableName() string {
	return "case_updates"
}

// CaseNearby represents a case with distance information for nearby queries
type CaseNearby struct {
	ID             uuid.UUID         `json:"id"`
	CaseType       enum.CaseType     `json:"case_type"`
	Title          string            `json:"title"`
	Urgency        enum.UrgencyLevel `json:"urgency"`
	Status         enum.CaseStatus   `json:"status"`
	DistanceKm     float64           `json:"distance_km"`
	VolunteerCount int               `json:"volunteer_count"`
	CreatedAt      time.Time         `json:"created_at"`
	Latitude       float64           `json:"latitude"`
	Longitude      float64           `json:"longitude"`
}

// GetLocation returns a GeoPoint from the case's coordinates
func (c *CaseNearby) GetLocation() *GeoPoint {
	return NewGeoPoint(c.Latitude, c.Longitude)
}

// CaseComment represents a comment on a case
type CaseComment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"case_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for CaseComment
func (CaseComment) TableName() string {
	return "case_comments"
}
