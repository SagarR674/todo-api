// Command migrate applies or rolls back database migrations from the
// migrations/ directory. It is a thin wrapper around golang-migrate so a
// developer does not need the migrate CLI installed.
//
// Usage:
//
//	go run ./cmd/migrate up          # apply all pending migrations
//	go run ./cmd/migrate down        # roll back the most recent migration
//	go run ./cmd/migrate down-all    # roll back every migration
//	go run ./cmd/migrate version     # print current schema version
//	go run ./cmd/migrate force <v>   # clear a dirty state at version <v>
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SagarR674/todo-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|down-all|version|force <v>>")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := ensureDatabase(cfg); err != nil {
		log.Fatalf("ensure database: %v", err)
	}

	db, err := sql.Open("mysql", cfg.MigrationDSN())
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "up":
		finish(m.Up())
	case "down":
		finish(m.Steps(-1))
	case "down-all":
		finish(m.Down())
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("version=%d dirty=%t\n", v, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force <version>")
		}
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid version: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("force: %v", err)
		}
		fmt.Printf("forced to version %d\n", v)
	default:
		log.Fatalf("unknown command %q", cmd)
	}
}

func finish(err error) {
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("migration complete")
}

func ensureDatabase(cfg *config.Config) error {
	root, err := sql.Open("mysql", cfg.RootDSN())
	if err != nil {
		return err
	}
	defer root.Close()
	_, err = root.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	))
	return err
}
