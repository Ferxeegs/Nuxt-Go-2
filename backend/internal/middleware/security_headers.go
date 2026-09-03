package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

// SecurityHeaders sets the baseline hardening headers recommended for any
// production API: nosniff, frame denial, HSTS, and a conservative CSP.
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	sec := secure.New(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		SSLRedirect:           false,
		IsDevelopment:         !isProduction,
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
	})

	return func(c *gin.Context) {
		// Uploaded files are embedded in the admin UI (PDF preview). Allow
		// same-origin framing; API JSON routes stay frame-denied.
		if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
			c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
			c.Next()
			return
		}

		err := sec.Process(c.Writer, c.Request)
		if err != nil {
			c.Abort()
			return
		}
		c.Next()
	}
}
