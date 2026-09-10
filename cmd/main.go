// Command todo-api is the entrypoint for the Todo Management API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/database"
	"github.com/SagarR674/todo-api/middleware"
	"github.com/SagarR674/todo-api/pkg/logger"
	"github.com/SagarR674/todo-api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// logger not up yet; use the stdlib default.
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	log := logger.L()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("database connected", "name", cfg.DBName, "host", cfg.DBHost)

	app := fiber.New(fiber.Config{
		AppName:               "Todo Management API",
		ErrorHandler:          middleware.ErrorHandler,
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.RateLimiter(cfg.RateLimitMax, cfg.RateLimitWindow))

	routes.Setup(app, cfg, db)

	// graceful shutdown
	go func() {
		addr := ":" + cfg.Port
		log.Info("server listening", "addr", addr, "env", cfg.AppEnv)
		if err := app.Listen(addr); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Info("stopped")
}
