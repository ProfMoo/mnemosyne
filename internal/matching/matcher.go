package matching

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/fx"
)

// SimpleMatcher applies recording matching policy to catalog search results.
type SimpleMatcher struct {
	catalog Catalog
}

// NewSimpleMatcherParams supplies the candidate catalog.
type NewSimpleMatcherParams struct {
	fx.In

	Catalog Catalog
}

// NewService constructs a backend-independent recording matcher.
func NewService(params NewSimpleMatcherParams) *SimpleMatcher {
	return &SimpleMatcher{catalog: params.Catalog}
}

// Match accepts only a unique recording with exact normalized artist and title.
func (s *SimpleMatcher) Match(ctx context.Context, input Input) (Result, error) {
	const unknownDuration = 0

	input.Artist = strings.TrimSpace(input.Artist)
	input.Title = strings.TrimSpace(input.Title)
	input.Album = strings.TrimSpace(input.Album)
	if input.Artist == "" || input.Title == "" || input.Duration < unknownDuration {
		return Result{}, ErrInvalidInput
	}

	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("match recording: %w", err)
	}

	recordings, err := s.search(ctx, input)
	if err != nil {
		return Result{}, err
	}

	return rank(input, recordings), nil
}

func (s *SimpleMatcher) search(ctx context.Context, input Input) ([]Recording, error) {
	const (
		candidateLimit = 10
		noResults      = 0
	)

	query := Search{Artist: input.Artist, Title: input.Title, Album: input.Album, Limit: candidateLimit}
	recordings, err := s.catalog.SearchRecordings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search recordings: %w", err)
	}

	if len(recordings) == noResults && query.Album != "" {
		query.Album = ""
		recordings, err = s.catalog.SearchRecordings(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("search recordings without album: %w", err)
		}
	}

	return recordings, nil
}

type evaluation struct {
	candidate *Candidate
	eligible  bool
}

func rank(input Input, recordings []Recording) Result {
	const (
		oneMatch   = 1
		noMatches  = 0
		equalOrder = 0
	)

	byID := make(map[uuid.UUID]evaluation, len(recordings))
	for i := range recordings {
		recording := &recordings[i]
		if recording.MBID == uuid.Nil {
			continue
		}

		candidate, eligible := compare(input, recording)
		current := evaluation{candidate: &candidate, eligible: eligible}
		existing, found := byID[recording.MBID]
		if !found || orderEvaluation(current, existing) < equalOrder {
			byID[recording.MBID] = current
		}
	}

	var ranked []evaluation
	acceptable := noMatches
	for _, item := range byID {
		ranked = append(ranked, item)
		if item.eligible {
			acceptable++
		}
	}
	slices.SortFunc(ranked, orderEvaluation)

	result := Result{Status: StatusUnmatched}
	for _, item := range ranked {
		result.Candidates = append(result.Candidates, *item.candidate)
		result.Status = StatusAmbiguous
	}
	if acceptable == oneMatch {
		result.Status = StatusMatched
	}

	return result
}

func orderEvaluation(a, b evaluation) int {
	const (
		before     = -1
		after      = 1
		equalOrder = 0
	)

	if a.eligible != b.eligible {
		if a.eligible {
			return before
		}

		return after
	}
	if order := cmp.Compare(b.candidate.Score, a.candidate.Score); order != equalOrder {
		return order
	}

	return strings.Compare(a.candidate.Recording.MBID.String(), b.candidate.Recording.MBID.String())
}
