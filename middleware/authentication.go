package middleware

import (
	"crypto/rsa"
	"lab/iam/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"

	jwt "github.com/dgrijalva/jwt-go"
)

// Session represents a session rebuilt from a JWT
type Session struct {
	ID               *string    `json:"session_id"`
	CreatedAt        *time.Time `json:"created_at"`
	SessionExpiresAt *time.Time `json:"expires_at"`
	UserID           *string    `json:"user_id"`
	Name             *string    `json:"user_name"`
	Username         *string    `json:"username"`
	jwt.StandardClaims
}

// ValidateJWT check if the request contains a valid JWT token
func ValidateJWT() gin.HandlerFunc {
	keys := config.GetConfig().PublicKeys
	return func(c *gin.Context) {
		var (
			sess  *Session
			token string
			err   error
		)

		if token = c.GetHeader("Authorization"); token == "" || len(token) < 10 {
			oops.GinHandleError(c, oops.ThrowError("Invalid request", nil), http.StatusUnauthorized)
			return
		}

		if sess, err = decodeJWT(token[7:], keys); err != nil {
			oops.GinHandleError(c, err, http.StatusUnauthorized)
			return
		}

		c.Set("USession", sess)
		c.Set("UID", sess.Subject)

		c.Next()
	}
}

func decodeJWT(token string, keys map[string]*rsa.PublicKey) (sess *Session, err error) {
	if len(keys) == 0 {
		return nil, oops.ThrowError("no RSA keys was supplied", err)
	}

	for i := range keys {
		if sess, err = tryDecodeJWT(token, keys[i]); err == nil {
			break
		}
	}

	if sess == nil || sess.Subject == "" {
		return nil, oops.ThrowError("Invalid/Empty token", nil)
	}

	return
}

func tryDecodeJWT(token string, key *rsa.PublicKey) (sess *Session, err error) {
	var (
		decoded *jwt.Token
	)
	sess = new(Session)

	if decoded, err = jwt.ParseWithClaims(token, sess, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, oops.ThrowError("unexpected signing method: "+token.Header["alg"].(string), nil)
		}

		return key, nil
	}); err != nil {
		return nil, err
	}

	if decoded == nil {
		return nil, oops.ThrowError("was not possible to decode the token", nil)
	}

	if claims, ok := decoded.Claims.(*Session); ok && decoded.Valid {
		if claims.Issuer != config.GetConfig().JWT.Issuer {
			return nil, oops.ThrowError("token issuer is not valid", nil)
		}
		return claims, nil
	}

	return nil, oops.ThrowError("token content is not valid", nil)
}
