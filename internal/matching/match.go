// Package matching identifies recordings independently of the catalog backend.
package matching

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidInput identifies missing artist/title metadata or a negative duration.
var ErrInvalidInput = errors.New("artist and title are required; duration must not be negative")

// Matcher resolves metadata to recording candidates.
type Matcher interface {
	Match(context.Context, Input) (Result, error)
}

// Input describes a listen.
type Input struct {
	// Required; cannot be empty or whitespace-only.
	Artist string
	// Required; cannot be empty or whitespace-only.
	Title string
	// Optional; empty means unknown.
	Album string
	// Optional; zero means unknown. Negative values are invalid.
	Duration time.Duration
}

// Status describes the recording decision, not release identity.
type Status string

// Recording decisions distinguish accepted, uncertain, and empty searches.
const (
	StatusMatched   Status = "matched"
	StatusAmbiguous Status = "ambiguous"
	StatusUnmatched Status = "unmatched"
)

// Result contains candidates ordered by acceptance eligibility and then score.
type Result struct {
	Status Status
	// For matched results, the first candidate is the accepted recording.
	Candidates []Candidate
}

// Candidate pairs a recording with its metadata evidence score.
type Candidate struct {
	Recording Recording
	// Score ranks evidence from zero to one; it is not a probability.
	Score float64
}

// Catalog retrieves candidates without making matching decisions.
type Catalog interface {
	SearchRecordings(context.Context, Search) ([]Recording, error)
}

// Search contains backend-independent retrieval constraints.
type Search struct {
	Artist string
	Title  string
	Album  string
	Limit  int
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
