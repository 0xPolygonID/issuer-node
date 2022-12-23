package schema

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/polygonid/sh-id-platform/internal/log"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migrate runs migrations on the databaseURL
func Migrate(databaseURL string) error {
	var db *sql.DB
	// setup database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("error open connection with database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error(context.Background(), "closing database", err)
		}
	}()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("error setting dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("error trying to run migrations: %w", err)
	}

	return nil
}
