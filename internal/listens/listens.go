package listens

import (
	"time"

	"github.com/google/uuid"
)

// NOTE: Not sure if I want to validate these fields anywhere :thinking:. Probably should validate during OpenAPI validation.

// The absolute minimal amount of information that we need to accept a listen from ANY source, with an ID attached.
type ListenImport struct {
	ID         *uuid.UUID `json:"id"`
	ArtistName *string    `json:"artist_name"`
	TrackName  *string    `json:"track_name"`
}

// Fully imported listen in the DB
type Listen struct {
	ID        *uuid.UUID `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Recording *Recording `json:"recording"`
}

// Recording identifies audio independently of any particular release edition.
type Recording struct {
	Title        string
	ArtistCredit []ArtistCredit
	Albums       []Album
	MBID         uuid.UUID
	Duration     time.Duration
}

// ArtistCredit is slightly different from Artist. An Artist will be the canonical name and MBID,
// but ArtistCredit can match what the music artist chose to credit themselves for that track.
// Ex: Amelia Stone might credit herself as "A. Stone" for a particular track.
type ArtistCredit struct {
	Name string
	// Empty if this is the last artist on the track.
	// If there are artists after this in the list, the join phrase is used to combine the names correctly.
	// Ex: "A. Stone feat. Joe and Jane"
	JoinPhrase string
	Artists    Artist
}

type Album struct {
	Title        string
	ArtistCredit []ArtistCredit
	// Will be a release ID.
	MBID uuid.UUID
}

type Artist struct {
	Name string
	MBID uuid.UUID
}

type MBIDResult struct {
	MBID       *string `json:"mbid"`
	Confidence *int    `json:"confidence"`
}
