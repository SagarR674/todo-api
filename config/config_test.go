package config

import (
	"testing"
	"time"
)

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("X_STR", "hello")
	if got := getEnv("X_STR", "fallback"); got != "hello" {
		t.Errorf("getEnv = %q", got)
	}
	if got := getEnv("X_MISSING", "fallback"); got != "fallback" {
		t.Errorf("getEnv fallback = %q", got)
	}

	t.Setenv("X_INT", "42")
	if got := getEnvInt("X_INT", 0); got != 42 {
		t.Errorf("getEnvInt = %d", got)
	}
	t.Setenv("X_INT_BAD", "notint")
	if got := getEnvInt("X_INT_BAD", 7); got != 7 {
		t.Errorf("getEnvInt bad value should fall back, got %d", got)
	}

	t.Setenv("X_BOOL", "true")
	if !getEnvBool("X_BOOL", false) {
		t.Error("getEnvBool should parse true")
	}
	if getEnvBool("X_BOOL_MISSING", false) {
		t.Error("getEnvBool missing should be fallback")
	}

	t.Setenv("X_DUR", "90s")
	if got := getEnvDuration("X_DUR", time.Minute); got != 90*time.Second {
		t.Errorf("getEnvDuration = %v", got)
	}
	t.Setenv("X_DUR_BAD", "soon")
	if got := getEnvDuration("X_DUR_BAD", time.Minute); got != time.Minute {
		t.Errorf("getEnvDuration bad value should fall back, got %v", got)
	}
}

func TestConfig_Validate(t *testing.T) {
	base := Config{JWTSecret: "0123456789abcdef0", DBName: "todo_db"}
	if err := base.validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	missing := base
	missing.JWTSecret = ""
	if err := missing.validate(); err == nil {
		t.Error("empty JWT_SECRET should be rejected")
	}

	short := base
	short.JWTSecret = "tooshort"
	if err := short.validate(); err == nil {
		t.Error("short JWT_SECRET should be rejected")
	}

	noDB := base
	noDB.DBName = ""
	if err := noDB.validate(); err == nil {
		t.Error("empty DB_NAME should be rejected")
	}
}

func TestConfig_DSN(t *testing.T) {
	c := Config{
		DBUser: "root", DBPassword: "p@ss", DBHost: "localhost",
		DBPort: "3306", DBName: "todo_db",
	}
	want := "root:p@ss@tcp(localhost:3306)/todo_db?charset=utf8mb4&parseTime=True&loc=Local"
	if got := c.DSN(); got != want {
		t.Errorf("DSN = %q", got)
	}
	if got := c.MigrationDSN(); got != want+"&multiStatements=true" {
		t.Errorf("MigrationDSN = %q", got)
	}
	if got := c.RootDSN(); got != "root:p@ss@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Errorf("RootDSN = %q", got)
	}
}

func TestConfig_IsProduction(t *testing.T) {
	if (&Config{AppEnv: "development"}).IsProduction() {
		t.Error("development is not production")
	}
	for _, env := range []string{"production", "prod"} {
		if !(&Config{AppEnv: env}).IsProduction() {
			t.Errorf("%q should be production", env)
		}
	}
}

func TestLoad_DefaultsAndValidation(t *testing.T) {
	// Isolate from any real .env by pointing the loader at an empty dir.
	dir := t.TempDir()
	t.Chdir(dir)

	// Missing JWT_SECRET -> error.
	if _, err := Load(); err == nil {
		t.Fatal("Load should fail without JWT_SECRET")
	}

	t.Setenv("JWT_SECRET", "0123456789abcdef0123")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8080" || cfg.DBName != "todo_db" || cfg.JWTExpiry != 24*time.Hour {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
	if cfg.RateLimitMax != 100 || cfg.AuthRateLimitMax != 10 {
		t.Errorf("unexpected rate-limit defaults: %+v", cfg)
	}
}
