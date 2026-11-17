package server

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/gen/db"
)

type server struct {
	queries *db.Queries
}

var _ api.StrictServerInterface = server{}

// NewHandler returns the HTTP handler for the Mnemosyne API.
func NewHandler(database *sql.DB) http.Handler {
	return api.Handler(api.NewStrictHandler(server{queries: db.New(database)}, nil))
}

func (s server) GetHello(ctx context.Context, _ api.GetHelloRequestObject) (api.GetHelloResponseObject, error) {
	message, err := s.queries.GetGreeting(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetHello200JSONResponse{Message: message}, nil
}
