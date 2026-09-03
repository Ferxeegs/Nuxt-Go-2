package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/response"
)

// RequirePermission checks that the authenticated user's permission list
// (embedded in the JWT by Auth middleware) contains the given permission
// key, e.g. RequirePermission("user.delete"). SuperAdmin always passes.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, ok := c.Get(ContextRole); ok {
			if roleName, ok := role.(string); ok && roleName == domain.RoleSuperAdmin {
				c.Next()
				return
			}
		}

		raw, exists := c.Get(ContextPermissions)
		if !exists {
			response.Error(c, http.StatusForbidden, "Access denied", nil)
			c.Abort()
			return
		}

		permissions, ok := raw.([]string)
		if !ok {
			response.Error(c, http.StatusForbidden, "Access denied", nil)
			c.Abort()
			return
		}

		for _, p := range permissions {
			if p == permission {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "You do not have permission to perform this action", nil)
		c.Abort()
	}
}

// RequireRole restricts an endpoint to one or more specific role names
// (e.g. RequireRole(domain.RoleSuperAdmin)) when a permission check alone
// isn't granular enough.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists || !allowed[role.(string)] {
			response.Error(c, http.StatusForbidden, "You do not have permission to perform this action", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
