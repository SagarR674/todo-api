package test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

func TestAuthMiddleware_Rejects(t *testing.T) {
	c := newClient(t)

	cases := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"wrong scheme", "Token abc"},
		{"bearer without token", "Bearer "},
		{"malformed token", "Bearer not.a.jwt"},
		{"garbage", "xxxxx"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/api/todos", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			resp, err := c.app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", resp.StatusCode)
			}
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	c := newClient(t)

	// mint a token that is already expired using the same secret as the app
	jm := utils.NewJWTManager(testCfg.JWTSecret, -time.Hour)
	expired, err := jm.Generate(1)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	req := httptest.NewRequest(fiber.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	resp, _ := c.app.Test(req, -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expired token status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	c := newClient(t)

	jm := utils.NewJWTManager("a-different-secret-not-the-apps-one", time.Hour)
	forged, _ := jm.Generate(1)

	req := httptest.NewRequest(fiber.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+forged)
	resp, _ := c.app.Test(req, -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("forged token status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	c := newClient(t)
	token := c.authUser("valid@example.com")

	res := c.do(fiber.MethodGet, "/api/todos", token, nil)
	if res.status != fiber.StatusOK {
		t.Fatalf("valid token status = %d, want 200 (%s)", res.status, res.raw)
	}
}

func TestCORS_PreflightAllowed(t *testing.T) {
	c := newClient(t)

	req := httptest.NewRequest(fiber.MethodOptions, "/api/todos", nil)
	req.Header.Set("Origin", "http://localhost:5500")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, _ := c.app.Test(req, -1)

	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected an Access-Control-Allow-Origin header")
	}
}

func TestRequestID_HeaderPresent(t *testing.T) {
	c := newClient(t)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	resp, _ := c.app.Test(req, -1)
	if resp.Header.Get("X-Request-Id") == "" {
		t.Error("expected an X-Request-Id response header")
	}
}
