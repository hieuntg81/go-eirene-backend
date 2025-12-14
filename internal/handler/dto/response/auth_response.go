package response

import (
	"time"

	"bamboo-rescue/internal/domain/entity"
	"bamboo-rescue/pkg/jwt"
)

// AuthResponse represents authentication response
type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// TokenResponse represents tokens in response
type TokenResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// ToAuthResponse converts user and tokens to AuthResponse
func ToAuthResponse(user *entity.User, tokens *jwt.TokenPair) *AuthResponse {
	return &AuthResponse{
		User: *ToUserResponse(user),
		Tokens: TokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    tokens.ExpiresAt,
		},
	}
}
