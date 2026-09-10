// Package utils holds small cross-cutting helpers: the HTTP response envelope,
// JWT helpers, password hashing and request validation.
package utils

import "github.com/gofiber/fiber/v2"

// Response is the single JSON shape returned by every endpoint.
//
//	{ "success": true,  "message": "...", "data": {...} }
//	{ "success": false, "message": "...", "errors": {...} }
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Success writes a success envelope with the given status code.
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error writes an error envelope with the given status code.
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
	})
}

// ValidationError writes a 400 envelope with a per-field error map.
func ValidationError(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Message: "Validation failed",
		Errors:  fields,
	})
}
