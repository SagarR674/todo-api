package controllers

import (
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// HealthController reports service and database health.
type HealthController struct {
	db *gorm.DB
}

// NewHealthController builds a HealthController.
func NewHealthController(db *gorm.DB) *HealthController {
	return &HealthController{db: db}
}

// Check handles GET /health.
func (h *HealthController) Check(c *fiber.Ctx) error {
	sqlDB, err := h.db.DB()
	if err == nil {
		err = sqlDB.Ping()
	}
	if err != nil {
		return utils.Error(c, fiber.StatusServiceUnavailable, "Database unreachable")
	}
	return utils.Success(c, fiber.StatusOK, "Service healthy", fiber.Map{
		"status":   "ok",
		"database": "connected",
	})
}
