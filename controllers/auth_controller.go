package controllers

import (
	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/services"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// AuthController exposes registration and login endpoints.
type AuthController struct {
	auth *services.AuthService
}

// NewAuthController builds an AuthController.
func NewAuthController(auth *services.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

// Register handles POST /api/auth/register.
func (h *AuthController) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	user, err := h.auth.Register(req)
	if err != nil {
		return serviceError(c, err)
	}

	return utils.Success(c, fiber.StatusCreated, "User registered successfully", dto.NewUserResponse(user))
}

// Login handles POST /api/auth/login.
func (h *AuthController) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	token, user, err := h.auth.Login(req)
	if err != nil {
		return serviceError(c, err)
	}

	return utils.Success(c, fiber.StatusOK, "Login successful", dto.AuthResponse{
		Token: token,
		User:  dto.NewUserResponse(user),
	})
}
