package listens

import (
	"time"

	"github.com/google/uuid"
)

// NOTE: Not sure if I want to validate these fields anywhere :thinking:. Probably should validate during OpenAPI validation.

// The minimal amount of information that we need to accept a listen from ANY source, with an ID attached.
type ListenImport struct {
	ID        *uuid.UUID `json:"id"`
	URL       *string    `json:"url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// Fully imported listen in the DB
type Listen struct {
	ID            *uuid.UUID `json:"id"`
	URL           *string    `json:"url"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	MusicBrainzID *string    `json:"musicbrainz_id"`
}
