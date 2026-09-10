package repository

import (
	"errors"

	"github.com/SagarR674/todo-api/models"
	"gorm.io/gorm"
)

// ErrDuplicate is returned when a unique constraint would be violated.
var ErrDuplicate = errors.New("record already exists")

// CategoryRepository provides access to the categories table, scoped per user.
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository builds a CategoryRepository.
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create inserts a new category. Returns ErrDuplicate if the user already has a
// category with the same name.
func (r *CategoryRepository) Create(category *models.Category) error {
	var count int64
	if err := r.db.Model(&models.Category{}).
		Where("user_id = ? AND name = ?", category.UserID, category.Name).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicate
	}
	return r.db.Create(category).Error
}

// ListForUser returns all categories owned by userID, ordered by name.
func (r *CategoryRepository) ListForUser(userID uint) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Where("user_id = ?", userID).Order("name asc").Find(&categories).Error
	return categories, err
}
