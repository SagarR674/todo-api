// Package routes wires configuration, the database, middleware and controllers
// into the Fiber app.
package routes

import (
	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/controllers"
	"github.com/SagarR674/todo-api/middleware"
	"github.com/SagarR674/todo-api/repository"
	"github.com/SagarR674/todo-api/services"
	"github.com/SagarR674/todo-api/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Setup builds the dependency graph and registers all routes on app.
func Setup(app *fiber.App, cfg *config.Config, db *gorm.DB) {
	// --- dependency graph -------------------------------------------------
	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	userRepo := repository.NewUserRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	authService := services.NewAuthService(userRepo, jwtManager)
	todoService := services.NewTodoService(todoRepo)
	categoryService := services.NewCategoryService(categoryRepo)

	authController := controllers.NewAuthController(authService)
	todoController := controllers.NewTodoController(todoService)
	categoryController := controllers.NewCategoryController(categoryService)
	healthController := controllers.NewHealthController(db)

	authGuard := middleware.Auth(jwtManager)

	// --- public ---------------------------------------------------------
	app.Get("/health", healthController.Check)

	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Use(middleware.RateLimiter(cfg.AuthRateLimitMax, cfg.AuthRateLimitWindow))
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)

	// --- protected -----------------------------------------------------
	todos := api.Group("/todos", authGuard)
	todos.Post("/", todoController.Create)
	todos.Get("/", todoController.List)
	todos.Get("/:id", todoController.Get)
	todos.Put("/:id", todoController.Update)
	todos.Patch("/:id/status", todoController.UpdateStatus)
	todos.Delete("/:id", todoController.Delete)

	categories := api.Group("/categories", authGuard)
	categories.Post("/", categoryController.Create)
	categories.Get("/", categoryController.List)

	// --- fallback ------------------------------------------------------
	app.Use(func(c *fiber.Ctx) error {
		return utils.Error(c, fiber.StatusNotFound, "Route not found")
	})
}
