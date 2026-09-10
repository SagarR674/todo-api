// Package controllers contains the Fiber HTTP handlers. Handlers are thin: parse
// and validate input, delegate to a service, and translate the result (or a
// sentinel error) into the standard response envelope.
package controllers

import (
	"errors"

	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/services"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// bindAndValidate parses the JSON body into dst and runs struct validation.
// On failure it writes the appropriate 400 response and returns false.
func bindAndValidate(c *fiber.Ctx, dst interface{}) bool {
	if err := c.BodyParser(dst); err != nil {
		_ = utils.Error(c, fiber.StatusBadRequest, "Request body must be valid JSON")
		return false
	}
	if fields := utils.ValidateStruct(dst); fields != nil {
		_ = utils.ValidationError(c, fields)
		return false
	}
	return true
}

// serviceError maps a domain error from the service layer onto an HTTP response.
func serviceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrEmailTaken):
		return utils.Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidCredentials):
		return utils.Error(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, services.ErrTodoNotFound):
		// Also returned when a todo exists but belongs to another user: 404
		// (rather than 403) avoids leaking that the id exists.
		return utils.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrCategoryNotFound):
		return utils.Error(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrCategoryExists):
		return utils.Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, models.ErrInvalidDate):
		return utils.Error(c, fiber.StatusBadRequest, err.Error())
	default:
		// Unknown error: let the central error handler log and return 500.
		return err
	}
}
