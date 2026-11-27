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

type SubmitHandler struct {
	matcher          matching.Matcher
	listenRepository ListenRepository
	timeSource       func() time.Time
}

type ListenRepository interface {
	Create(ctx context.Context, listen *listens.Listen) error
}

func NewSubmitHandler(params SubmitHandlerParams) *SubmitHandler {
	return &SubmitHandler{
		matcher:          params.Matcher,
		listenRepository: params.ListenRepository,
		timeSource:       params.TimeSource,
	}
}

type SubmitHandlerParams struct {
	fx.In

	Matcher          matching.Matcher
	ListenRepository ListenRepository
	TimeSource       func() time.Time
}

func (h *SubmitHandler) Handle(
	ctx context.Context,
	req api.SubmitListensRequestObject,
) (api.SubmitListensResponseObject, error) {
	listen, err := convertAPIListenToInternalListen(req.Body)
	if err != nil {
		return nil, err
	}

	input := matching.Input{
		Artist: *listen.ArtistName,
		Title:  *listen.TrackName,
	}

	match, err := h.matcher.Match(ctx, input)
	if err != nil {
		return nil, err
	}

	fullListen, err := h.constructFullListen(&match, listen)
	if err != nil {
		return nil, err
	}

	err = h.listenRepository.Create(ctx, fullListen)
	if err != nil {
		return nil, err
	}

	// 1. convert to internal representation of listen
	// 2. use the mbidmapper package to return the mbid and confidence level.
	// 3. insert the listen into the database.

	return api.SubmitListens201JSONResponse{
		Status: "ok",
	}, nil
}

func convertAPIListenToInternalListen(listen *api.SubmitListensJSONRequestBody) (*listens.ListenImport, error) {
	body := api.SubmitListensJSONBody(*listen)

	submission, err := body.AsSingleSubmission()
	if err != nil {
		return nil, fmt.Errorf("decode single submission: %w", err)
	}

	if len(submission.Payload) != 1 {
		return nil, errors.New("single submission requires one listen")
	}

	item := submission.Payload[0]
	metadata := item.TrackMetadata
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
) (*listens.Listen, error) {
	now := h.timeSource()
	bestCandidate := match.Candidates[0]
	return &listens.Listen{
		ID:        listen.ID,
		CreatedAt: &now,
		UpdatedAt: &now,
		Recording: &listens.Recording{
			Title:        bestCandidate.Recording.Title,
			ArtistCredit: convertMatchingArtistCreditToInternalArtistCredit(bestCandidate.Recording.ArtistCredit),
			Albums:       convertMatchingAlbumsToInternalAlbums(bestCandidate.Recording.Albums),
			MBID:         bestCandidate.Recording.MBID,
			Duration:     bestCandidate.Recording.Duration,
		},
	}, nil
}

func convertMatchingArtistCreditToInternalArtistCredit(artistCredit []matching.ArtistCredit) []listens.ArtistCredit {
	internalArtistCredit := make([]listens.ArtistCredit, len(artistCredit))
	for i, credit := range artistCredit {
		internalArtistCredit[i] = listens.ArtistCredit{
			Name: credit.Name,
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
			Title:        album.Title,
			ArtistCredit: convertMatchingArtistCreditToInternalArtistCredit(album.ArtistCredit),
			MBID:         album.MBID,
		}
	}
	return internalAlbums
}
