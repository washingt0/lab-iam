package middleware

import (
	"crypto/rsa"
	"lab/iam/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Session represents a session rebuilt from a JWT
type Session struct {
	ID               *string    `json:"session_id"`
	CreatedAt        *time.Time `json:"created_at"`
	SessionExpiresAt *time.Time `json:"expires_at"`
	UserID           *string    `json:"user_id"`
	Name             *string    `json:"user_name"`
	Username         *string    `json:"username"`
	jwt.Claims
}

// ValidateJWT check if the request contains a valid JWT token
func ValidateJWT() gin.HandlerFunc {
	jwtConfig := config.GetConfig().JWT
	return func(c *gin.Context) {
		var (
			sess  *Session = new(Session)
			raw   string
			token *jwt.JSONWebToken
			err   error
		)

		if raw = c.GetHeader("Authorization"); raw == "" || len(raw) < 10 {
			oops.GinHandleError(c, oops.ThrowError("Invalid request", nil), http.StatusUnauthorized)
			return
		}

		if token, err = jwt.ParseSigned(raw[7:]); err != nil {
			oops.GinHandleError(c, oops.ThrowError("Invalid JWT token", err), http.StatusUnauthorized)
			return
		}

		if err = token.Claims(getKey(token.Headers, &jwtConfig), sess); err != nil {
			oops.GinHandleError(c, oops.ThrowError("Invalid JWT signature", err), http.StatusUnauthorized)
			return
		}

		c.Set("USession", sess)
		c.Set("UID", sess.Subject)

		c.Next()
	}
}

func getKey(headers []jose.Header, jwt *config.JWTConfig) (key *rsa.PublicKey) {
	for i := range headers {
		if headers[i].KeyID != "" {
			for j := range jwt.Keys {
				if headers[i].KeyID == jwt.Keys[j].ID {
					return jwt.Keys[j].PublicKey
				}
			}
		}
	}

	zap.L().Debug("NO KEY FOUND", zap.Any("JOSE-HEADERS", headers), zap.Any("JWT-CONFIG", jwt))

	return
}
