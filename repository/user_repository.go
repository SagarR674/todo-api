// Package repository is the data-access layer. Each method is a thin, testable
// wrapper over GORM; all business rules live in the service layer.
package repository

import (
	"errors"

	"github.com/SagarR674/todo-api/models"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("record not found")

// UserRepository provides access to the users table.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail returns the user with the given email, or ErrNotFound.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID returns the user with the given id, or ErrNotFound.
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail reports whether an account already uses the given email.
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
