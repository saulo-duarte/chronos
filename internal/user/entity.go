package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                          uuid.UUID `json:"id" db:"id"`
	ProviderID                  string    `json:"provider_id" db:"provider_id"`
	Username                    string    `json:"username" db:"username"`
	Email                       string    `json:"email" db:"email"`
	AvatarURL                   string    `json:"avatar_url" db:"avatar_url"`
	Role                        string    `json:"role" db:"role"`
	EncryptedGoogleAccessToken  string    `json:"-" db:"encrypted_google_access_token"`
	EncryptedGoogleRefreshToken string    `json:"-" db:"encrypted_google_refresh_token"`
	CreatedAt                   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at" db:"updated_at"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u *User) HasRole(role string) bool {
	return u.Role == role
}

func (u *User) IsAdmin() bool {
	return u.HasRole("ADMIN")
}

func (u *User) CanAccess(resource string) bool {
	switch resource {
	case "admin":
		return u.IsAdmin()
	default:
		return false
	}
}
