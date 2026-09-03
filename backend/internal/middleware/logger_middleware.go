package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger emits one structured JSON log line per request: timestamp, method,
// path, status_code, latency_ms, client_ip, request_id.
func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		requestID, _ := c.Get(RequestIDContextKey)

		event := log.Info()
		if c.Writer.Status() >= 500 {
			event = log.Error()
		} else if c.Writer.Status() >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status_code", c.Writer.Status()).
			Float64("latency_ms", float64(latency.Microseconds())/1000.0).
			Str("client_ip", c.ClientIP()).
			Any("request_id", requestID).
			Msg("request handled")
	}
}
