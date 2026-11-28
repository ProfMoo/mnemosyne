package main

import (
	"context"
	"embed"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/database"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
	"github.com/profmoo/mnemosyne/internal/repository"
	"github.com/profmoo/mnemosyne/internal/server"
)

//go:embed config/*.yaml
var configFiles embed.FS

func main() {
	ctx := context.Background()

	configPath := flag.String("config", "config/local.yaml", "full path to an embedded configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(configFiles, *configPath)
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	app := fx.New(
		fx.Supply(
			cfg,
			cfg.MusicBrainz,
			&http.Client{},
			time.Now,
			fx.Annotate(ctx, fx.As(new(context.Context))),
		),
		fx.Provide(
			database.NewDatabaseClient,
			database.NewDatabaseQuerier,
			fx.Annotate(musicbrainz.NewClient, fx.As(new(matching.Catalog))),
			fx.Annotate(matching.NewService, fx.As(new(matching.Matcher))),
			server.NewServer,
			server.NewHandler,
			server.ConstructMiddlewareStack,
			server.NewLoggingMiddleware,
			server.NewSubmitHandler,
			server.NewGetListenHandler,
			server.NewGetListensHandler,
		),
		repository.Module,
		fx.Invoke(server.RegisterServer),
	)

	app.Run()
}
