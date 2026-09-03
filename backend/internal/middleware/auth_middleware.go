package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/cache"
	"go-cp-boilerplate/internal/pkg/jwtutil"
	"go-cp-boilerplate/internal/pkg/response"
)

const (
	ContextUserID         = "user_id"
	ContextRole           = "role"
	ContextPermissions    = "permissions"
	ContextTokenJTI       = "token_jti"
	ContextTokenExpiry    = "token_expiry"
	ContextImpersonatorID = "impersonator_id"
	ContextActiveRoleID   = "active_role_id"
	ContextActiveScopeID  = "active_scope_id"
)

// Auth validates the Bearer JWT (signature, expiry, blacklist) and injects
// user_id/role/permissions into the Gin context for downstream handlers and
// RBAC middleware.
func Auth(jwtManager *jwtutil.Manager, redisClient *cache.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "Missing or invalid Authorization header", nil)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtManager.ParseAccessToken(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token", nil)
			c.Abort()
			return
		}

		if redisClient != nil {
			blacklisted, err := redisClient.IsBlacklisted(c.Request.Context(), claims.ID)
			if err == nil && blacklisted {
				response.Error(c, http.StatusUnauthorized, "Token has been revoked", nil)
				c.Abort()
				return
			}
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Set(ContextPermissions, claims.Permissions)
		c.Set(ContextTokenJTI, claims.ID)
		c.Set(ContextActiveRoleID, claims.ActiveRoleID)
		c.Set(ContextActiveScopeID, claims.ActiveScopeID)
		if claims.ExpiresAt != nil {
			c.Set(ContextTokenExpiry, claims.ExpiresAt.Time)
		}
		if claims.ImpersonatorID != nil {
			c.Set(ContextImpersonatorID, *claims.ImpersonatorID)
		}
		c.Next()
	}
}
