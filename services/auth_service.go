package services

import (
	"errors"
	"strings"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/repository"
	"github.com/SagarR674/todo-api/utils"
)

// AuthService handles registration and login.
type AuthService struct {
	users UserRepo
	jwt   TokenIssuer
}

// NewAuthService builds an AuthService.
func NewAuthService(users UserRepo, jwt TokenIssuer) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

// Register creates a new account with a securely hashed password.
func (s *AuthService) Register(req dto.RegisterRequest) (*models.User, error) {
	email := normalizeEmail(req.Email)

	exists, err := s.users.ExistsByEmail(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    email,
		Password: hash,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login verifies credentials and returns a signed JWT plus the user.
func (s *AuthService) Login(req dto.LoginRequest) (string, *models.User, error) {
	user, err := s.users.FindByEmail(normalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !utils.CheckPassword(user.Password, req.Password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

// Profile returns the account for the given user ID (used by GET /api/auth/me).
func (s *AuthService) Profile(userID uint) (*models.User, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
