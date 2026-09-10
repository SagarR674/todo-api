package controllers

import (
	"github.com/SagarR674/todo-api/dto"
	"github.com/SagarR674/todo-api/middleware"
	"github.com/SagarR674/todo-api/services"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
)

// CategoryController exposes the authenticated category/tag endpoints.
type CategoryController struct {
	categories *services.CategoryService
}

// NewCategoryController builds a CategoryController.
func NewCategoryController(categories *services.CategoryService) *CategoryController {
	return &CategoryController{categories: categories}
}

// Create handles POST /api/categories.
func (h *CategoryController) Create(c *fiber.Ctx) error {
	var req dto.CreateCategoryRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	category, err := h.categories.Create(middleware.UserID(c), req)
	if err != nil {
		return serviceError(c, err)
	}
	return utils.Success(c, fiber.StatusCreated, "Category created successfully", dto.NewCategoryResponse(category))
}

// List handles GET /api/categories.
func (h *CategoryController) List(c *fiber.Ctx) error {
	categories, err := h.categories.List(middleware.UserID(c))
	if err != nil {
		return serviceError(c, err)
	}

	out := make([]dto.CategoryResponse, 0, len(categories))
	for i := range categories {
		out = append(out, dto.NewCategoryResponse(&categories[i]))
	}
	return utils.Success(c, fiber.StatusOK, "Categories retrieved successfully", out)
}
