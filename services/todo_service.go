package services

import (
	"errors"
	"strings"

	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/repository"
)

// TodoService holds the business logic for todo management.
type TodoService struct {
	todos TodoRepo
}

// NewTodoService builds a TodoService.
func NewTodoService(todos TodoRepo) *TodoService {
	return &TodoService{todos: todos}
}

// Create makes a new todo for userID, attaching any valid categories.
func (s *TodoService) Create(userID uint, req dto.CreateTodoRequest) (*models.Todo, error) {
	status := models.Status(req.Status)
	if status == "" {
		status = models.StatusPending
	}
	priority := models.Priority(req.Priority)
	if priority == "" {
		priority = models.PriorityMedium
	}

	todo := &models.Todo{
		UserID:      userID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      status,
		Priority:    priority,
	}

	if req.DueDate != "" {
		d, err := models.ParseDate(req.DueDate)
		if err != nil {
			return nil, err
		}
		todo.DueDate = &d
	}

	categories, err := s.resolveCategories(userID, req.CategoryIDs)
	if err != nil {
		return nil, err
	}
	todo.Categories = categories

	if err := s.todos.Create(todo); err != nil {
		return nil, err
	}
	return s.todos.FindByIDForUser(todo.ID, userID)
}

// List returns a page of the user's todos plus pagination metadata.
func (s *TodoService) List(userID uint, q *dto.ListTodosQuery) (*dto.PaginatedTodos, error) {
	q.Normalize()

	todos, total, err := s.todos.List(userID, q)
	if err != nil {
		return nil, err
	}

	items := make([]dto.TodoResponse, 0, len(todos))
	for i := range todos {
		items = append(items, dto.NewTodoResponse(&todos[i]))
	}

	return &dto.PaginatedTodos{
		Items:      items,
		Pagination: dto.NewPagination(q.Page, q.Limit, total),
	}, nil
}

// Get returns a single todo owned by userID.
func (s *TodoService) Get(userID, id uint) (*models.Todo, error) {
	todo, err := s.todos.FindByIDForUser(id, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return todo, nil
}

// Update applies a partial update to a todo owned by userID.
func (s *TodoService) Update(userID, id uint, req dto.UpdateTodoRequest) (*models.Todo, error) {
	todo, err := s.todos.FindByIDForUser(id, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}

	if req.Title != nil {
		todo.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Status != nil {
		todo.Status = models.Status(*req.Status)
	}
	if req.Priority != nil {
		todo.Priority = models.Priority(*req.Priority)
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			todo.DueDate = nil
		} else {
			d, err := models.ParseDate(*req.DueDate)
			if err != nil {
				return nil, err
			}
			todo.DueDate = &d
		}
	}

	if err := s.todos.Save(todo); err != nil {
		return nil, err
	}

	if req.CategoryIDs != nil {
		categories, err := s.resolveCategories(userID, *req.CategoryIDs)
		if err != nil {
			return nil, err
		}
		if err := s.todos.ReplaceCategories(todo, categories); err != nil {
			return nil, err
		}
	}

	return s.todos.FindByIDForUser(id, userID)
}

// UpdateStatus changes only the status of a todo owned by userID.
func (s *TodoService) UpdateStatus(userID, id uint, status models.Status) (*models.Todo, error) {
	if err := s.todos.UpdateStatus(id, userID, status); err != nil {
		return nil, mapNotFound(err)
	}
	return s.todos.FindByIDForUser(id, userID)
}

// Delete soft-deletes a todo owned by userID.
func (s *TodoService) Delete(userID, id uint) error {
	if err := s.todos.SoftDelete(id, userID); err != nil {
		return mapNotFound(err)
	}
	return nil
}

// resolveCategories validates that every requested category id belongs to the
// user and returns the loaded rows.
func (s *TodoService) resolveCategories(userID uint, ids []uint) ([]models.Category, error) {
	unique := dedupe(ids)
	if len(unique) == 0 {
		return nil, nil
	}
	categories, err := s.todos.CategoriesForUser(userID, unique)
	if err != nil {
		return nil, err
	}
	if len(categories) != len(unique) {
		return nil, ErrCategoryNotFound
	}
	return categories, nil
}

func dedupe(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func mapNotFound(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrTodoNotFound
	}
	return err
}
