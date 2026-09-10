package repository

import (
	"errors"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TodoRepository provides access to the todos table. Every method is scoped by
// userID so a user can only ever touch their own rows.
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository builds a TodoRepository.
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create inserts a todo (with any category associations already set).
func (r *TodoRepository) Create(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

// FindByIDForUser returns a single todo owned by userID, or ErrNotFound.
func (r *TodoRepository) FindByIDForUser(id, userID uint) (*models.Todo, error) {
	var todo models.Todo
	err := r.db.Preload("Categories").
		Where("id = ? AND user_id = ?", id, userID).
		First(&todo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// List returns a page of todos for userID plus the total count matching the
// filters (ignoring pagination).
func (r *TodoRepository) List(userID uint, q *dto.ListTodosQuery) ([]models.Todo, int64, error) {
	base := r.db.Model(&models.Todo{}).Where("todos.user_id = ?", userID)

	if q.Status != "" {
		base = base.Where("todos.status = ?", q.Status)
	}
	if q.Priority != "" {
		base = base.Where("todos.priority = ?", q.Priority)
	}
	if q.Search != "" {
		base = base.Where("todos.title LIKE ?", "%"+q.Search+"%")
	}
	if q.Category != "" {
		base = base.
			Joins("JOIN todo_categories tc ON tc.todo_id = todos.id").
			Joins("JOIN categories c ON c.id = tc.category_id").
			Where("c.name = ?", q.Category)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var todos []models.Todo
	err := base.
		Preload("Categories").
		Order(q.OrderClause()).
		Limit(q.Limit).
		Offset(q.Offset()).
		Find(&todos).Error
	if err != nil {
		return nil, 0, err
	}
	return todos, total, nil
}

// Save persists changes to an existing todo's scalar fields.
func (r *TodoRepository) Save(todo *models.Todo) error {
	return r.db.Model(todo).
		Select("title", "description", "status", "priority", "due_date").
		Updates(map[string]interface{}{
			"title":       todo.Title,
			"description": todo.Description,
			"status":      todo.Status,
			"priority":    todo.Priority,
			"due_date":    todo.DueDate,
		}).Error
}

// UpdateStatus changes just the status column of a todo owned by userID.
func (r *TodoRepository) UpdateStatus(id, userID uint, status models.Status) error {
	res := r.db.Model(&models.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ReplaceCategories sets the exact category associations for a todo.
func (r *TodoRepository) ReplaceCategories(todo *models.Todo, categories []models.Category) error {
	return r.db.Model(todo).Association("Categories").Replace(categories)
}

// SoftDelete marks a todo owned by userID as deleted (sets deleted_at).
func (r *TodoRepository) SoftDelete(id, userID uint) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Todo{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CategoriesForUser returns the subset of ids that are real categories owned by
// userID (used to validate category_ids on a request).
func (r *TodoRepository) CategoriesForUser(userID uint, ids []uint) ([]models.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var categories []models.Category
	err := r.db.
		Clauses(clause.OrderBy{Columns: []clause.OrderByColumn{{Column: clause.Column{Name: "id"}}}}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Find(&categories).Error
	return categories, err
}
