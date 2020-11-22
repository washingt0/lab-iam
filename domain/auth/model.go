package auth

import "time"

type IAuth interface {
	GetCredentials(username *string) (string, error)
	CreateSession(username, userAgent, loginIP, loginLocation *string) (*Session, error)
	DropSession(sessionID *string) error
}

type Session struct {
	ID        *string
	CreatedAt *time.Time
	ExpiresAt *time.Time
	UserID    *string
	Name      *string
	Username  *string
}
