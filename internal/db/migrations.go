package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
)

func RunMigrations(cfg *config.Config) error {
	databaseURL := cfg.DatabaseURL
	migrationPath := cfg.Database.Migrations.Path
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	sqlDB, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to connect for migrations: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set Goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, migrationPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
