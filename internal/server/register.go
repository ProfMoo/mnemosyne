package server

import (
	"context"
	"log/slog"
	"net/http"

	"go.uber.org/fx"
)

func RegisterServer(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				slog.Info("listening", "addr", server.Addr)

				if err := server.ListenAndServe(); err != nil {
					slog.Error("serve HTTP", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
