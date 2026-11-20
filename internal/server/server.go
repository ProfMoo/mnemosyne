package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/config"
)

// NewHandler returns the HTTP handler for the Mnemosyne API.
func NewServer(params NewServerParams) *http.Server {
	strictHandler := api.NewStrictHandler(params.Handler, params.MiddlewareStack)

	apiHandler := api.Handler(strictHandler)

	apiServer := &http.Server{
		Addr:              params.Config.BindAddress,
		Handler:           apiHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return apiServer
}

type NewServerParams struct {
	fx.In

	Handler *handler

	Config *config.Config

	MiddlewareStack []api.StrictMiddlewareFunc
}

func ConstructMiddlewareStack(params ConstructMiddlewareStackParams) []api.StrictMiddlewareFunc {
	middlewares := []api.StrictMiddlewareFunc{}

	middlewares = append(middlewares, params.LoggingMiddleware)

	return middlewares
}

type ConstructMiddlewareStackParams struct {
	fx.In

	LoggingMiddleware api.StrictMiddlewareFunc
}

func NewLoggingMiddleware() api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			slog.Info("request", "url", r.URL.Path)
			return f(ctx, w, r, request)
		}
	}
}
