package server

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/listens"
)

// GetListenHandler retrieves one persisted listen.
type GetListenHandler struct {
	repository listens.ListenGetter
}

// GetListenHandlerParams supplies domain persistence.
type GetListenHandlerParams struct {
	fx.In

	Repository listens.ListenGetter
}

// NewGetListenHandler constructs the single-listen handler.
func NewGetListenHandler(params GetListenHandlerParams) *GetListenHandler {
	return &GetListenHandler{repository: params.Repository}
}

// Handle retrieves a listen and maps it to the public response.
func (handler *GetListenHandler) Handle(
	ctx context.Context,
	request api.GetListenRequestObject,
) (api.GetListenResponseObject, error) {
	listen, err := handler.repository.Get(ctx, request.ListenId)
	if errors.Is(err, listens.ErrNotFound) {
		return api.GetListen404JSONResponse{Code: "NOT_FOUND", Message: "Listen not found"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load listen: %w", err)
	}
	response, err := listens.ToAPI(listen)
	if err != nil {
		return nil, fmt.Errorf("map listen: %w", err)
	}
	return api.GetListen200JSONResponse(response), nil
}
