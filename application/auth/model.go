package auth

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Login represents a login request data
type Login struct {
	Username   *string `json:"username" binding:"required"`
	Password   *string `json:"password" binding:"required"`
	UserAgent  *string `json:"-"`
	IPAddress  *string `json:"-"`
	IPLocation *string `json:"-"`
}

type session struct {
	ID               *string    `json:"session_id"`
	CreatedAt        *time.Time `json:"created_at"`
	SessionExpiresAt *time.Time `json:"expires_at"`
	UserID           *string    `json:"user_id"`
	Name             *string    `json:"user_name"`
	Username         *string    `json:"username"`
	jwt.StandardClaims
}
