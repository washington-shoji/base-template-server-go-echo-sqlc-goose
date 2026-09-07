package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Logging   LoggingConfig
	Metrics   MetricsConfig
	Auth      AuthConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Jobs      JobsConfig
	Reports   ReportsConfig
	BodyLimit string
}

type HTTPConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationsDir   string
}

func (dc DatabaseConfig) ToDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode)
}

type LoggingConfig struct {
	Level string
}

type MetricsConfig struct {
	Enabled       bool
	BasicAuthUser string
	BasicAuthPass string
	// Public allows unauthenticated /metrics (dev only).
	Public bool
}

type AuthConfig struct {
	Disabled       bool
	SessionCookie  string
	SessionTTL     time.Duration
	CSRFCookieName string
	TokenTTL       time.Duration
	BcryptCost     int
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	// Requests is tokens per second (golang.org/x/time/rate).
	Requests      float64
	Burst         int
	UseMemory     bool // process-local; unsuitable alone for multi-instance prod
	PostgresStore bool
}

type JobsConfig struct {
	Enabled       bool
	PollInterval  time.Duration
	MaxAttempts   int
	EmbedInServer bool
}

// ReportsConfig controls async report lifecycle lease and reconciliation.
type ReportsConfig struct {
	ProcessingLease   time.Duration // stale processing reclaim window
	ReconcileAfter    time.Duration // pending without progress → re-enqueue
	ReconcileInterval time.Duration // ticker period for ReconcilePending
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv: getenv("APP_ENV", "development"),
		HTTP: HTTPConfig{
			Port:            getenv("PORT", "8080"),
			ReadTimeout:     durationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    durationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: durationEnv("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
			TrustedProxies:  splitCSV(os.Getenv("TRUSTED_PROXIES")),
		},
		Database: DatabaseConfig{
			Host:            getenv("DB_HOST", "localhost"),
			Port:            intEnv("DB_PORT", 5432),
			User:            getenv("DB_USER", "postgres"),
			Password:        getenv("DB_PASSWORD", "password"),
			DBName:          getenv("DB_NAME", "app_db"),
			SSLMode:         getenv("DB_SSLMODE", "disable"),
			MaxOpenConns:    intEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    intEnv("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: durationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			MigrationsDir:   getenv("DB_MIGRATIONS_DIR", "sql/migrations"),
		},
		Logging: LoggingConfig{Level: getenv("LOG_LEVEL", "")},
		Metrics: MetricsConfig{
			Enabled:       boolEnv("METRICS_ENABLED", true),
			BasicAuthUser: os.Getenv("METRICS_BASIC_AUTH_USER"),
			BasicAuthPass: os.Getenv("METRICS_BASIC_AUTH_PASS"),
			Public:        boolEnv("METRICS_PUBLIC", false),
		},
		Auth: AuthConfig{
			Disabled:       boolEnv("AUTH_DISABLED", true),
			SessionCookie:  getenv("AUTH_SESSION_COOKIE", "session_id"),
			SessionTTL:     durationEnv("AUTH_SESSION_TTL", 24*time.Hour),
			CSRFCookieName: getenv("AUTH_CSRF_COOKIE", "csrf_token"),
			TokenTTL:       durationEnv("AUTH_TOKEN_TTL", 24*time.Hour),
			BcryptCost:     intEnv("AUTH_BCRYPT_COST", 10),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		RateLimit: RateLimitConfig{
			Requests:      floatEnv("RATE_LIMIT_REQUESTS", 100),
			Burst:         intEnv("RATE_LIMIT_BURST", 50),
			UseMemory:     boolEnv("RATE_LIMIT_USE_MEMORY", true),
			PostgresStore: boolEnv("RATE_LIMIT_POSTGRES", false),
		},
		Jobs: JobsConfig{
			Enabled:       boolEnv("JOBS_ENABLED", true),
			PollInterval:  durationEnv("JOBS_POLL_INTERVAL", 2*time.Second),
			MaxAttempts:   intEnv("JOBS_MAX_ATTEMPTS", 5),
			EmbedInServer: boolEnv("JOBS_EMBED_IN_SERVER", true),
		},
		Reports: ReportsConfig{
			ProcessingLease:   durationEnv("REPORTS_PROCESSING_LEASE", 5*time.Minute),
			ReconcileAfter:    durationEnv("REPORTS_RECONCILE_AFTER", time.Minute),
			ReconcileInterval: durationEnv("REPORTS_RECONCILE_INTERVAL", 30*time.Second),
		},
		BodyLimit: getenv("BODY_LIMIT", "2M"),
	}

	if cfg.Logging.Level == "" {
		switch cfg.AppEnv {
		case "production":
			cfg.Logging.Level = "INFO"
		default:
			cfg.Logging.Level = "DEBUG"
		}
	}

	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.CORS.AllowedOrigins = splitCSV(origins)
	}

	// Dev convenience: public metrics if unset and not production.
	if os.Getenv("METRICS_PUBLIC") == "" && cfg.AppEnv != "production" {
		cfg.Metrics.Public = true
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.HTTP.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.Database.Host == "" || c.Database.DBName == "" {
		return fmt.Errorf("database host and name are required")
	}
	if c.AppEnv == "production" {
		if c.Metrics.Public && c.Metrics.BasicAuthUser == "" {
			return fmt.Errorf("production requires METRICS_BASIC_AUTH_USER/PASS or METRICS_PUBLIC=false")
		}
		if c.Auth.Disabled {
			return fmt.Errorf("AUTH_DISABLED cannot be true in production")
		}
	}
	return nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func intEnv(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func floatEnv(k string, def float64) float64 {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func boolEnv(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func durationEnv(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
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
