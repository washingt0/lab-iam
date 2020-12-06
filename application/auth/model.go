package auth

import (
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
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
	SessionID        *string    `json:"-"`
	CreatedAt        *time.Time `json:"-"`
	SessionExpiresAt *time.Time `json:"-"`
	UserID           *string    `json:"-"`
	Name             *string    `json:"user_name"`
	Username         *string    `json:"username"`
	jwt.Claims
}
