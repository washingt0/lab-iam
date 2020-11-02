package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"
	"go.uber.org/zap"
)

// RequestLogger logs all requests and its data
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			initialTime time.Time = time.Now()
		)

		c.Next()

		var (
			errors []error = oops.GetGinError(c)
			fields         = []zap.Field{
				zap.Int("status_code", c.Writer.Status()),
				zap.String("client_ip", c.ClientIP()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Float64("latency", float64(time.Since(initialTime))/float64(time.Millisecond)),
				zap.String("client_user_agent", c.Request.UserAgent()),
				zap.String("log_type", "access"),
				zap.String("request_id", c.Value("RID").(string)),
				zap.String("query_params", c.Request.URL.Query().Encode()),
			}
		)

		if len(errors) > 0 {
			zap.L().Error("request handling failed", append(fields, zap.Errors("errors", errors))...)
		} else {
			zap.L().Info("request handled", fields...)
		}

	}
}
