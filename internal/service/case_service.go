package service

import (
	"context"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"bamboo-rescue/internal/handler/dto/request"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/repository"
	"go.uber.org/zap"
)

// CaseService defines the interface for case operations
type CaseService interface {
	Create(ctx context.Context, req *request.CreateCaseRequest, userID *uuid.UUID) (*entity.Case, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Case, error)
	GetNearby(ctx context.Context, req *request.GetNearbyCasesRequest) ([]entity.CaseNearby, error)
	GetCases(ctx context.Context, req *request.GetCasesRequest) ([]entity.Case, int64, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *request.UpdateCaseRequest) (*entity.Case, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Accept(ctx context.Context, caseID, volunteerID uuid.UUID, req *request.AcceptCaseRequest) error
	Withdraw(ctx context.Context, caseID, volunteerID uuid.UUID) error
	UpdateVolunteerStatus(ctx context.Context, caseID, volunteerID uuid.UUID, req *request.UpdateVolunteerStatusRequest) error
	GetVolunteers(ctx context.Context, caseID uuid.UUID) ([]entity.CaseVolunteer, error)
	CreateUpdate(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, req *request.CreateCaseUpdateRequest) (*entity.CaseUpdate, error)
	GetUpdates(ctx context.Context, caseID uuid.UUID, page *request.PaginationRequest) ([]entity.CaseUpdate, int64, error)
	GetUserReportedCases(ctx context.Context, userID uuid.UUID, page *request.PaginationRequest) ([]entity.Case, int64, error)
	GetUserAcceptedCases(ctx context.Context, userID uuid.UUID, page *request.PaginationRequest) ([]entity.Case, int64, error)

	// Comments
	CreateComment(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, req *request.CreateCommentRequest) (*entity.CaseComment, error)
	GetComments(ctx context.Context, caseID uuid.UUID, page *request.PaginationRequest) ([]entity.CaseComment, int64, error)
	DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error
}

type caseService struct {
	caseRepo        repository.CaseRepository
	userRepo        repository.UserRepository
	notificationSvc NotificationService
	fcmSvc          FCMService
	log             *zap.Logger
}

// NewCaseService creates a new CaseService
func NewCaseService(
	caseRepo repository.CaseRepository,
	userRepo repository.UserRepository,
	notificationSvc NotificationService,
	fcmSvc FCMService,
	log *zap.Logger,
) CaseService {
	return &caseService{
		caseRepo:        caseRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		fcmSvc:          fcmSvc,
		log:             log,
	}
}

func (s *caseService) Create(ctx context.Context, req *request.CreateCaseRequest, userID *uuid.UUID) (*entity.Case, error) {
	// Build case entity
	c := &entity.Case{
		CaseType:  req.CaseType,
		Urgency:   req.Urgency,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Address:   req.Address,
		LocationNote:  req.LocationNote,
		Title:         req.Title,
		Description:   req.Description,
		ReporterID:    userID,
		ReporterName:  req.ReporterName,
		ReporterPhone: req.ReporterPhone,
		IsAnonymous:   req.IsAnonymous,
		Status:        enum.CaseStatusPending,
	}

	// Add type-specific details
	switch req.CaseType {
	case enum.CaseTypeAnimal:
		if req.AnimalType != nil && req.AnimalCondition != nil {
			c.AnimalDetails = &entity.CaseAnimalDetails{
				AnimalType:           *req.AnimalType,
				AnimalTypeOther:      req.AnimalTypeOther,
				Condition:            *req.AnimalCondition,
				ConditionDescription: req.ConditionDescription,
				EstimatedCount:       getIntOrDefault(req.EstimatedCount, 1),
			}
		}
	case enum.CaseTypeFlood:
		c.FloodDetails = &entity.CaseFloodDetails{
			PeopleCount:  req.PeopleCount,
			HasChildren:  getBoolOrDefault(req.HasChildren, false),
			HasElderly:   getBoolOrDefault(req.HasElderly, false),
			HasDisabled:  getBoolOrDefault(req.HasDisabled, false),
			WaterLevelCm: req.WaterLevelCm,
			FloorLevel:   req.FloorLevel,
			HasPower:     req.HasPower,
			HasFoodWater: req.HasFoodWater,
			MedicalNeeds: req.MedicalNeeds,
		}
	case enum.CaseTypeAccident:
		if req.AccidentType != nil {
			c.AccidentDetails = &entity.CaseAccidentDetails{
				AccidentType:      *req.AccidentType,
				VictimCount:       getIntOrDefault(req.VictimCount, 1),
				HasUnconscious:    getBoolOrDefault(req.HasUnconscious, false),
				HasBleeding:       getBoolOrDefault(req.HasBleeding, false),
				HasFracture:       getBoolOrDefault(req.HasFracture, false),
				IsTrapped:         getBoolOrDefault(req.IsTrapped, false),
				HazardPresent:     getBoolOrDefault(req.HazardPresent, false),
				HazardDescription: req.HazardDescription,
			}
		}
	}

	// Create case
	if err := s.caseRepo.Create(ctx, c); err != nil {
		s.log.Error("Failed to create case", zap.Error(err))
		return nil, err
	}

	// Increment user's reported cases count
	if userID != nil {
		go func() {
			if err := s.userRepo.IncrementCasesReported(context.Background(), *userID); err != nil {
				s.log.Warn("Failed to increment cases reported", zap.Error(err))
			}
		}()
	}

	// Notify nearby volunteers asynchronously
	go s.notifyNearbyVolunteers(c)

	s.log.Info("Case created",
		zap.String("case_id", c.ID.String()),
		zap.String("type", string(c.CaseType)),
		zap.String("urgency", string(c.Urgency)),
	)

	return c, nil
}

func (s *caseService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Case, error) {
	c, err := s.caseRepo.GetByIDWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, middleware.ErrCaseNotFound
	}
	return c, nil
}

func (s *caseService) GetNearby(ctx context.Context, req *request.GetNearbyCasesRequest) ([]entity.CaseNearby, error) {
	// Set defaults
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 10
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	cases, err := s.caseRepo.GetNearby(ctx, req.Latitude, req.Longitude, radiusKm, req.Types, limit)
	if err != nil {
		s.log.Error("Failed to get nearby cases", zap.Error(err))
		return nil, err
	}

	return cases, nil
}

func (s *caseService) GetCases(ctx context.Context, req *request.GetCasesRequest) ([]entity.Case, int64, error) {
	cases, total, err := s.caseRepo.GetCases(ctx, req.Query, req.Type, req.Status, req.Urgency, req.GetDefaultLimit(), req.GetOffset())
	if err != nil {
		s.log.Error("Failed to get cases", zap.Error(err))
		return nil, 0, err
	}
	return cases, total, nil
}

func (s *caseService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *request.UpdateCaseRequest) (*entity.Case, error) {
	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, middleware.ErrCaseNotFound
	}

	// Check permission - only reporter can update
	if c.ReporterID == nil || *c.ReporterID != userID {
		return nil, middleware.ErrForbidden
	}

	// Update fields
	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.Description != nil {
		c.Description = req.Description
	}
	if req.Urgency != nil {
		c.Urgency = *req.Urgency
	}
	if req.Status != nil {
		oldStatus := c.Status
		c.Status = *req.Status

		// Create status change update
		update := &entity.CaseUpdate{
			CaseID:     c.ID,
			UpdateType: enum.UpdateTypeStatusChange,
			UserID:     &userID,
			OldStatus:  &oldStatus,
			NewStatus:  req.Status,
		}
		if err := s.caseRepo.CreateUpdate(ctx, update); err != nil {
			s.log.Warn("Failed to create status update", zap.Error(err))
		}
	}
	if req.Address != nil {
		c.Address = req.Address
	}
	if req.LocationNote != nil {
		c.LocationNote = req.LocationNote
	}

	if err := s.caseRepo.Update(ctx, c); err != nil {
		s.log.Error("Failed to update case", zap.Error(err))
		return nil, err
	}

	return c, nil
}

func (s *caseService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return middleware.ErrCaseNotFound
	}

	// Check permission - only reporter can delete
	if c.ReporterID == nil || *c.ReporterID != userID {
		return middleware.ErrForbidden
	}

	if err := s.caseRepo.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete case", zap.Error(err))
		return err
	}

	s.log.Info("Case deleted", zap.String("case_id", id.String()))

	return nil
}

func (s *caseService) Accept(ctx context.Context, caseID, volunteerID uuid.UUID, req *request.AcceptCaseRequest) error {
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return err
	}
	if c == nil {
		return middleware.ErrCaseNotFound
	}

	// Check if case is still open
	if !c.Status.IsActive() {
		return middleware.NewAppError("CASE_CLOSED", "This case is no longer accepting volunteers", 400)
	}

	// Check if already at max volunteers
	if c.VolunteerCount >= c.MaxVolunteers {
		return middleware.NewAppError("MAX_VOLUNTEERS", "Maximum volunteers reached for this case", 400)
	}

	// Get volunteer info
	volunteer, err := s.userRepo.GetByID(ctx, volunteerID)
	if err != nil {
		return err
	}
	if volunteer == nil {
		return middleware.ErrUserNotFound
	}

	// Calculate location and distance if provided
	var latitude, longitude, distanceKm *float64
	if req.Latitude != nil && req.Longitude != nil {
		latitude = req.Latitude
		longitude = req.Longitude
		caseLocation := c.GetLocation()
		volunteerLocation := entity.NewGeoPoint(*req.Latitude, *req.Longitude)
		distance := caseLocation.DistanceKm(volunteerLocation)
		distanceKm = &distance
	}

	// Check if already accepted
	existing, err := s.caseRepo.GetVolunteer(ctx, caseID, volunteerID)
	if err != nil {
		return err
	}

	if existing != nil {
		// If previously withdrawn, allow rejoin
		if existing.Status == enum.VolunteerStatusWithdrawn {
			if err := s.caseRepo.ReactivateVolunteer(ctx, caseID, volunteerID, latitude, longitude, distanceKm); err != nil {
				s.log.Error("Failed to reactivate volunteer", zap.Error(err))
				return err
			}
		} else {
			return middleware.NewAppError("ALREADY_ACCEPTED", "You have already accepted this case", 400)
		}
	} else {
		// Create new volunteer entry
		cv := &entity.CaseVolunteer{
			CaseID:            caseID,
			VolunteerID:       volunteerID,
			Status:            enum.VolunteerStatusAccepted,
			AcceptedLatitude:  latitude,
			AcceptedLongitude: longitude,
			DistanceKm:        distanceKm,
		}

		if err := s.caseRepo.AddVolunteer(ctx, cv); err != nil {
			s.log.Error("Failed to add volunteer", zap.Error(err))
			return err
		}
	}

	// Create update entry
	content := volunteer.DisplayName + " đã nhận case này"
	update := &entity.CaseUpdate{
		CaseID:     caseID,
		UpdateType: enum.UpdateTypeVolunteerJoined,
		UserID:     &volunteerID,
		Content:    &content,
	}
	if err := s.caseRepo.CreateUpdate(ctx, update); err != nil {
		s.log.Warn("Failed to create update", zap.Error(err))
	}

	// Notify reporter
	go s.notifyReporterOfAcceptance(c, volunteer)

	s.log.Info("Volunteer accepted case",
		zap.String("case_id", caseID.String()),
		zap.String("volunteer_id", volunteerID.String()),
	)

	return nil
}

func (s *caseService) Withdraw(ctx context.Context, caseID, volunteerID uuid.UUID) error {
	cv, err := s.caseRepo.GetVolunteer(ctx, caseID, volunteerID)
	if err != nil {
		return err
	}
	if cv == nil {
		return middleware.NewAppError("NOT_ACCEPTED", "You have not accepted this case", 400)
	}

	// Get volunteer info for the update entry
	volunteer, _ := s.userRepo.GetByID(ctx, volunteerID)
	volunteerName := "A volunteer"
	if volunteer != nil {
		volunteerName = volunteer.DisplayName
	}

	if err := s.caseRepo.RemoveVolunteer(ctx, caseID, volunteerID); err != nil {
		s.log.Error("Failed to withdraw volunteer", zap.Error(err))
		return err
	}

	// Create update entry
	content := volunteerName + " has left this case"
	update := &entity.CaseUpdate{
		CaseID:     caseID,
		UpdateType: enum.UpdateTypeVolunteerWithdrawn,
		UserID:     &volunteerID,
		Content:    &content,
	}
	if err := s.caseRepo.CreateUpdate(ctx, update); err != nil {
		s.log.Warn("Failed to create withdraw update", zap.Error(err))
	}

	s.log.Info("Volunteer withdrew from case",
		zap.String("case_id", caseID.String()),
		zap.String("volunteer_id", volunteerID.String()),
	)

	return nil
}

func (s *caseService) UpdateVolunteerStatus(ctx context.Context, caseID, volunteerID uuid.UUID, req *request.UpdateVolunteerStatusRequest) error {
	cv, err := s.caseRepo.GetVolunteer(ctx, caseID, volunteerID)
	if err != nil {
		return err
	}
	if cv == nil {
		return middleware.NewAppError("NOT_ACCEPTED", "You have not accepted this case", 400)
	}

	if err := s.caseRepo.UpdateVolunteerStatus(ctx, caseID, volunteerID, req.Status); err != nil {
		s.log.Error("Failed to update volunteer status", zap.Error(err))
		return err
	}

	// Create update entry
	volunteer, _ := s.userRepo.GetByID(ctx, volunteerID)
	volunteerName := "Tình nguyện viên"
	if volunteer != nil {
		volunteerName = volunteer.DisplayName
	}

	statusText := getStatusText(req.Status)
	content := volunteerName + " " + statusText
	if req.Note != nil && *req.Note != "" {
		content += ": " + *req.Note
	}

	update := &entity.CaseUpdate{
		CaseID:     caseID,
		UpdateType: enum.UpdateTypeVolunteerUpdate,
		UserID:     &volunteerID,
		Content:    &content,
	}
	if err := s.caseRepo.CreateUpdate(ctx, update); err != nil {
		s.log.Warn("Failed to create update", zap.Error(err))
	}

	// If completed, check if all volunteers completed and update case status
	if req.Status == enum.VolunteerStatusCompleted {
		go s.checkCaseCompletion(caseID, volunteerID)
	}

	s.log.Info("Volunteer status updated",
		zap.String("case_id", caseID.String()),
		zap.String("volunteer_id", volunteerID.String()),
		zap.String("status", string(req.Status)),
	)

	return nil
}

func (s *caseService) GetVolunteers(ctx context.Context, caseID uuid.UUID) ([]entity.CaseVolunteer, error) {
	// Validate case exists
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, middleware.ErrCaseNotFound
	}

	// Get volunteers (repository already filters by case ID and preloads volunteer data)
	volunteers, err := s.caseRepo.GetVolunteersByCaseID(ctx, caseID)
	if err != nil {
		s.log.Error("Failed to get volunteers", zap.Error(err))
		return nil, err
	}

	// Filter out withdrawn volunteers for public view
	activeVolunteers := make([]entity.CaseVolunteer, 0)
	for _, v := range volunteers {
		if v.Status != enum.VolunteerStatusWithdrawn {
			activeVolunteers = append(activeVolunteers, v)
		}
	}

	return activeVolunteers, nil
}

func (s *caseService) CreateUpdate(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, req *request.CreateCaseUpdateRequest) (*entity.CaseUpdate, error) {
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, middleware.ErrCaseNotFound
	}

	// Determine update type based on user role in case
	updateType := enum.UpdateTypeReporterUpdate
	if c.ReporterID == nil || *c.ReporterID != userID {
		// Check if user is a volunteer
		cv, _ := s.caseRepo.GetVolunteer(ctx, caseID, userID)
		if cv != nil {
			updateType = enum.UpdateTypeVolunteerUpdate
		}
	}

	update := &entity.CaseUpdate{
		CaseID:     caseID,
		UpdateType: updateType,
		UserID:     &userID,
		Content:    &req.Content,
		MediaURLs:  req.MediaURLs,
	}

	if err := s.caseRepo.CreateUpdate(ctx, update); err != nil {
		s.log.Error("Failed to create update", zap.Error(err))
		return nil, err
	}

	return update, nil
}

func (s *caseService) GetUpdates(ctx context.Context, caseID uuid.UUID, page *request.PaginationRequest) ([]entity.CaseUpdate, int64, error) {
	return s.caseRepo.GetUpdates(ctx, caseID, page.GetDefaultLimit(), page.GetOffset())
}

func (s *caseService) GetUserReportedCases(ctx context.Context, userID uuid.UUID, page *request.PaginationRequest) ([]entity.Case, int64, error) {
	return s.caseRepo.GetUserReportedCases(ctx, userID, page.GetDefaultLimit(), page.GetOffset())
}

func (s *caseService) GetUserAcceptedCases(ctx context.Context, userID uuid.UUID, page *request.PaginationRequest) ([]entity.Case, int64, error) {
	return s.caseRepo.GetUserAcceptedCases(ctx, userID, page.GetDefaultLimit(), page.GetOffset())
}

// Helper functions

func (s *caseService) notifyNearbyVolunteers(c *entity.Case) {
	if s.notificationSvc == nil {
		return
	}

	ctx := context.Background()

	// Find nearby volunteers
	volunteers, err := s.userRepo.FindAvailableVolunteers(ctx, c.Latitude, c.Longitude, 10, string(c.CaseType), 100)
	if err != nil {
		s.log.Warn("Failed to find nearby volunteers", zap.Error(err))
		return
	}

	for _, v := range volunteers {
		payload := entity.NewCaseNotificationPayload(c, v.DistanceKm)
		if err := s.fcmSvc.SendToUser(ctx, v.User.ID, payload); err != nil {
			s.log.Warn("Failed to send notification", zap.Error(err), zap.String("user_id", v.User.ID.String()))
		}
	}

	s.log.Info("Notified nearby volunteers",
		zap.String("case_id", c.ID.String()),
		zap.Int("count", len(volunteers)),
	)
}

func (s *caseService) notifyReporterOfAcceptance(c *entity.Case, volunteer *entity.User) {
	if s.fcmSvc == nil || c.ReporterID == nil {
		return
	}

	ctx := context.Background()
	payload := entity.CaseAcceptedNotificationPayload(c, volunteer.DisplayName)
	if err := s.fcmSvc.SendToUser(ctx, *c.ReporterID, payload); err != nil {
		s.log.Warn("Failed to notify reporter", zap.Error(err))
	}
}

func (s *caseService) checkCaseCompletion(caseID, volunteerID uuid.UUID) {
	ctx := context.Background()

	// Get all volunteers
	volunteers, err := s.caseRepo.GetVolunteersByCaseID(ctx, caseID)
	if err != nil {
		s.log.Warn("Failed to get volunteers", zap.Error(err))
		return
	}

	// Check if all volunteers are completed or withdrawn
	allDone := true
	for _, v := range volunteers {
		if v.Status != enum.VolunteerStatusCompleted && v.Status != enum.VolunteerStatusWithdrawn {
			allDone = false
			break
		}
	}

	if allDone && len(volunteers) > 0 {
		// Update case status to resolved
		if err := s.caseRepo.UpdateStatus(ctx, caseID, enum.CaseStatusResolved); err != nil {
			s.log.Warn("Failed to resolve case", zap.Error(err))
			return
		}

		// Increment resolved count for all completed volunteers
		for _, v := range volunteers {
			if v.Status == enum.VolunteerStatusCompleted {
				if err := s.userRepo.IncrementCasesResolved(ctx, v.VolunteerID); err != nil {
					s.log.Warn("Failed to increment resolved count", zap.Error(err))
				}
			}
		}

		s.log.Info("Case resolved", zap.String("case_id", caseID.String()))
	}
}

func getIntOrDefault(ptr *int, def int) int {
	if ptr == nil {
		return def
	}
	return *ptr
}

func getBoolOrDefault(ptr *bool, def bool) bool {
	if ptr == nil {
		return def
	}
	return *ptr
}

func getStatusText(status enum.VolunteerStatus) string {
	switch status {
	case enum.VolunteerStatusEnRoute:
		return "đang trên đường đến"
	case enum.VolunteerStatusOnSite:
		return "đã đến hiện trường"
	case enum.VolunteerStatusHandling:
		return "đang xử lý"
	case enum.VolunteerStatusCompleted:
		return "đã hoàn thành"
	case enum.VolunteerStatusWithdrawn:
		return "đã rút khỏi case"
	default:
		return "cập nhật trạng thái"
	}
}

// Comment methods

func (s *caseService) CreateComment(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, req *request.CreateCommentRequest) (*entity.CaseComment, error) {
	// Validate case exists
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, middleware.ErrCaseNotFound
	}

	comment := &entity.CaseComment{
		CaseID:  caseID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := s.caseRepo.CreateComment(ctx, comment); err != nil {
		s.log.Error("Failed to create comment", zap.Error(err))
		return nil, err
	}

	// Load user info for response
	comment.User, _ = s.userRepo.GetByID(ctx, userID)

	s.log.Info("Comment created",
		zap.String("case_id", caseID.String()),
		zap.String("user_id", userID.String()),
	)

	return comment, nil
}

func (s *caseService) GetComments(ctx context.Context, caseID uuid.UUID, page *request.PaginationRequest) ([]entity.CaseComment, int64, error) {
	// Validate case exists
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, 0, err
	}
	if c == nil {
		return nil, 0, middleware.ErrCaseNotFound
	}

	return s.caseRepo.GetCommentsByCaseID(ctx, caseID, page.GetDefaultLimit(), page.GetOffset())
}

func (s *caseService) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	if err := s.caseRepo.DeleteComment(ctx, commentID, userID); err != nil {
		s.log.Error("Failed to delete comment", zap.Error(err))
		return middleware.NewAppError("DELETE_FAILED", "Failed to delete comment", 400)
	}

	s.log.Info("Comment deleted",
		zap.String("comment_id", commentID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}
