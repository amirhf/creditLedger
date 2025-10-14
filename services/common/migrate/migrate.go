package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations runs database migrations from embedded files
func RunMigrations(db *sql.DB, fs embed.FS, serviceName string) error {
	log.Printf("Running migrations for %s...", serviceName)

	// Create driver from database connection
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create source from embedded filesystem
	// The path is relative to where files were embedded (migrations folder)
	sourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Printf("Migrations complete for %s: no migrations applied (empty database)", serviceName)
	} else {
		log.Printf("Migrations complete for %s: current version = %d, dirty = %v", serviceName, version, dirty)
	}

	return nil
}
