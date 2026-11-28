package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/listens"
	"github.com/profmoo/mnemosyne/internal/matching"
)

var errInvalidSubmission = errors.New("single submission requires one listen with artist and track names")

// SubmitHandler matches and persists submitted listens.
type SubmitHandler struct {
	matcher          matching.Matcher
	listenRepository listens.ListenCreator
	timeSource       func() time.Time
}

// NewSubmitHandler constructs the listen submission handler.
func NewSubmitHandler(params SubmitHandlerParams) *SubmitHandler {
	return &SubmitHandler{
		matcher:          params.Matcher,
		listenRepository: params.ListenRepository,
		timeSource:       params.TimeSource,
	}
}

// SubmitHandlerParams supplies matching, persistence, and the clock.
type SubmitHandlerParams struct {
	fx.In

	Matcher          matching.Matcher
	ListenRepository listens.ListenCreator
	TimeSource       func() time.Time
}

// Handle matches and saves a single submission.
func (h *SubmitHandler) Handle(
	ctx context.Context,
	req api.SubmitListensRequestObject,
) (api.SubmitListensResponseObject, error) {
	listen, err := convertAPIListenToInternalListen(req.Body)
	if errors.Is(err, errInvalidSubmission) {
		return api.SubmitListens400Response{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("decode submitted listen: %w", err)
	}

	input := matching.Input{
		Artist: *listen.ArtistName,
		Title:  *listen.TrackName,
	}

	match, err := h.matcher.Match(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("match submitted listen: %w", err)
	}

	const noCandidates = 0
	if match.Status != matching.StatusMatched || len(match.Candidates) == noCandidates {
		return api.SubmitListens400Response{}, nil
	}
	fullListen := h.constructFullListen(&match, listen)

	err = h.listenRepository.Create(ctx, fullListen)
	if err != nil {
		return nil, fmt.Errorf("save submitted listen: %w", err)
	}

	// 1. convert to internal representation of listen
	// 2. use the mbidmapper package to return the mbid and confidence level.
	// 3. insert the listen into the database.

	return api.SubmitListens201JSONResponse{
		Status: "ok",
	}, nil
}

func convertAPIListenToInternalListen(listen *api.SubmitListensJSONRequestBody) (*listens.ListenImport, error) {
	if listen == nil {
		return nil, errInvalidSubmission
	}
	body := *listen

	submission, err := body.AsSingleSubmission()
	if err != nil {
		return nil, fmt.Errorf("decode single submission: %w", errors.Join(errInvalidSubmission, err))
	}

	const singleListen = 1
	if submission.ListenType != "single" || len(submission.Payload) != singleListen {
		return nil, errInvalidSubmission
	}

	const firstListen = 0
	item := submission.Payload[firstListen]
	metadata := item.TrackMetadata
	if metadata.ArtistName == "" || metadata.TrackName == "" {
		return nil, errInvalidSubmission
	}
	id := uuid.New()

	return &listens.ListenImport{
		ID:         &id,
		ArtistName: &metadata.ArtistName,
		TrackName:  &metadata.TrackName,
	}, nil
}

func (h *SubmitHandler) constructFullListen(
	match *matching.Result,
	listen *listens.ListenImport,
) *listens.Listen {
	now := h.timeSource()
	const bestCandidateIndex = 0
	bestCandidate := match.Candidates[bestCandidateIndex]
	return &listens.Listen{
		ID:        listen.ID,
		CreatedAt: &now,
		UpdatedAt: &now,
		Recording: &listens.Recording{
			Title:         bestCandidate.Recording.Title,
			ArtistCredits: convertMatchingArtistCreditToInternalArtistCredit(bestCandidate.Recording.ArtistCredit),
			Albums:        convertMatchingAlbumsToInternalAlbums(bestCandidate.Recording.Albums),
			MBID:          bestCandidate.Recording.MBID,
			Duration:      bestCandidate.Recording.Duration,
		},
	}
}

func convertMatchingArtistCreditToInternalArtistCredit(artistCredit []matching.ArtistCredit) []listens.ArtistCredit {
	internalArtistCredit := make([]listens.ArtistCredit, len(artistCredit))
	for i, credit := range artistCredit {
		internalArtistCredit[i] = listens.ArtistCredit{
			Name:       credit.Name,
			JoinPhrase: credit.JoinPhrase,
			Artists: listens.Artist{
				Name: credit.Artists.Name,
				MBID: credit.Artists.MBID,
			},
		}
	}
	return internalArtistCredit
}

func convertMatchingAlbumsToInternalAlbums(albums []matching.Album) []listens.Album {
	internalAlbums := make([]listens.Album, len(albums))
	for i, album := range albums {
		internalAlbums[i] = listens.Album{
			Title:         album.Title,
			ArtistCredits: convertMatchingArtistCreditToInternalArtistCredit(album.ArtistCredit),
			MBID:          album.MBID,
		}
	}
	return internalAlbums
}
