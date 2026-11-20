package database

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/db"
	db_gen "github.com/profmoo/mnemosyne/gen/db"
)

func NewDatabaseClient(ctx context.Context, lc fx.Lifecycle) (*sql.DB, error) {
	database, err := db.Open(ctx, "mnemosyne.sqlite")
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return database.Close()
		},
	})

	return database, nil
}

func NewDatabaseQuerier(params NewDatabaseQuerierParams) *db_gen.Queries {
	return db_gen.New(params.Database)
}

type NewDatabaseQuerierParams struct {
	fx.In

	Database *sql.DB
}
