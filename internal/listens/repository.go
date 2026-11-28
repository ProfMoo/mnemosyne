package listens

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNotFound identifies a listen that does not exist.
var ErrNotFound = errors.New("listen not found")

// ErrInvalidListen identifies an incomplete or invalid listen.
var ErrInvalidListen = errors.New("listen requires ID, timestamps, recording, and nonnegative duration")

// ListenCreator persists new domain listens.
type ListenCreator interface {
	Create(context.Context, *Listen) error
}

// ListenGetter retrieves a domain listen by ID.
type ListenGetter interface {
	Get(context.Context, uuid.UUID) (*Listen, error)
}

// ListenLister retrieves domain listens in storage order.
type ListenLister interface {
	List(context.Context) ([]Listen, error)
}

// Validate checks the required fields shared by storage and API mapping.
func (listen *Listen) Validate() error {
	const minimumDuration = 0
	if listen == nil || listen.ID == nil || listen.CreatedAt == nil || listen.UpdatedAt == nil ||
		listen.Recording == nil || listen.Recording.Duration < minimumDuration {
		return ErrInvalidListen
	}
	return nil
}
