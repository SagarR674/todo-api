package services

import (
	"errors"
	"strings"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/repository"
)

// CategoryService holds the business logic for categories/tags.
type CategoryService struct {
	categories *repository.CategoryRepository
}

// NewCategoryService builds a CategoryService.
func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

// Create makes a new category for userID.
func (s *CategoryService) Create(userID uint, req dto.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		UserID: userID,
		Name:   strings.TrimSpace(req.Name),
	}
	if err := s.categories.Create(category); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrCategoryExists
		}
		return nil, err
	}
	return category, nil
}

// List returns all categories owned by userID.
func (s *CategoryService) List(userID uint) ([]models.Category, error) {
	return s.categories.ListForUser(userID)
}
