// Package test holds black-box integration tests that exercise the fully wired
// HTTP application against a real MySQL database.
//
// The suite needs its own schema so it can create and truncate tables freely.
// Configuration is taken from TEST_DB_* environment variables, falling back to
// the DB_* values with "_test" appended to the database name.
//
// If the database cannot be reached the suite is skipped, unless REQUIRE_DB=1
// (set in CI) in which case it fails.
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/database"
	"github.com/SagarR674/todo-api/pkg/logger"
	"github.com/SagarR674/todo-api/server"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	testCfg *config.Config
	testDB  *gorm.DB
)

func TestMain(m *testing.M) {
	testCfg = loadTestConfig()
	logger.Init(testCfg.LogLevel)
	utils.BcryptCost = bcrypt.MinCost // keep the suite fast

	if err := database.Migrate(testCfg, database.Up); err != nil {
		skipOrFail(fmt.Sprintf("cannot prepare test database: %v", err))
	}

	db, err := database.Connect(testCfg)
	if err != nil {
		skipOrFail(fmt.Sprintf("cannot connect to test database: %v", err))
	}
	testDB = db

	code := m.Run()

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	os.Exit(code)
}

func skipOrFail(msg string) {
	if os.Getenv("REQUIRE_DB") == "1" {
		fmt.Fprintln(os.Stderr, "integration tests: "+msg)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "integration tests skipped: "+msg)
	os.Exit(0)
}

func loadTestConfig() *config.Config {
	env := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = env("DB_NAME", "todo_db") + "_test"
	}

	return &config.Config{
		AppEnv:              "test",
		Port:                "0",
		DBHost:              env("TEST_DB_HOST", env("DB_HOST", "localhost")),
		DBPort:              env("TEST_DB_PORT", env("DB_PORT", "3306")),
		DBUser:              env("TEST_DB_USER", env("DB_USER", "root")),
		DBPassword:          env("TEST_DB_PASSWORD", os.Getenv("DB_PASSWORD")),
		DBName:              dbName,
		JWTSecret:           "integration-test-secret-key-0123456789",
		JWTExpiry:           24 * time.Hour,
		CORSOrigins:         "*",
		RateLimitMax:        100000,
		RateLimitWindow:     time.Minute,
		AuthRateLimitMax:    100000,
		AuthRateLimitWindow: time.Minute,
		LogLevel:            "error",
	}
}

// ---- per-test API client ----------------------------------------------

type apiClient struct {
	t   *testing.T
	app *fiber.App
}

// newClient truncates every table and returns a client bound to a freshly
// wired app using the shared test config.
func newClient(t *testing.T) *apiClient {
	t.Helper()
	truncateAll(t)
	return &apiClient{t: t, app: server.New(testCfg, testDB)}
}

// newClientWith is like newClient but lets a test tweak the config (e.g. to set
// a low rate limit).
func newClientWith(t *testing.T, mutate func(*config.Config)) *apiClient {
	t.Helper()
	truncateAll(t)
	cfg := *testCfg
	mutate(&cfg)
	return &apiClient{t: t, app: server.New(&cfg, testDB)}
}

func truncateAll(t *testing.T) {
	t.Helper()
	stmts := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE todo_categories",
		"TRUNCATE TABLE categories",
		"TRUNCATE TABLE todos",
		"TRUNCATE TABLE users",
		"SET FOREIGN_KEY_CHECKS = 1",
	}
	for _, s := range stmts {
		if err := testDB.Exec(s).Error; err != nil {
			t.Fatalf("truncate (%s): %v", s, err)
		}
	}
}

// envelope mirrors utils.Response for assertions.
type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Errors  map[string]any  `json:"errors"`
}

type result struct {
	status int
	body   envelope
	raw    []byte
}

func (c *apiClient) do(method, path, token string, body any) result {
	c.t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.app.Test(req, -1)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var env envelope
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &env); err != nil {
			c.t.Fatalf("%s %s: bad JSON %s: %v", method, path, raw, err)
		}
	}
	return result{status: resp.StatusCode, body: env, raw: raw}
}

// doRaw sends an arbitrary request body (used to test malformed JSON).
func (c *apiClient) doRaw(method, path, token, body string) result {
	c.t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.app.Test(req, -1)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var env envelope
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &env)
	}
	return result{status: resp.StatusCode, body: env, raw: raw}
}

func (r result) decodeData(t *testing.T, dst any) {
	t.Helper()
	if err := json.Unmarshal(r.body.Data, dst); err != nil {
		t.Fatalf("decode data %s: %v", r.body.Data, err)
	}
}

// ---- convenience helpers ---------------------------------------------

func (c *apiClient) register(name, email, password string) result {
	return c.do(fiber.MethodPost, "/api/auth/register", "", fiber.Map{
		"name": name, "email": email, "password": password,
	})
}

func (c *apiClient) login(email, password string) result {
	return c.do(fiber.MethodPost, "/api/auth/login", "", fiber.Map{
		"email": email, "password": password,
	})
}

// authUser registers and logs in a user, returning its bearer token.
func (c *apiClient) authUser(email string) string {
	c.t.Helper()
	if res := c.register("User "+email, email, "password123"); res.status != fiber.StatusCreated {
		c.t.Fatalf("register %s: status %d (%s)", email, res.status, res.raw)
	}
	res := c.login(email, "password123")
	if res.status != fiber.StatusOK {
		c.t.Fatalf("login %s: status %d (%s)", email, res.status, res.raw)
	}
	var out struct {
		Token string `json:"token"`
	}
	res.decodeData(c.t, &out)
	if out.Token == "" {
		c.t.Fatalf("login %s: empty token", email)
	}
	return out.Token
}

func (c *apiClient) createTodo(token string, body fiber.Map) result {
	return c.do(fiber.MethodPost, "/api/todos", token, body)
}
