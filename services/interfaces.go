package services

import (
	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
)

// The interfaces below are declared by the services (the consumer), following
// the Go convention of "accept interfaces, return structs". The concrete
// repository types satisfy them, and tests can substitute lightweight fakes.

// UserRepo is the persistence surface the AuthService needs.
type UserRepo interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	ExistsByEmail(email string) (bool, error)
}

// TokenIssuer issues signed JWTs for authenticated users.
type TokenIssuer interface {
	Generate(userID uint) (string, error)
}

// TodoRepo is the persistence surface the TodoService needs.
type TodoRepo interface {
	Create(todo *models.Todo) error
	FindByIDForUser(id, userID uint) (*models.Todo, error)
	List(userID uint, q *dto.ListTodosQuery) ([]models.Todo, int64, error)
	Save(todo *models.Todo) error
	UpdateStatus(id, userID uint, status models.Status) error
	ReplaceCategories(todo *models.Todo, categories []models.Category) error
	SoftDelete(id, userID uint) error
	CategoriesForUser(userID uint, ids []uint) ([]models.Category, error)
}

// CategoryRepo is the persistence surface the CategoryService needs.
type CategoryRepo interface {
	Create(category *models.Category) error
	ListForUser(userID uint) ([]models.Category, error)
}
