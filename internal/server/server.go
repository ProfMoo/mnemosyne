package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/config"
)

// NewServer returns the HTTP server for the Mnemosyne UI and API.
func NewServer(params NewServerParams) *http.Server {
	strictHandler := api.NewStrictHandler(params.Handler, params.MiddlewareStack)

	mux := http.NewServeMux()
	apiHandler := api.HandlerFromMux(strictHandler, mux)
	// Listen for API requests on the API prefix.
	prefixedAPI := http.StripPrefix(strings.TrimSuffix(params.Config.APIPrefix, "/"), api.Handler(strictHandler))
	mux.Handle(params.Config.APIPrefix, prefixedAPI)
	// Serve static files on the root path.
	static := os.DirFS(params.Config.FrontendDir)
	mux.Handle("/", http.FileServer(http.FS(static)))

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
			started := time.Now()
			logger := slog.Default().With("operation", operationID, "method", r.Method, "path", r.URL.Path)
			logger.LogAttrs(ctx, slog.LevelInfo, "request", logPayload(request))

			response, err := f(ctx, w, r, request)
			attrs := []slog.Attr{
				slog.Duration("duration", time.Since(started)),
				logPayload(response),
			}
			level := slog.LevelInfo
			if err != nil {
				level = slog.LevelError
				attrs = append(attrs, slog.Any("error", err))
			}
			logger.LogAttrs(ctx, level, "response", attrs...)

			return response, err
		}
	}
}

func logPayload(payload any) slog.Attr {
	body, err := json.Marshal(payload)
	if err != nil {
		return slog.String("payload_error", err.Error())
	}

	return slog.String("payload", string(body))
}
