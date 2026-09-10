package dto

import "github.com/SagarR674/todo-api/models"

// CreateCategoryRequest is the body for POST /api/categories.
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CategoryResponse is the API view of a category.
type CategoryResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// NewCategoryResponse maps a model to its API representation.
func NewCategoryResponse(c *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt.Format(timeLayout),
	}
}
