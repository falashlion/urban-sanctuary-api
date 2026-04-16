package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Recovery creates a panic recovery middleware.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log full stack trace
				log.Error().
					Str("request_id", c.GetString("request_id")).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("stack", string(debug.Stack())).
					Str("error", fmt.Sprintf("%v", err)).
					Msg("panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success":    false,
					"status":     500,
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					},
					"request_id": c.GetString("request_id"),
				})
			}
		}()
		c.Next()
	}
}
