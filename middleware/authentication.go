package middleware

import (
	"lab/iam/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"

	jwt "github.com/dgrijalva/jwt-go"
)

// Session represents a session rebuilt from a JWT
type Session struct {
	jwt.StandardClaims
}

// ValidateJWT check if the request contains a valid JWT token
func ValidateJWT() gin.HandlerFunc {
	secrets := config.GetConfig().Secrets
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

		if sess, err = decodeJWT(token[7:], secrets); err != nil {
			oops.GinHandleError(c, err, http.StatusUnauthorized)
			return
		}

		c.Set("USession", sess)
		c.Set("UID", sess.Subject)

		c.Next()
	}
}

func decodeJWT(token string, secrets []string) (sess *Session, err error) {
	if len(secrets) == 0 {
		return nil, oops.ThrowError("no secret was supplied", err)
	}

	for i := range secrets {
		if sess, err = tryDecodeJWT(token, secrets[i]); err == nil {
			break
		}
	}

	if sess == nil || sess.Subject == "" {
		return nil, oops.ThrowError("Invalid/Empty token", nil)
	}

	return
}

func tryDecodeJWT(token string, secret string) (sess *Session, err error) {
	var (
		decoded *jwt.Token
	)
	sess = new(Session)

	if decoded, err = jwt.ParseWithClaims(token, sess, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, oops.ThrowError("unexpected signing method: "+token.Header["alg"].(string), nil)
		}

		return []byte(secret), nil
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
