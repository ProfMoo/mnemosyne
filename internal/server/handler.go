package server

import (
	"context"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	db_gen "github.com/profmoo/mnemosyne/gen/db"
)

type handler struct {
	queries *db_gen.Queries
}

func NewHandler(params NewHandlerParams) *handler {
	return &handler{
		queries: params.Queries,
	}
}

type NewHandlerParams struct {
	fx.In

	Queries *db_gen.Queries
}

func (s *handler) GetHello(ctx context.Context, _ api.GetHelloRequestObject) (api.GetHelloResponseObject, error) {
	message, err := s.queries.GetGreeting(ctx)
	if err != nil {
		return nil, err
	}
	return api.GetHello200JSONResponse{Message: message}, nil
}

func (s *handler) GetListen(ctx context.Context, _ api.GetListenRequestObject) (api.GetListenResponseObject, error) {
	return api.GetListen500JSONResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: "Not implemented",
	}, nil
}

func (s *handler) GetListens(ctx context.Context, _ api.GetListensRequestObject) (api.GetListensResponseObject, error) {
	return api.GetListens500JSONResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: "Not implemented",
	}, nil
}

func (s *handler) SubmitListens(
	ctx context.Context,
	_ api.SubmitListensRequestObject,
) (api.SubmitListensResponseObject, error) {
	return api.SubmitListens500JSONResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: "Not implemented",
	}, nil
}
