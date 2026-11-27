package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/listens"
)

type sqliteListenRepository struct {
	db *sql.DB
}

func NewSqliteListenRepository(db *sql.DB) *sqliteListenRepository {
	return &sqliteListenRepository{db: db}
}

type NewSqliteListenRepositoryParams struct {
	fx.In

	DB *sql.DB
}

func (r *sqliteListenRepository) Create(_ context.Context, listen *listens.Listen) error {
	// print the info here for debugging
	fmt.Println("Saving listen:", listen)

	return nil
}
