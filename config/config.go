// Package config loads application configuration from environment variables.
//
// All configuration comes from the environment (optionally seeded from a .env
// file in development). Nothing sensitive is hard-coded.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the application.
type Config struct {
	AppEnv string
	Port   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string
	JWTExpiry time.Duration

	RateLimitMax        int
	RateLimitWindow     time.Duration
	AuthRateLimitMax    int
	AuthRateLimitWindow time.Duration

	CORSOrigins string

	AutoMigrate bool

	LogLevel string
}

// Load reads configuration from the environment. In development it first tries
// to load a .env file from the working directory; a missing .env is not fatal
// because real deployments inject env vars directly.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "root"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "todo_db"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	cfg.JWTExpiry = getEnvDuration("JWT_EXPIRY", 24*time.Hour)
	cfg.RateLimitMax = getEnvInt("RATE_LIMIT_MAX", 100)
	cfg.RateLimitWindow = getEnvDuration("RATE_LIMIT_WINDOW", time.Minute)
	cfg.AuthRateLimitMax = getEnvInt("AUTH_RATE_LIMIT_MAX", 10)
	cfg.AuthRateLimitWindow = getEnvDuration("AUTH_RATE_LIMIT_WINDOW", time.Minute)
	cfg.AutoMigrate = getEnvBool("AUTO_MIGRATE", false)

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate fails fast when required secrets are missing.
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must be at least 16 characters")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

// IsProduction reports whether the app is running in a production environment.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production" || c.AppEnv == "prod"
}

// DSN returns the MySQL data source name for the configured database.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// MigrationDSN is the DSN used by the migration runner. It enables
// multiStatements so a single migration file may contain several statements.
func (c *Config) MigrationDSN() string {
	return c.DSN() + "&multiStatements=true"
}

// RootDSN returns a MySQL DSN without a database selected. It is used once at
// startup to create the database if it does not yet exist (dev convenience).
func (c *Config) RootDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
