package server

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/listens"
)

// GetListensHandler lists persisted listens in repository order.
type GetListensHandler struct {
	repository listens.ListenLister
}

// GetListensHandlerParams supplies domain persistence.
type GetListensHandlerParams struct {
	fx.In

	Repository listens.ListenLister
}

// NewGetListensHandler constructs the list handler.
func NewGetListensHandler(params GetListensHandlerParams) *GetListensHandler {
	return &GetListensHandler{repository: params.Repository}
}

// Handle lists listens and maps each row to the public response.
func (handler *GetListensHandler) Handle(
	ctx context.Context,
	_ api.GetListensRequestObject,
) (api.GetListensResponseObject, error) {
	items, err := handler.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("load listens: %w", err)
	}
	result := make([]api.Listen, len(items))
	for index := range items {
		result[index], err = listens.ToAPI(&items[index])
		if err != nil {
			return nil, fmt.Errorf("map listen: %w", err)
		}
	}
	return api.GetListens200JSONResponse{Listens: result}, nil
}
