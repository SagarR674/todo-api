package middleware

import (
	"log/slog"
	"time"

	"github.com/SagarR674/todo-api/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// RequestLogger logs one structured line per request: method, path, status,
// latency, client IP and the request id set by the requestid middleware.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.OriginalURL()),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.IP()),
			slog.String("request_id", c.GetRespHeader(fiber.HeaderXRequestID)),
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		switch {
		case status >= 500:
			logger.L().Error("request", attrs...)
		case status >= 400:
			logger.L().Warn("request", attrs...)
		default:
			logger.L().Info("request", attrs...)
		}
		return err
	}
}
