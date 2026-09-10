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
	"github.com/SagarR674/todo-api/pkg/logger"
	"github.com/SagarR674/todo-api/server"
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

	if cfg.AutoMigrate {
		if err := database.MigrateUp(cfg); err != nil {
			log.Error("auto-migration failed", "error", err)
			os.Exit(1)
		}
		log.Info("database migrations applied")
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("database connected", "name", cfg.DBName, "host", cfg.DBHost)

	app := server.New(cfg, db)

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
