package users

import (
	"time"

	"github.com/google/uuid"
)

type Users struct {
	Id                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	Role                string     `json:"role"`
	Address             string     `json:"address"`
	CategoryPreferences []string   `json:"category_preferences"`
	CreatedAt           *time.Time `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at"`
}

type UsersLogin struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type OauthUserData struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	Hd            string `json:"hd"`
}

type LoginResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	RefreshToken         string    `json:"refresh_token"`
	RefreshTokenExpiryAt time.Time `json:"refresh_token_expiry_at"`
	*Users
}

type RequestUsers struct {
	Search string    `json:"search"`
	UserId uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Page   int       `json:"page"`
	Limit  int       `json:"limit"`
	Role   string    `json:"role"`
}
