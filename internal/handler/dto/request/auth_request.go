package request

// LoginRequest represents login request body
type LoginRequest struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone" validate:"omitempty,min=10,max=20"`
	Password string  `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents registration request body
type RegisterRequest struct {
	Email       *string `json:"email" validate:"omitempty,email"`
	Phone       *string `json:"phone" validate:"omitempty,min=10,max=20"`
	Password    string  `json:"password" validate:"required,min=6"`
	DisplayName string  `json:"display_name" validate:"required,min=2,max=100"`
}

// OAuthRequest represents OAuth login request body
type OAuthRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google facebook"`
	Token    string `json:"token" validate:"required"`
}

// RefreshTokenRequest represents token refresh request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
