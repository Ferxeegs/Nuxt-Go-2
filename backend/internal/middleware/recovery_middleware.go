package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"go-cp-boilerplate/internal/pkg/response"
)

// Recovery catches panics, logs the stack trace and returns a standardized
// 500 response without ever leaking the stack trace to the client.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get(RequestIDContextKey)
				log.Error().
					Any("request_id", requestID).
					Any("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("panic recovered")

				response.Error(c, http.StatusInternalServerError, "Internal server error", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
