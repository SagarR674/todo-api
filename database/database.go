// Package database manages the GORM connection to MySQL.
package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/pkg/logger"
	_ "github.com/go-sql-driver/mysql" // registers the "mysql" driver for database/sql
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect opens a pooled GORM connection to the configured database. For
// developer convenience it first ensures the target schema exists (CREATE
// DATABASE IF NOT EXISTS); production deployments can pre-create the schema.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	if err := ensureDatabase(cfg); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                                   logger.NewGormLogger(cfg.IsProduction()),
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// ensureDatabase connects without selecting a schema and creates it if needed.
func ensureDatabase(cfg *config.Config) error {
	root, err := sql.Open("mysql", cfg.RootDSN())
	if err != nil {
		return fmt.Errorf("open root connection: %w", err)
	}
	defer root.Close()

	stmt := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	)
	if _, err := root.Exec(stmt); err != nil {
		return fmt.Errorf("create database %q: %w", cfg.DBName, err)
	}
	return nil
}
