// Package main is the application entry point: loads config, connects
// infra, wires dependencies, and runs the HTTP server with graceful
// shutdown.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	_ "go-cp-boilerplate/docs"
	"go-cp-boilerplate/internal/cache"
	"go-cp-boilerplate/internal/config"
	"go-cp-boilerplate/internal/database"
	"go-cp-boilerplate/internal/pkg/jwtutil"
	appLogger "go-cp-boilerplate/internal/pkg/logger"
	"go-cp-boilerplate/internal/pkg/turnstile"
	"go-cp-boilerplate/internal/pkg/uploader"
	"go-cp-boilerplate/internal/router"

	"go-cp-boilerplate/internal/modules/activitylog"
	authmod "go-cp-boilerplate/internal/modules/auth"
	"go-cp-boilerplate/internal/modules/media"
	"go-cp-boilerplate/internal/modules/role"
	"go-cp-boilerplate/internal/modules/scope"
	"go-cp-boilerplate/internal/modules/user"
	"go-cp-boilerplate/internal/modules/websetting"
)

// @title                       Go-Gin Company Profile Boilerplate API
// @version                     1.0
// @description                 Headless backend API for company profile / landing page projects.
// @BasePath                    /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	logger := appLogger.Init(cfg.LogLevel, cfg.IsProduction())
	log.Logger = logger

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	if err := database.AutoMigrate(db, cfg.IsProduction()); err != nil {
		log.Fatal().Err(err).Msg("failed to run auto migration")
	}

	if cfg.DBSeed {
		log.Info().Msg("running database seed")
		if err := database.Seed(db, cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to seed database")
		}
	} else {
		log.Info().Msg("database seed disabled (DB_SEED=false)")
	}

	redisClient := cache.NewRedisClient(cfg)
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Warn().Err(err).Msg("redis is not reachable at startup - rate limiting/blacklist will fail open")
	}

	jwtManager := jwtutil.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	fileUploader, err := uploader.NewFromConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize file uploader")
	}
	if fileUploader.UsesS3() {
		log.Info().Str("public_prefix", fileUploader.PublicURLPrefix()).Msg("media uploads use RustFS/S3")
	} else {
		log.Info().Str("dir", cfg.UploadDir).Msg("media uploads use local disk")
	}
	turnstileVerifier := turnstile.New(cfg.TurnstileSecretKey)
	if !turnstileVerifier.Enabled() {
		log.Warn().Msg("TURNSTILE_SECRET_KEY is empty - login captcha verification is disabled")
	}

	handlers := buildHandlers(db, cfg, jwtManager, redisClient, fileUploader, turnstileVerifier)

	engine := router.New(cfg, logger, handlers, jwtManager, redisClient, &database.Pinger{DB: db})

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: engine,
	}

	go func() {
		log.Info().Str("port", cfg.AppPort).Str("env", cfg.AppEnv).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	gracefulShutdown(srv, redisClient)
}

func buildHandlers(
	db *gorm.DB,
	cfg *config.Config,
	jwtManager *jwtutil.Manager,
	redisClient *cache.RedisClient,
	fileUploader *uploader.Uploader,
	turnstileVerifier turnstile.Verifier,
) router.Handlers {
	authRepo := authmod.NewRepository(db)
	authService := authmod.NewService(authRepo, jwtManager, redisClient, cfg, cfg.BcryptCost, turnstileVerifier)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	roleRepo := role.NewRepository(db)
	roleService := role.NewService(roleRepo)

	scopeRepo := scope.NewRepository(db)
	scopeService := scope.NewService(scopeRepo)

	settingRepo := websetting.NewRepository(db)
	settingService := websetting.NewService(settingRepo)

	mediaRepo := media.NewRepository(db)
	mediaService := media.NewService(mediaRepo, fileUploader, fileUploader.PublicURLPrefix())

	activityLogRepo := activitylog.NewRepository(db)
	activityLogService := activitylog.NewService(activityLogRepo)

	return router.Handlers{
		Auth:        authmod.NewHandler(authService),
		User:        user.NewHandler(userService, cfg.BcryptCost, cfg.DefaultPassword),
		Role:        role.NewHandler(roleService),
		Scope:       scope.NewHandler(scopeService),
		WebSetting:  websetting.NewHandler(settingService),
		Media:       media.NewHandler(mediaService),
		ActivityLog: activitylog.NewHandler(activityLogService),
	}
}

// gracefulShutdown waits for SIGINT/SIGTERM so in-flight requests finish
// before the process exits (important on `docker stop` / redeploy).
func gracefulShutdown(srv *http.Server, redisClient *cache.RedisClient) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	if redisClient != nil {
		_ = redisClient.Close()
	}

	log.Info().Msg("server exited gracefully")
}
