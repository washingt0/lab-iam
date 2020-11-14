package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIdentifier add a uuid to every request
func RequestIdentifier() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RID", uuid.New().String())
	}
}
