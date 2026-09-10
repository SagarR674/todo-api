// Package middleware holds Fiber middleware: JWT auth, rate limiting and
// structured request logging.
package middleware

import (
	"strings"

	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// ctxUserID is the Locals key under which the authenticated user's ID is stored.
const ctxUserID = "userID"

// Auth returns middleware that requires a valid `Authorization: Bearer <token>`
// header. Missing, malformed, invalid or expired tokens all yield 401.
func Auth(jwt *utils.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return utils.Error(c, fiber.StatusUnauthorized, "Authorization header is required")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return utils.Error(c, fiber.StatusUnauthorized, "Authorization header must be in the format: Bearer <token>")
		}

		claims, err := jwt.Parse(parts[1])
		if err != nil {
			return utils.Error(c, fiber.StatusUnauthorized, "Invalid or expired token")
		}

		c.Locals(ctxUserID, claims.UserID)
		return c.Next()
	}
}

// UserID returns the authenticated user's ID from the request context. It must
// only be called from handlers behind Auth.
func UserID(c *fiber.Ctx) uint {
	id, _ := c.Locals(ctxUserID).(uint)
	return id
}
