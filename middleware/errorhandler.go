package middleware

import (
	"errors"

	"github.com/SagarR674/todo-api/pkg/logger"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is the app-wide Fiber error handler. Any error returned by a
// handler (or a panic caught by the recover middleware) that was not already
// written as a response envelope ends up here and is converted into a
// consistent JSON error. 5xx errors are logged with context.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		if code < fiber.StatusInternalServerError {
			message = fe.Message
		}
	}

	if code >= fiber.StatusInternalServerError {
		logger.L().Error("unhandled error",
			"method", c.Method(),
			"path", c.OriginalURL(),
			"request_id", c.GetRespHeader(fiber.HeaderXRequestID),
			"error", err.Error(),
		)
	}

	return utils.Error(c, code, message)
}
