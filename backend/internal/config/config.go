package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all environment-driven configuration for the application.
type Config struct {
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisPrefix   string

	JWTSecret          string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	BcryptCost         int
	SuperAdminEmail    string
	SuperAdminPassword string
	// DefaultPassword is applied when an admin resets a user's password.
	DefaultPassword string

	AllowedOrigins []string

	RateLimitPublicPerMinute int
	RateLimitLoginPerMinute  int

	MaxUploadSizeMB int64
	UploadDir       string

	// Optional RustFS / S3 settings (local disk uploader still used until wired).
	S3EndpointURL       string
	S3PublicEndpointURL string
	S3AccessKey         string
	S3SecretKey         string
	S3Region            string
	S3Bucket            string
	S3PublicBucket      string
	S3UseSSL            bool
	S3ForcePathStyle    bool

	LogLevel string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	NotifyEmail  string

	// TurnstileSecretKey verifies Cloudflare Turnstile tokens on login (server-only).
	// Required in production; optional in development (verification skipped when empty).
	TurnstileSecretKey string

	// DBSeed runs idempotent startup seeding (roles, permissions, settings, etc.).
	DBSeed bool
}

// Load reads configuration from a .env file (if present) and environment
// variables, validating that required values are set. It fails fast when a
// required variable is missing so misconfiguration is caught at startup
// rather than mid-request. Docker Compose injects env from the repo-root
// `.env`; the `../.env` fallback covers `go run` from backend/ on the host.
func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	v := viper.New()
	v.SetConfigType("env")
	v.AutomaticEnv()

	setDefaults(v)

	cfg := &Config{
		AppEnv:  v.GetString("APP_ENV"),
		AppPort: v.GetString("APP_PORT"),

		DBHost:     v.GetString("DB_HOST"),
		DBPort:     v.GetString("DB_PORT"),
		DBUser:     v.GetString("DB_USER"),
		DBPassword: v.GetString("DB_PASSWORD"),
		DBName:     v.GetString("DB_NAME"),

		DBMaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),

		RedisHost:     v.GetString("REDIS_HOST"),
		RedisPort:     v.GetString("REDIS_PORT"),
		RedisPassword: v.GetString("REDIS_PASSWORD"),
		RedisDB:       v.GetInt("REDIS_DB"),
		RedisPrefix:   v.GetString("REDIS_PREFIX"),

		JWTSecret:          v.GetString("JWT_SECRET"),
		JWTAccessTTL:       v.GetDuration("JWT_ACCESS_TTL"),
		JWTRefreshTTL:      v.GetDuration("JWT_REFRESH_TTL"),
		BcryptCost:         v.GetInt("BCRYPT_COST"),
		SuperAdminEmail:    v.GetString("SUPERADMIN_EMAIL"),
		SuperAdminPassword: v.GetString("SUPERADMIN_PASSWORD"),
		DefaultPassword:    v.GetString("DEFAULT_PASSWORD"),

		AllowedOrigins: splitAndTrim(v.GetString("ALLOWED_ORIGINS")),

		RateLimitPublicPerMinute: v.GetInt("RATE_LIMIT_PUBLIC_PER_MINUTE"),
		RateLimitLoginPerMinute:  v.GetInt("RATE_LIMIT_LOGIN_PER_MINUTE"),

		MaxUploadSizeMB: v.GetInt64("MAX_UPLOAD_SIZE_MB"),
		UploadDir:       v.GetString("UPLOAD_DIR"),

		S3EndpointURL:       v.GetString("S3_ENDPOINT_URL"),
		S3PublicEndpointURL: v.GetString("S3_PUBLIC_ENDPOINT_URL"),
		S3AccessKey:         v.GetString("S3_ACCESS_KEY"),
		S3SecretKey:         v.GetString("S3_SECRET_KEY"),
		S3Region:            v.GetString("S3_REGION"),
		S3Bucket:            v.GetString("S3_BUCKET"),
		S3PublicBucket:      v.GetString("S3_PUBLIC_BUCKET"),
		S3UseSSL:            v.GetBool("S3_USE_SSL"),
		S3ForcePathStyle:    v.GetBool("S3_FORCE_PATH_STYLE"),

		LogLevel: v.GetString("LOG_LEVEL"),

		SMTPHost:     v.GetString("SMTP_HOST"),
		SMTPPort:     v.GetString("SMTP_PORT"),
		SMTPUser:     v.GetString("SMTP_USER"),
		SMTPPassword: v.GetString("SMTP_PASSWORD"),
		SMTPFrom:     v.GetString("SMTP_FROM"),
		NotifyEmail:  v.GetString("NOTIFY_EMAIL"),

		TurnstileSecretKey: v.GetString("TURNSTILE_SECRET_KEY"),
	}

	cfg.DBSeed = envBoolDefault(v, "DB_SEED", cfg.AppEnv != "production")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("DB_PORT", "3306")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_PREFIX", "cp_")
	v.SetDefault("JWT_ACCESS_TTL", "30m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("ALLOWED_ORIGINS", "")
	v.SetDefault("RATE_LIMIT_PUBLIC_PER_MINUTE", 100)
	v.SetDefault("RATE_LIMIT_LOGIN_PER_MINUTE", 5)
	v.SetDefault("MAX_UPLOAD_SIZE_MB", 10)
	v.SetDefault("UPLOAD_DIR", "./storage/uploads")
	v.SetDefault("S3_REGION", "us-east-1")
	v.SetDefault("S3_USE_SSL", false)
	v.SetDefault("S3_FORCE_PATH_STYLE", true)
	v.SetDefault("LOG_LEVEL", "info")
}

func (c *Config) validate() error {
	required := map[string]string{
		"DB_HOST":             c.DBHost,
		"DB_USER":             c.DBUser,
		"DB_NAME":             c.DBName,
		"JWT_SECRET":          c.JWTSecret,
		"SUPERADMIN_EMAIL":    c.SuperAdminEmail,
		"SUPERADMIN_PASSWORD": c.SuperAdminPassword,
		"DEFAULT_PASSWORD":    c.DefaultPassword,
	}
	var missing []string
	for name, val := range required {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, name)
		}
	}
	if c.IsProduction() && strings.TrimSpace(c.TurnstileSecretKey) == "" {
		missing = append(missing, "TURNSTILE_SECRET_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	if len(strings.TrimSpace(c.DefaultPassword)) < 6 {
		return fmt.Errorf("DEFAULT_PASSWORD must be at least 6 characters")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// envBoolDefault reads a boolean env var; when unset, returns fallbackDefault.
func envBoolDefault(v *viper.Viper, key string, fallbackDefault bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	return fallbackDefault
}

func splitAndTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
