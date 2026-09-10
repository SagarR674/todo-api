package database

import (
	"errors"
	"fmt"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/migrations"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Direction selects which way Migrate should move the schema.
type Direction int

const (
	// Up applies all pending migrations.
	Up Direction = iota
	// Down rolls the schema all the way back.
	Down
	// Step rolls back exactly one migration.
	StepDown
)

// newMigrator builds a *migrate.Migrate backed by the embedded SQL files and a
// fresh database/sql connection. multiStatements is required because some
// migration files contain more than one statement. The target database is
// created if it does not yet exist.
func newMigrator(cfg *config.Config) (*migrate.Migrate, error) {
	if err := ensureDatabase(cfg); err != nil {
		return nil, err
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("load embedded migrations: %w", err)
	}

	db, err := openSQL(cfg.MigrationDSN())
	if err != nil {
		return nil, err
	}

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("migrate driver: %w", err)
	}

	return migrate.NewWithInstance("iofs", src, "mysql", driver)
}

// Migrate runs the embedded migrations in the requested direction. It first
// ensures the target database exists. ErrNoChange is treated as success.
func Migrate(cfg *config.Config, dir Direction) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	switch dir {
	case Up:
		err = m.Up()
	case Down:
		err = m.Down()
	case StepDown:
		err = m.Steps(-1)
	default:
		return fmt.Errorf("unknown migration direction %d", dir)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// MigrateUp is a convenience wrapper for the common case.
func MigrateUp(cfg *config.Config) error { return Migrate(cfg, Up) }

// MigrationVersion reports the current schema version and whether the last
// migration left the schema in a dirty state.
func MigrationVersion(cfg *config.Config) (version uint, dirty bool, err error) {
	m, err := newMigrator(cfg)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	v, d, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return v, d, err
}

// MigrationForce clears a dirty state by pinning the schema to version v
// without running any migration.
func MigrationForce(cfg *config.Config, v int) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Force(v)
}
