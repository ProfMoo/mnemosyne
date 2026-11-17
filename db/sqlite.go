// Package db provides SQLite setup and SQL sources for code generation.
package db

//go:generate go tool sqlc generate -f sqlc.yaml

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open opens SQLite and initializes the example schema and greeting.
// This bootstrap is not a migration system for future schema changes.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	// Keep in-memory databases on one connection and serialize example writes.
	database.SetMaxOpenConns(1)
	if _, err := database.ExecContext(ctx, schema); err != nil {
		database.Close()
		return nil, fmt.Errorf("initialize SQLite: %w", err)
	}
	return database, nil
}
