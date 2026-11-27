//revive:disable:add-constant // Inline numeric values describe individual test fixtures.
package musicbrainz_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
)

type ClientSuite struct {
	suite.Suite
}

const (
	testArtist            = "A"
	testTitle             = "B"
	testUserAgent         = "mnemosyne/test (https://example.org)"
	testAlbum             = "Album"
	collaborationResponse = `{"recordings":[{
		"id":"11111111-1111-4111-8111-111111111111","title":"Song","length":123456,
		"artist-credit":[
			{"name":"A","joinphrase":" feat. ",
			 "artist":{"id":"22222222-2222-4222-8222-222222222222","name":"Artist A"}},
			{"name":"B","joinphrase":" & ",
			 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist B"}},
			{"artist":{"id":"44444444-4444-4444-8444-444444444444","name":"Artist C"}}
		],
		"releases":[
			{"id":"55555555-5555-4555-8555-555555555555","title":"Album","artist-credit":[
				{"name":"Various Artists",
				 "artist":{"id":"89ad4ac3-39f7-470e-963a-56509c546377","name":"Various Artists"}}
			]},
			{"id":"66666666-6666-4666-8666-666666666666","title":"Album"}
		],"score":100
	}]}`
)

func TestClient(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) client(handler http.HandlerFunc) *musicbrainz.Client {
	s.T().Helper()
	server := httptest.NewServer(handler)
	s.T().Cleanup(server.Close)

	return musicbrainz.NewClient(musicbrainz.NewClientParams{
		Config: config.MusicBrainzConfig{
			Endpoint: server.URL, UserAgent: testUserAgent,
			RequestTimeout: time.Second,
		},
		HTTPClient: server.Client(),
	})
}

func (s *ClientSuite) TestQueryAndConversion() {
	client := s.client(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/ws/2/recording/", r.URL.Path)
		s.Equal("application/json", r.Header.Get("Accept"))
		s.Empty(r.Header.Get("Authorization"))
		s.Equal(testUserAgent, r.UserAgent())
		s.Equal("json", r.URL.Query().Get("fmt"))
		s.Equal("10", r.URL.Query().Get("limit"))
		s.Equal(`artist:"A \& B" AND recording:"Song \"One\"" AND release:"Album"`, r.URL.Query().Get("query"))
		_, err := fmt.Fprint(w, collaborationResponse)
		s.NoError(err)
	})
	recordings, err := client.SearchRecordings(s.T().Context(), matching.Search{
		Artist: "A & B", Title: `Song "One"`, Album: testAlbum, Limit: 10,
	})
	s.Require().NoError(err)
	s.Require().Len(recordings, 1)
	s.Equal(matching.Recording{
		MBID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Title: "Song",
		Duration: 123456 * time.Millisecond,
		ArtistCredit: []matching.ArtistCredit{
			{Name: "A", JoinPhrase: " feat. ", Artists: matching.Artist{
				Name: "Artist A", MBID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
			}},
			{Name: "B", JoinPhrase: " & ", Artists: matching.Artist{
				Name: "Artist B", MBID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
			}},
			{Name: "Artist C", Artists: matching.Artist{
				Name: "Artist C", MBID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
			}},
		},
		Albums: []matching.Album{
			{
				Title: testAlbum, MBID: uuid.MustParse("55555555-5555-4555-8555-555555555555"),
				ArtistCredit: []matching.ArtistCredit{
					{Name: "Various Artists", Artists: matching.Artist{
						Name: "Various Artists", MBID: uuid.MustParse("89ad4ac3-39f7-470e-963a-56509c546377"),
					}},
				},
			},
			{Title: testAlbum, MBID: uuid.MustParse("66666666-6666-4666-8666-666666666666")},
		},
	}, recordings[0])
}

func (s *ClientSuite) TestHTTPFailure() {
	client := s.client(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	_, err := client.SearchRecordings(s.T().Context(), matching.Search{Artist: testArtist, Title: testTitle, Limit: 10})
	var upstream *musicbrainz.HTTPError
	s.Require().ErrorAs(err, &upstream)
	s.Equal(http.StatusServiceUnavailable, upstream.StatusCode)
	s.Equal("5", upstream.RetryAfter)
}

func (s *ClientSuite) TestInvalidResponses() {
	for _, body := range []string{
		`not json`,
		`{"recordings":[{"id":"bad"}]}`,
		`{"recordings":[{"id":"00000000-0000-0000-0000-000000000000"}]}`,
		strings.ReplaceAll(collaborationResponse, "22222222-2222-4222-8222-222222222222", "bad-artist-id"),
		strings.ReplaceAll(collaborationResponse, "55555555-5555-4555-8555-555555555555", "bad-release-id"),
		strings.ReplaceAll(collaborationResponse, "89ad4ac3-39f7-470e-963a-56509c546377", "bad-release-artist-id"),
		strings.Repeat(" ", (4<<20)+1),
	} {
		client := s.client(func(w http.ResponseWriter, _ *http.Request) {
			_, err := fmt.Fprint(w, body)
			s.NoError(err)
		})
		_, err := client.SearchRecordings(
			s.T().Context(),
			matching.Search{Artist: testArtist, Title: testTitle, Limit: 10},
		)
		s.Error(err)
	}
}

func (s *ClientSuite) TestInvalidSearch() {
	client := s.client(func(_ http.ResponseWriter, _ *http.Request) {
		s.Fail("invalid searches must not reach upstream")
	})
	_, err := client.SearchRecordings(s.T().Context(), matching.Search{Artist: testArtist, Title: testTitle})
	s.ErrorIs(err, musicbrainz.ErrInvalidSearch)
}

func (s *ClientSuite) TestCancellation() {
	client := s.client(func(_ http.ResponseWriter, _ *http.Request) {
		s.Fail("canceled searches must not reach upstream")
	})
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err := client.SearchRecordings(ctx, matching.Search{Artist: testArtist, Title: testTitle, Limit: 10})
	s.ErrorIs(err, context.Canceled)
}

func (s *ClientSuite) TestNoRedirect() {
	var requests atomic.Int32
	client := s.client(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "/other")
		w.WriteHeader(http.StatusFound)
	})
	_, err := client.SearchRecordings(s.T().Context(), matching.Search{Artist: testArtist, Title: testTitle, Limit: 10})
	s.Require().Error(err)
	s.Equal(int32(1), requests.Load())
}

func (s *ClientSuite) TestConfigValidation() {
	validate := validator.New()
	settings := config.MusicBrainzConfig{
		Endpoint: "https://musicbrainz.org", UserAgent: testUserAgent,
		RequestTimeout: 10 * time.Second, RequestInterval: time.Second,
	}
	s.Require().NoError(validate.Struct(settings))
	settings.RequestInterval = -time.Millisecond
	s.Require().Error(validate.Struct(settings))
	settings.RequestInterval = time.Second
	settings.Endpoint = "not a URL"
	s.Error(validate.Struct(settings))
}
