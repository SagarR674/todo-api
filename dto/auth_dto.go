// Package dto holds request/response payload shapes and their validation rules.
package dto

import "github.com/SagarR674/todo-api/models"

// RegisterRequest is the body for POST /api/auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=255"`
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest is the body for POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse is the safe, public view of a user (no password).
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// AuthResponse is returned on successful login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// NewUserResponse maps a model to its public representation.
func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
