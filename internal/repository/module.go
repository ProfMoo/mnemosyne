package repository

import (
	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/listens"
)

// Module provides the domain listen repository.
var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(
			NewSqliteListenRepository,
			fx.As(new(listens.ListenCreator)),
			fx.As(new(listens.ListenGetter)),
			fx.As(new(listens.ListenLister)),
		),
	),
)
