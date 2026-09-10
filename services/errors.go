// Package services holds the business logic layer. Services orchestrate
// repositories, enforce ownership and domain rules, and return sentinel errors
// that controllers map to HTTP status codes.
package services

import "errors"

var (
	// ErrEmailTaken is returned when registering with an existing email.
	ErrEmailTaken = errors.New("email is already registered")
	// ErrInvalidCredentials is returned for a failed login.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrTodoNotFound is returned when a todo does not exist or is not owned
	// by the requesting user.
	ErrTodoNotFound = errors.New("todo not found")
	// ErrCategoryNotFound is returned when a referenced category is missing or
	// not owned by the user.
	ErrCategoryNotFound = errors.New("category not found")
	// ErrCategoryExists is returned on a duplicate category name.
	ErrCategoryExists = errors.New("category already exists")
)
