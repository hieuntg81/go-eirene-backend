package enum

import (
	"database/sql/driver"
	"fmt"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleReporter  UserRole = "reporter"
	UserRoleVolunteer UserRole = "volunteer"
	UserRoleBoth      UserRole = "both"
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleReporter, UserRoleVolunteer, UserRoleBoth:
		return true
	}
	return false
}

func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = UserRoleBoth
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan UserRole: %v", value)
	}
	*r = UserRole(str)
	return nil
}

// CaseType represents the type of a rescue case
type CaseType string

const (
	CaseTypeAnimal   CaseType = "animal"
	CaseTypeFlood    CaseType = "flood"
	CaseTypeAccident CaseType = "accident"
)

func (t CaseType) IsValid() bool {
	switch t {
	case CaseTypeAnimal, CaseTypeFlood, CaseTypeAccident:
		return true
	}
	return false
}

func (t CaseType) Value() (driver.Value, error) {
	return string(t), nil
}

func (t *CaseType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan CaseType: %v", value)
	}
	*t = CaseType(str)
	return nil
}

// CaseStatus represents the status of a case
type CaseStatus string

const (
	CaseStatusPending    CaseStatus = "pending"
	CaseStatusAccepted   CaseStatus = "accepted"
	CaseStatusInProgress CaseStatus = "in_progress"
	CaseStatusResolved   CaseStatus = "resolved"
	CaseStatusCancelled  CaseStatus = "cancelled"
	CaseStatusExpired    CaseStatus = "expired"
)

func (s CaseStatus) IsValid() bool {
	switch s {
	case CaseStatusPending, CaseStatusAccepted, CaseStatusInProgress, CaseStatusResolved, CaseStatusCancelled, CaseStatusExpired:
		return true
	}
	return false
}

func (s CaseStatus) IsActive() bool {
	return s == CaseStatusPending || s == CaseStatusAccepted || s == CaseStatusInProgress
}

func (s CaseStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *CaseStatus) Scan(value interface{}) error {
	if value == nil {
		*s = CaseStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan CaseStatus: %v", value)
	}
	*s = CaseStatus(str)
	return nil
}

// UrgencyLevel represents the urgency of a case
type UrgencyLevel string

const (
	UrgencyLow      UrgencyLevel = "low"
	UrgencyMedium   UrgencyLevel = "medium"
	UrgencyHigh     UrgencyLevel = "high"
	UrgencyCritical UrgencyLevel = "critical"
)

func (u UrgencyLevel) IsValid() bool {
	switch u {
	case UrgencyLow, UrgencyMedium, UrgencyHigh, UrgencyCritical:
		return true
	}
	return false
}

func (u UrgencyLevel) Priority() int {
	switch u {
	case UrgencyCritical:
		return 4
	case UrgencyHigh:
		return 3
	case UrgencyMedium:
		return 2
	case UrgencyLow:
		return 1
	default:
		return 0
	}
}

func (u UrgencyLevel) Value() (driver.Value, error) {
	return string(u), nil
}

func (u *UrgencyLevel) Scan(value interface{}) error {
	if value == nil {
		*u = UrgencyMedium
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan UrgencyLevel: %v", value)
	}
	*u = UrgencyLevel(str)
	return nil
}

// VolunteerStatus represents the status of a volunteer in a case
type VolunteerStatus string

const (
	VolunteerStatusAccepted  VolunteerStatus = "accepted"
	VolunteerStatusEnRoute   VolunteerStatus = "en_route"
	VolunteerStatusOnSite    VolunteerStatus = "on_site"
	VolunteerStatusHandling  VolunteerStatus = "handling"
	VolunteerStatusCompleted VolunteerStatus = "completed"
	VolunteerStatusWithdrawn VolunteerStatus = "withdrawn"
)

func (v VolunteerStatus) IsValid() bool {
	switch v {
	case VolunteerStatusAccepted, VolunteerStatusEnRoute, VolunteerStatusOnSite, VolunteerStatusHandling, VolunteerStatusCompleted, VolunteerStatusWithdrawn:
		return true
	}
	return false
}

func (v VolunteerStatus) Value() (driver.Value, error) {
	return string(v), nil
}

func (v *VolunteerStatus) Scan(value interface{}) error {
	if value == nil {
		*v = VolunteerStatusAccepted
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan VolunteerStatus: %v", value)
	}
	*v = VolunteerStatus(str)
	return nil
}

// AnimalType represents the type of animal
type AnimalType string

const (
	AnimalTypeDog   AnimalType = "dog"
	AnimalTypeCat   AnimalType = "cat"
	AnimalTypeBird  AnimalType = "bird"
	AnimalTypeOther AnimalType = "other"
)

func (a AnimalType) IsValid() bool {
	switch a {
	case AnimalTypeDog, AnimalTypeCat, AnimalTypeBird, AnimalTypeOther:
		return true
	}
	return false
}

func (a AnimalType) Value() (driver.Value, error) {
	return string(a), nil
}

func (a *AnimalType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan AnimalType: %v", value)
	}
	*a = AnimalType(str)
	return nil
}

// AnimalCondition represents the condition of an animal
type AnimalCondition string

const (
	AnimalConditionInjured   AnimalCondition = "injured"
	AnimalConditionTrapped   AnimalCondition = "trapped"
	AnimalConditionSick      AnimalCondition = "sick"
	AnimalConditionAbandoned AnimalCondition = "abandoned"
	AnimalConditionOther     AnimalCondition = "other"
)

func (a AnimalCondition) IsValid() bool {
	switch a {
	case AnimalConditionInjured, AnimalConditionTrapped, AnimalConditionSick, AnimalConditionAbandoned, AnimalConditionOther:
		return true
	}
	return false
}

func (a AnimalCondition) Value() (driver.Value, error) {
	return string(a), nil
}

func (a *AnimalCondition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan AnimalCondition: %v", value)
	}
	*a = AnimalCondition(str)
	return nil
}

// AccidentType represents the type of accident
type AccidentType string

const (
	AccidentTypeTraffic  AccidentType = "traffic"
	AccidentTypeFall     AccidentType = "fall"
	AccidentTypeFire     AccidentType = "fire"
	AccidentTypeDrowning AccidentType = "drowning"
	AccidentTypeElectric AccidentType = "electric"
	AccidentTypeOther    AccidentType = "other"
)

func (a AccidentType) IsValid() bool {
	switch a {
	case AccidentTypeTraffic, AccidentTypeFall, AccidentTypeFire, AccidentTypeDrowning, AccidentTypeElectric, AccidentTypeOther:
		return true
	}
	return false
}

func (a AccidentType) Value() (driver.Value, error) {
	return string(a), nil
}

func (a *AccidentType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan AccidentType: %v", value)
	}
	*a = AccidentType(str)
	return nil
}

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

func (m MediaType) IsValid() bool {
	switch m {
	case MediaTypeImage, MediaTypeVideo:
		return true
	}
	return false
}

func (m MediaType) Value() (driver.Value, error) {
	return string(m), nil
}

func (m *MediaType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan MediaType: %v", value)
	}
	*m = MediaType(str)
	return nil
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeNewCaseNearby   NotificationType = "new_case_nearby"
	NotificationTypeCaseAccepted    NotificationType = "case_accepted"
	NotificationTypeCaseUpdate      NotificationType = "case_update"
	NotificationTypeCaseResolved    NotificationType = "case_resolved"
	NotificationTypeVolunteerJoined NotificationType = "volunteer_joined"
	NotificationTypeSystem          NotificationType = "system"
)

func (n NotificationType) IsValid() bool {
	switch n {
	case NotificationTypeNewCaseNearby, NotificationTypeCaseAccepted, NotificationTypeCaseUpdate, NotificationTypeCaseResolved, NotificationTypeVolunteerJoined, NotificationTypeSystem:
		return true
	}
	return false
}

func (n NotificationType) Value() (driver.Value, error) {
	return string(n), nil
}

func (n *NotificationType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan NotificationType: %v", value)
	}
	*n = NotificationType(str)
	return nil
}

// DevicePlatform represents the platform of a device
type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformWeb     DevicePlatform = "web"
)

func (d DevicePlatform) IsValid() bool {
	switch d {
	case DevicePlatformIOS, DevicePlatformAndroid, DevicePlatformWeb:
		return true
	}
	return false
}

func (d DevicePlatform) Value() (driver.Value, error) {
	return string(d), nil
}

func (d *DevicePlatform) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan DevicePlatform: %v", value)
	}
	*d = DevicePlatform(str)
	return nil
}

// UpdateType represents the type of case update
type UpdateType string

const (
	UpdateTypeStatusChange       UpdateType = "status_change"
	UpdateTypeVolunteerJoined    UpdateType = "volunteer_joined"
	UpdateTypeVolunteerUpdate    UpdateType = "volunteer_update"
	UpdateTypeVolunteerWithdrawn UpdateType = "volunteer_withdrawn"
	UpdateTypeReporterUpdate     UpdateType = "reporter_update"
	UpdateTypeSystem             UpdateType = "system"
)

func (u UpdateType) IsValid() bool {
	switch u {
	case UpdateTypeStatusChange, UpdateTypeVolunteerJoined, UpdateTypeVolunteerUpdate, UpdateTypeVolunteerWithdrawn, UpdateTypeReporterUpdate, UpdateTypeSystem:
		return true
	}
	return false
}

func (u UpdateType) Value() (driver.Value, error) {
	return string(u), nil
}

func (u *UpdateType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan UpdateType: %v", value)
	}
	*u = UpdateType(str)
	return nil
}
