package controllers

import (
	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/middleware"
	"github.com/SagarR674/todo-api/models"
	"github.com/SagarR674/todo-api/services"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// TodoController exposes the authenticated todo CRUD endpoints.
type TodoController struct {
	todos *services.TodoService
}

// NewTodoController builds a TodoController.
func NewTodoController(todos *services.TodoService) *TodoController {
	return &TodoController{todos: todos}
}

// Create handles POST /api/todos.
func (h *TodoController) Create(c *fiber.Ctx) error {
	var req dto.CreateTodoRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	todo, err := h.todos.Create(middleware.UserID(c), req)
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusCreated, "Todo created successfully", dto.NewTodoResponse(todo))
}

// List handles GET /api/todos with pagination, filtering, search and sorting.
func (h *TodoController) List(c *fiber.Ctx) error {
	q := &dto.ListTodosQuery{
		Page:     c.QueryInt("page", 1),
		Limit:    c.QueryInt("limit", 10),
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Sort:     c.Query("sort"),
		Order:    c.Query("order"),
	}

	if q.Status != "" && !models.Status(q.Status).IsValid() {
		return utils.ValidationError(c, map[string]string{"status": "must be one of: pending, in_progress, completed"})
	}
	if q.Priority != "" && !models.Priority(q.Priority).IsValid() {
		return utils.ValidationError(c, map[string]string{"priority": "must be one of: low, medium, high"})
	}

	result, err := h.todos.List(middleware.UserID(c), q)
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusOK, "Todos retrieved successfully", result)
}

// Get handles GET /api/todos/:id.
func (h *TodoController) Get(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid todo id")
	}

	todo, err := h.todos.Get(middleware.UserID(c), id)
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusOK, "Todo retrieved successfully", dto.NewTodoResponse(todo))
}

// Update handles PUT /api/todos/:id.
func (h *TodoController) Update(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid todo id")
	}

	var req dto.UpdateTodoRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	todo, err := h.todos.Update(middleware.UserID(c), id, req)
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusOK, "Todo updated successfully", dto.NewTodoResponse(todo))
}

// UpdateStatus handles PATCH /api/todos/:id/status.
func (h *TodoController) UpdateStatus(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid todo id")
	}

	var req dto.UpdateStatusRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	todo, err := h.todos.UpdateStatus(middleware.UserID(c), id, models.Status(req.Status))
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusOK, "Todo status updated successfully", dto.NewTodoResponse(todo))
}

// Delete handles DELETE /api/todos/:id (soft delete).
func (h *TodoController) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid todo id")
	}

	if err := h.todos.Delete(middleware.UserID(c), id); err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusOK, "Todo deleted successfully", nil)
}

func parseID(c *fiber.Ctx) (uint, bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, false
	}
	return uint(id), true
}
