package middleware

import (
	"time"

	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter returns a fixed-window rate limiter keyed by client IP. When the
// limit is exceeded it responds with a 429 in the standard envelope.
func RateLimiter(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later")
		},
	})
}
