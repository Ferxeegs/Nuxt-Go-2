// Package router is the composition root: it wires global middleware and
// mounts the public and admin route groups.
package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-cp-boilerplate/internal/cache"
	"go-cp-boilerplate/internal/config"
	"go-cp-boilerplate/internal/middleware"
	"go-cp-boilerplate/internal/modules/activitylog"
	authmod "go-cp-boilerplate/internal/modules/auth"
	"go-cp-boilerplate/internal/modules/media"
	"go-cp-boilerplate/internal/modules/role"
	"go-cp-boilerplate/internal/modules/scope"
	"go-cp-boilerplate/internal/modules/user"
	"go-cp-boilerplate/internal/modules/websetting"
	"go-cp-boilerplate/internal/pkg/jwtutil"

	"github.com/rs/zerolog"
)

// Handlers bundles every module's HTTP handler so main.go only needs to
// construct them once and hand them to the router.
type Handlers struct {
	Auth        *authmod.Handler
	User        *user.Handler
	Role        *role.Handler
	Scope       *scope.Handler
	WebSetting  *websetting.Handler
	Media       *media.Handler
	ActivityLog *activitylog.Handler
}

type HealthChecker interface {
	Ping() error
}

func New(cfg *config.Config, log zerolog.Logger, h Handlers, jwtManager *jwtutil.Manager, redisClient *cache.RedisClient, dbPing HealthChecker) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.SecurityHeaders(cfg.IsProduction()))
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	r.Static("/uploads", cfg.UploadDir)

	registerHealthRoutes(r, dbPing)

	if !cfg.IsProduction() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	rateLimitPublic := middleware.RateLimit(redisClient, "public", cfg.RateLimitPublicPerMinute, time.Minute)
	rateLimitLogin := middleware.RateLimit(redisClient, "login", cfg.RateLimitLoginPerMinute, time.Minute)
	authMiddleware := middleware.Auth(jwtManager, redisClient)

	api := r.Group("/api/v1")
	{
		RegisterPublicRoutes(api, h, rateLimitPublic, rateLimitLogin, authMiddleware)
		RegisterAdminRoutes(api, h, authMiddleware)
	}

	return r
}

func registerHealthRoutes(r *gin.Engine, dbPing HealthChecker) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if dbPing != nil {
			if err := dbPing.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
