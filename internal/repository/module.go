package repository

import (
	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/server"
)

var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(
			NewSqliteListenRepository,
			fx.As(new(server.ListenRepository)),
		),
	),
)
