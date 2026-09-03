package router

import (
	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/constants"
	"go-cp-boilerplate/internal/middleware"
)

// RegisterAdminRoutes mounts every /admin/* route behind Auth + a
// per-endpoint RequirePermission check.
func RegisterAdminRoutes(api *gin.RouterGroup, h Handlers, authMiddleware gin.HandlerFunc) {
	admin := api.Group("/admin", authMiddleware)

	users := admin.Group("/users")
	{
		users.GET("", middleware.RequirePermission(constants.PermUserView), h.User.List)
		users.GET("/:id", middleware.RequirePermission(constants.PermUserView), h.User.Get)
		users.POST("", middleware.RequirePermission(constants.PermUserCreate), h.User.Create)
		users.PUT("/:id", middleware.RequirePermission(constants.PermUserUpdate), h.User.Update)
		users.POST("/:id/reset-password", middleware.RequirePermission(constants.PermUserUpdate), h.User.ResetPassword)
		users.DELETE("/:id", middleware.RequirePermission(constants.PermUserDelete), h.User.Delete)
	}

	roles := admin.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission(constants.PermRoleView), h.Role.List)
		roles.GET("/permissions", middleware.RequirePermission(constants.PermRoleView), h.Role.ListPermissions)
		roles.GET("/:id", middleware.RequirePermission(constants.PermRoleView), h.Role.Get)
		roles.POST("", middleware.RequirePermission(constants.PermRoleCreate), h.Role.Create)
		roles.PUT("/:id", middleware.RequirePermission(constants.PermRoleUpdate), h.Role.Update)
		roles.DELETE("/:id", middleware.RequirePermission(constants.PermRoleDelete), h.Role.Delete)
	}

	scopes := admin.Group("/scopes")
	{
		scopes.GET("", middleware.RequirePermission(constants.PermScopeView), h.Scope.List)
		scopes.GET("/:id", middleware.RequirePermission(constants.PermScopeView), h.Scope.Get)
		scopes.POST("", middleware.RequirePermission(constants.PermScopeCreate), h.Scope.Create)
		scopes.PUT("/:id", middleware.RequirePermission(constants.PermScopeUpdate), h.Scope.Update)
		scopes.DELETE("/:id", middleware.RequirePermission(constants.PermScopeDelete), h.Scope.Delete)
	}

	settings := admin.Group("/settings")
	{
		settings.GET("", middleware.RequirePermission(constants.PermSettingView), h.WebSetting.ListAdmin)
		settings.PUT("", middleware.RequirePermission(constants.PermSettingUpdate), h.WebSetting.Update)
	}

	media := admin.Group("/media")
	{
		media.GET("", middleware.RequirePermission(constants.PermMediaUpload), h.Media.List)
		media.POST("/upload", middleware.RequirePermission(constants.PermMediaUpload), h.Media.Upload)
		media.DELETE("/:id", middleware.RequirePermission(constants.PermMediaDelete), h.Media.Delete)
	}

	activityLogs := admin.Group("/activity-logs")
	{
		activityLogs.GET("", middleware.RequirePermission(constants.PermActivityView), h.ActivityLog.List)
		activityLogs.GET("/filters", middleware.RequirePermission(constants.PermActivityView), h.ActivityLog.ListFilters)
		activityLogs.GET("/:id", middleware.RequirePermission(constants.PermActivityView), h.ActivityLog.Get)
		activityLogs.DELETE("/:id", middleware.RequirePermission(constants.PermActivityDelete), h.ActivityLog.Delete)
	}
}
