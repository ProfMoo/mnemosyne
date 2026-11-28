// Package db provides SQLite setup and SQL sources for code generation.
package db

//go:generate go tool sqlc generate -f sqlc.yaml

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	_ "modernc.org/sqlite" // Register the SQLite database/sql driver.
)

//go:embed schema/*.sql
var schema embed.FS

// Open opens SQLite and initializes the example schema and greeting.
// This bootstrap is not a migration system for future schema changes.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	statements, err := readSchema()
	if err != nil {
		return nil, err
	}

	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}
	// Keep in-memory databases on one connection and serialize example writes.
	const maxOpenConnections = 1
	database.SetMaxOpenConns(maxOpenConnections)
	if _, execErr := database.ExecContext(ctx, statements); execErr != nil {
		return nil, fmt.Errorf("initialize SQLite: %w", errors.Join(execErr, database.Close()))
	}
	return database, nil
}

func readSchema() (string, error) {
	paths, err := fs.Glob(schema, "schema/*.sql")
	if err != nil {
		return "", fmt.Errorf("find SQLite schema files: %w", err)
	}

	statements := make([]string, len(paths))
	for index, path := range paths {
		content, readErr := schema.ReadFile(path)
		if readErr != nil {
			return "", fmt.Errorf("read SQLite schema %q: %w", path, readErr)
		}
		statements[index] = string(content)
	}
	return strings.Join(statements, "\n"), nil
}
