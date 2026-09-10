// Package server assembles the Fiber application: global middleware plus all
// routes. It is shared by the main binary and the integration test suite so
// both exercise exactly the same wiring.
package server

import (
	"time"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/middleware"
	"github.com/SagarR674/todo-api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"
)

// New builds the fully wired Fiber app for the given config and database.
func New(cfg *config.Config, db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Todo Management API",
		ErrorHandler:          middleware.ErrorHandler,
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(middleware.RequestLogger())
	app.Use(middleware.RateLimiter(cfg.RateLimitMax, cfg.RateLimitWindow))

	routes.Setup(app, cfg, db)
	return app
}
