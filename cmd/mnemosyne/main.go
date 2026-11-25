package main

import (
	"context"
	"embed"
	"flag"
	"log/slog"
	"os"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/database"
	"github.com/profmoo/mnemosyne/internal/server"
)

//go:embed config/*.yaml
var configFiles embed.FS

func main() {
	ctx := context.Background()

	environment := flag.String("config", "local", "embedded config to load: local or docker")
	flag.Parse()

	cfg, err := config.LoadConfig(configFiles, "config/"+*environment+".yaml")
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	app := fx.New(
		fx.Supply(
			cfg,
			fx.Annotate(ctx, fx.As(new(context.Context))),
		),
		fx.Provide(
			database.NewDatabaseClient,
			database.NewDatabaseQuerier,
			server.NewServer,
			server.NewHandler,
			server.ConstructMiddlewareStack,
			server.NewLoggingMiddleware,
		),
		fx.Invoke(server.RegisterServer),
	)

	app.Run()
}
