package router

import (
	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/constants"
	"go-cp-boilerplate/internal/middleware"
)

// RegisterPublicRoutes mounts /auth (login/refresh public, logout/me
// authenticated) and /public (no auth required, rate-limited) routes.
func RegisterPublicRoutes(api *gin.RouterGroup, h Handlers, rateLimitPublic, rateLimitLogin, authMiddleware gin.HandlerFunc) {
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", rateLimitLogin, h.Auth.Login)
		authGroup.POST("/refresh", rateLimitPublic, h.Auth.Refresh)
		authGroup.POST("/logout", authMiddleware, h.Auth.Logout)
		authGroup.GET("/me", authMiddleware, h.Auth.Me)
		authGroup.PATCH("/me", authMiddleware, h.Auth.UpdateProfile)
		authGroup.PATCH("/me/password", authMiddleware, h.Auth.ChangePassword)
		authGroup.DELETE("/me", authMiddleware, h.Auth.DeleteProfile)
		authGroup.POST("/impersonate", authMiddleware, middleware.RequirePermission(constants.PermUserImpersonate), h.Auth.Impersonate)
		authGroup.POST("/stop-impersonate", authMiddleware, h.Auth.StopImpersonate)
		authGroup.POST("/switch-role", authMiddleware, h.Auth.SwitchRole)
	}

	public := api.Group("/public", rateLimitPublic)
	{
		public.GET("/settings", h.WebSetting.ListPublic)
	}
}
