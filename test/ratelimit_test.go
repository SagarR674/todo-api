package test

import (
	"testing"
	"time"

	"github.com/SagarR674/todo-api/config"
	"github.com/gofiber/fiber/v2"
)

func TestRateLimit_AuthEndpoint(t *testing.T) {
	const limit = 5
	c := newClientWith(t, func(cfg *config.Config) {
		cfg.AuthRateLimitMax = limit
		cfg.AuthRateLimitWindow = time.Minute
	})

	var got429 bool
	for i := 0; i < limit+3; i++ {
		res := c.do(fiber.MethodPost, "/api/auth/login", "", fiber.Map{
			"email": "nobody@example.com", "password": "whatever",
		})
		if res.status == fiber.StatusTooManyRequests {
			got429 = true
			if res.body.Success {
				t.Error("429 body should have success=false")
			}
			break
		}
	}
	if !got429 {
		t.Fatalf("expected a 429 after %d requests", limit)
	}
}

func TestRateLimit_DoesNotBlockNormalTraffic(t *testing.T) {
	c := newClientWith(t, func(cfg *config.Config) {
		cfg.RateLimitMax = 100000
		cfg.AuthRateLimitMax = 100000
	})
	token := c.authUser("rl@example.com")

	for i := 0; i < 20; i++ {
		res := c.do(fiber.MethodGet, "/api/todos", token, nil)
		if res.status != fiber.StatusOK {
			t.Fatalf("request %d: status %d", i, res.status)
		}
	}
}
