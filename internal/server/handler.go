// Package server implements the HTTP API and static frontend server.
package server

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	dbgen "github.com/profmoo/mnemosyne/gen/db"
)

type handler struct {
	queries       *dbgen.Queries
	submitHandler *SubmitHandler
	getHandler    *GetListenHandler
	listHandler   *GetListensHandler
}

// NewHandler assembles the operation-specific API handlers.
func NewHandler(params NewHandlerParams) *handler {
	return &handler{
		queries:       params.Queries,
		submitHandler: params.SubmitHandler,
		getHandler:    params.GetListenHandler,
		listHandler:   params.GetListensHandler,
	}
}

// NewHandlerParams supplies queries and operation handlers.
type NewHandlerParams struct {
	fx.In

	Queries           *dbgen.Queries
	SubmitHandler     *SubmitHandler
	GetListenHandler  *GetListenHandler
	GetListensHandler *GetListensHandler
}

func (s *handler) GetHello(ctx context.Context, _ api.GetHelloRequestObject) (api.GetHelloResponseObject, error) {
	message, err := s.queries.GetGreeting(ctx)
	if err != nil {
		return nil, fmt.Errorf("get greeting: %w", err)
	}
	return api.GetHello200JSONResponse{Message: message}, nil
}

func (s *handler) GetListen(
	ctx context.Context,
	request api.GetListenRequestObject,
) (api.GetListenResponseObject, error) {
	return s.getHandler.Handle(ctx, request)
}

func (s *handler) GetListens(
	ctx context.Context,
	request api.GetListensRequestObject,
) (api.GetListensResponseObject, error) {
	return s.listHandler.Handle(ctx, request)
}

func (s *handler) SubmitListens(
	ctx context.Context,
	req api.SubmitListensRequestObject,
) (api.SubmitListensResponseObject, error) {
	return s.submitHandler.Handle(ctx, req)
}
