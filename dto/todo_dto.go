package dto

import (
	"strings"

	"github.com/SagarR674/todo-api/models"
)

// CreateTodoRequest is the body for POST /api/todos.
type CreateTodoRequest struct {
	Title       string `json:"title"       validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=5000"`
	Status      string `json:"status"      validate:"omitempty,oneof=pending in_progress completed"`
	Priority    string `json:"priority"    validate:"omitempty,oneof=low medium high"`
	DueDate     string `json:"due_date"    validate:"omitempty,datetime=2006-01-02"`
	CategoryIDs []uint `json:"category_ids" validate:"omitempty,dive,gt=0"`
}

// UpdateTodoRequest is the body for PUT /api/todos/:id. All fields are optional;
// only the ones supplied are changed.
type UpdateTodoRequest struct {
	Title       *string `json:"title"       validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
	Status      *string `json:"status"      validate:"omitempty,oneof=pending in_progress completed"`
	Priority    *string `json:"priority"    validate:"omitempty,oneof=low medium high"`
	DueDate     *string `json:"due_date"    validate:"omitempty,datetime=2006-01-02"`
	CategoryIDs *[]uint `json:"category_ids" validate:"omitempty,dive,gt=0"`
}

// UpdateStatusRequest is the body for PATCH /api/todos/:id/status.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending in_progress completed"`
}

// ListTodosQuery holds the parsed query string for GET /api/todos.
type ListTodosQuery struct {
	Page     int
	Limit    int
	Status   string
	Priority string
	Search   string
	Category string
	Sort     string
	Order    string
}

// sortWhitelist maps the public sort keys to real column names.
var sortWhitelist = map[string]string{
	"created_at": "todos.created_at",
	"updated_at": "todos.updated_at",
	"due_date":   "todos.due_date",
	"priority":   "todos.priority",
	"title":      "todos.title",
}

// Normalize applies defaults and clamps to the query.
func (q *ListTodosQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	if _, ok := sortWhitelist[q.Sort]; !ok {
		q.Sort = "created_at"
	}
	if o := strings.ToLower(q.Order); o == "asc" {
		q.Order = "asc"
	} else {
		q.Order = "desc"
	}
}

// OrderClause returns a safe "column direction" string for GORM's Order().
func (q *ListTodosQuery) OrderClause() string {
	col := sortWhitelist[q.Sort]
	return col + " " + q.Order
}

// Offset is the SQL offset for the current page.
func (q *ListTodosQuery) Offset() int { return (q.Page - 1) * q.Limit }

// TodoResponse is the API view of a todo.
type TodoResponse struct {
	ID          uint               `json:"id"`
	UserID      uint               `json:"user_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	Priority    string             `json:"priority"`
	DueDate     *models.Date       `json:"due_date"`
	Categories  []CategoryResponse `json:"categories"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// PaginatedTodos is the data payload for GET /api/todos.
type PaginatedTodos struct {
	Items      []TodoResponse `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination describes the current page of a list response.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

// NewTodoResponse maps a model to its API representation.
func NewTodoResponse(t *models.Todo) TodoResponse {
	cats := make([]CategoryResponse, 0, len(t.Categories))
	for i := range t.Categories {
		cats = append(cats, NewCategoryResponse(&t.Categories[i]))
	}
	return TodoResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		DueDate:     t.DueDate,
		Categories:  cats,
		CreatedAt:   t.CreatedAt.Format(timeLayout),
		UpdatedAt:   t.UpdatedAt.Format(timeLayout),
	}
}

// NewPagination builds pagination metadata.
func NewPagination(page, limit int, total int64) Pagination {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1 && total > 0,
	}
}
