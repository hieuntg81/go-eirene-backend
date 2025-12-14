package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/internal/domain/enum"
	"bamboo-rescue/internal/handler/dto/request"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/repository"
	"bamboo-rescue/pkg/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*entity.User, *jwt.TokenPair, error)
	Login(ctx context.Context, req *request.LoginRequest) (*entity.User, *jwt.TokenPair, error)
	OAuth(ctx context.Context, req *request.OAuthRequest) (*entity.User, *jwt.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.User, *jwt.TokenPair, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type authService struct {
	userRepo   repository.UserRepository
	jwtService *jwt.Service
	log        *zap.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepository, jwtSvc *jwt.Service, log *zap.Logger) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtSvc,
		log:        log,
	}
}

func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*entity.User, *jwt.TokenPair, error) {
	// Check if email already exists
	if req.Email != nil {
		existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return nil, nil, middleware.ErrEmailExists
		}
	}

	// Check if phone already exists
	if req.Phone != nil {
		existing, err := s.userRepo.GetByPhone(ctx, *req.Phone)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return nil, nil, middleware.ErrPhoneExists
		}
	}

	// Require at least email or phone
	if req.Email == nil && req.Phone == nil {
		return nil, nil, middleware.NewAppError("VALIDATION_ERROR", "Email or phone is required", 400)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return nil, nil, errors.New("failed to create user")
	}

	passwordHash := string(hashedPassword)

	// Create user
	user := &entity.User{
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: &passwordHash,
		DisplayName:  req.DisplayName,
		Role:         enum.UserRoleBoth,
		IsAvailable:  false,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("Failed to create user", zap.Error(err))
		return nil, nil, errors.New("failed to create user")
	}

	// Create default preferences
	prefs := &entity.UserPreferences{
		UserID:               user.ID,
		PushEnabled:          true,
		CaseTypes:            []string{"animal", "flood", "accident"},
		NotificationRadiusKm: 10,
		UseCurrentLocation:   true,
	}
	if err := s.userRepo.CreatePreferences(ctx, prefs); err != nil {
		s.log.Warn("Failed to create user preferences", zap.Error(err))
	}

	// Generate tokens
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, email)
	if err != nil {
		s.log.Error("Failed to generate tokens", zap.Error(err))
		return nil, nil, errors.New("failed to generate tokens")
	}

	s.log.Info("User registered", zap.String("user_id", user.ID.String()))

	return user, tokens, nil
}

func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*entity.User, *jwt.TokenPair, error) {
	var user *entity.User
	var err error

	// Find user by email or phone
	if req.Email != nil {
		user, err = s.userRepo.GetByEmail(ctx, *req.Email)
	} else if req.Phone != nil {
		user, err = s.userRepo.GetByPhone(ctx, *req.Phone)
	} else {
		return nil, nil, middleware.NewAppError("VALIDATION_ERROR", "Email or phone is required", 400)
	}

	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, middleware.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, middleware.NewAppError("USER_INACTIVE", "User account is inactive", 403)
	}

	// Verify password
	if user.PasswordHash == nil {
		return nil, nil, middleware.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, middleware.ErrInvalidCredentials
	}

	// Generate tokens
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, email)
	if err != nil {
		s.log.Error("Failed to generate tokens", zap.Error(err))
		return nil, nil, errors.New("failed to generate tokens")
	}

	s.log.Info("User logged in", zap.String("user_id", user.ID.String()))

	return user, tokens, nil
}

func (s *authService) OAuth(ctx context.Context, req *request.OAuthRequest) (*entity.User, *jwt.TokenPair, error) {
	// Validate OAuth token with provider
	oauthUser, err := s.validateOAuthToken(req.Provider, req.Token)
	if err != nil {
		return nil, nil, middleware.NewAppError("OAUTH_ERROR", "Invalid OAuth token", 401)
	}

	// Check if user exists with this OAuth
	user, err := s.userRepo.GetByOAuth(ctx, req.Provider, oauthUser.ID)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		// Check if user exists with same email
		if oauthUser.Email != "" {
			user, err = s.userRepo.GetByEmail(ctx, oauthUser.Email)
			if err != nil {
				return nil, nil, err
			}
		}

		if user == nil {
			// Create new user
			user = &entity.User{
				Email:         &oauthUser.Email,
				OAuthProvider: &req.Provider,
				OAuthID:       &oauthUser.ID,
				DisplayName:   oauthUser.Name,
				AvatarURL:     oauthUser.AvatarURL,
				Role:          enum.UserRoleBoth,
				IsAvailable:   false,
				IsActive:      true,
			}

			if err := s.userRepo.Create(ctx, user); err != nil {
				s.log.Error("Failed to create OAuth user", zap.Error(err))
				return nil, nil, errors.New("failed to create user")
			}

			// Create default preferences
			prefs := &entity.UserPreferences{
				UserID:               user.ID,
				PushEnabled:          true,
				CaseTypes:            []string{"animal", "flood", "accident"},
				NotificationRadiusKm: 10,
				UseCurrentLocation:   true,
			}
			if err := s.userRepo.CreatePreferences(ctx, prefs); err != nil {
				s.log.Warn("Failed to create user preferences", zap.Error(err))
			}
		} else {
			// Link OAuth to existing user
			user.OAuthProvider = &req.Provider
			user.OAuthID = &oauthUser.ID
			if err := s.userRepo.Update(ctx, user); err != nil {
				s.log.Warn("Failed to link OAuth", zap.Error(err))
			}
		}
	}

	// Generate tokens
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, email)
	if err != nil {
		s.log.Error("Failed to generate tokens", zap.Error(err))
		return nil, nil, errors.New("failed to generate tokens")
	}

	s.log.Info("User OAuth login", zap.String("user_id", user.ID.String()), zap.String("provider", req.Provider))

	return user, tokens, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*entity.User, *jwt.TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, middleware.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, middleware.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, nil, middleware.NewAppError("USER_INACTIVE", "User account is inactive", 403)
	}

	// Generate new tokens
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, email)
	if err != nil {
		s.log.Error("Failed to generate tokens", zap.Error(err))
		return nil, nil, errors.New("failed to generate tokens")
	}

	return user, tokens, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	// In a stateless JWT system, logout is handled client-side
	// Optionally, we can delete push tokens
	if err := s.userRepo.DeletePushTokensByUser(ctx, userID); err != nil {
		s.log.Warn("Failed to delete push tokens on logout", zap.Error(err))
	}

	s.log.Info("User logged out", zap.String("user_id", userID.String()))
	return nil
}

// OAuthUserInfo represents user info from OAuth provider
type OAuthUserInfo struct {
	ID        string
	Email     string
	Name      string
	AvatarURL *string
}

// validateOAuthToken validates token with OAuth provider and returns user info
func (s *authService) validateOAuthToken(provider, token string) (*OAuthUserInfo, error) {
	// TODO: Implement actual OAuth validation
	// This is a placeholder that should be replaced with actual provider APIs
	switch provider {
	case "google":
		return s.validateGoogleToken(token)
	case "facebook":
		return s.validateFacebookToken(token)
	default:
		return nil, errors.New("unsupported OAuth provider")
	}
}

func (s *authService) validateGoogleToken(token string) (*OAuthUserInfo, error) {
	// TODO: Implement Google token validation
	// Use Google OAuth2 API to validate token and get user info
	// For now, return a placeholder
	return nil, errors.New("Google OAuth not implemented")
}

func (s *authService) validateFacebookToken(token string) (*OAuthUserInfo, error) {
	// TODO: Implement Facebook token validation
	// Use Facebook Graph API to validate token and get user info
	return nil, errors.New("Facebook OAuth not implemented")
}
