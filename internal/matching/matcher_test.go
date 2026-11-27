//revive:disable:add-constant // Inline numeric values describe individual test fixtures.
package matching_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
)

type ServiceSuite struct {
	suite.Suite
}

const (
	testArtist      = "Artist"
	testTitle       = "Song"
	emptyRecordings = `{"recordings":[]}`
)

func TestService(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) matcher(payload string) *matching.SimpleMatcher {
	s.T().Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := fmt.Fprint(w, payload)
		s.NoError(err)
	}))
	s.T().Cleanup(server.Close)

	client := musicbrainz.NewClient(musicbrainz.NewClientParams{
		Config: config.MusicBrainzConfig{
			Endpoint: server.URL, UserAgent: "matching-test", RequestTimeout: time.Second,
		},
		HTTPClient: server.Client(),
	})

	return matching.NewService(matching.NewSimpleMatcherParams{Catalog: client})
}

func (s *ServiceSuite) TestUniqueRecording() {
	matcher := s.matcher(`{"recordings":[{
		"id":"11111111-1111-4111-8111-111111111111",
		"title":"Glory Box","artist-credit":[{"name":"Portishead",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Portishead"}}],"length":300000
	}]}`)
	result, err := matcher.Match(s.T().Context(), matching.Input{
		Artist: " PORTISHEAD ", Title: "glory   box", Duration: 302 * time.Second,
	})
	s.Require().NoError(err)
	s.Equal(matching.StatusMatched, result.Status)
	s.Require().Len(result.Candidates, 1)
	s.Equal(uuid.MustParse("11111111-1111-4111-8111-111111111111"), result.Candidates[0].Recording.MBID)
	s.InDelta(0.95, result.Candidates[0].Score, 0.0001)
}

func (s *ServiceSuite) TestAmbiguousRecordings() {
	matcher := s.matcher(`{"recordings":[
		{"id":"11111111-1111-4111-8111-111111111111","title":"Song","artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}]},
		{"id":"22222222-2222-4222-8222-222222222222","title":"Song","artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}]}
	]}`)
	result, err := matcher.Match(s.T().Context(), matching.Input{Artist: testArtist, Title: testTitle})
	s.Require().NoError(err)
	s.Equal(matching.StatusAmbiguous, result.Status)
	s.Len(result.Candidates, 2)
}

func (s *ServiceSuite) TestDuplicateRecording() {
	matcher := s.matcher(`{"recordings":[
		{"id":"11111111-1111-4111-8111-111111111111","title":"Song","artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}]},
		{"id":"11111111-1111-4111-8111-111111111111","title":"Song","artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}]}
	]}`)
	result, err := matcher.Match(s.T().Context(), matching.Input{Artist: testArtist, Title: testTitle})
	s.Require().NoError(err)
	s.Equal(matching.StatusMatched, result.Status)
	s.Len(result.Candidates, 1)
}

func (s *ServiceSuite) TestVersionAndDurationContradictions() {
	matcher := s.matcher(`{"recordings":[
		{"id":"11111111-1111-4111-8111-111111111111","title":"Song (Live)",
		 "artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}],"length":100000},
		{"id":"22222222-2222-4222-8222-222222222222","title":"Song",
		 "artist-credit":[{"name":"Artist",
		 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}],"length":150000}
	]}`)
	result, err := matcher.Match(s.T().Context(), matching.Input{
		Artist: testArtist, Title: testTitle, Duration: 100 * time.Second,
	})
	s.Require().NoError(err)
	s.Equal(matching.StatusAmbiguous, result.Status)
}

func (s *ServiceSuite) TestNoResults() {
	result, err := s.matcher(emptyRecordings).Match(s.T().Context(), matching.Input{
		Artist: testArtist, Title: testTitle, Album: "Album",
	})
	s.Require().NoError(err)
	s.Equal(matching.StatusUnmatched, result.Status)
	s.Empty(result.Candidates)
}

func (s *ServiceSuite) TestInvalidInput() {
	matcher := s.matcher(emptyRecordings)
	for _, input := range []matching.Input{
		{Artist: " ", Title: testTitle},
		{Artist: testArtist},
		{Artist: testArtist, Title: testTitle, Duration: -time.Second},
	} {
		_, err := matcher.Match(s.T().Context(), input)
		s.ErrorIs(err, matching.ErrInvalidInput)
	}
}

func (s *ServiceSuite) TestFailureIsNotUnmatched() {
	result, err := s.matcher(`not json`).Match(s.T().Context(), matching.Input{Artist: testArtist, Title: testTitle})
	s.Require().Error(err)
	s.Empty(result.Status)
}

func (s *ServiceSuite) TestCancellation() {
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err := s.matcher(emptyRecordings).Match(ctx, matching.Input{Artist: testArtist, Title: testTitle})
	s.ErrorIs(err, context.Canceled)
}

func (s *ServiceSuite) TestAlbumFallbackPreservesOriginalEvidence() {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		body := emptyRecordings
		if !strings.Contains(r.URL.Query().Get("query"), "release:") {
			body = `{"recordings":[{
				"id":"11111111-1111-4111-8111-111111111111","title":"Cafe\u0301",
				"artist-credit":[{"name":"Artist",
				 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist"}}],
				"releases":[{"id":"44444444-4444-4444-8444-444444444444","title":"Album"}]
			}]}`
		}
		_, err := fmt.Fprint(w, body)
		s.NoError(err)
	}))
	s.T().Cleanup(server.Close)
	client := musicbrainz.NewClient(musicbrainz.NewClientParams{
		Config: config.MusicBrainzConfig{
			Endpoint: server.URL, UserAgent: "matching-test", RequestTimeout: time.Second,
		},
		HTTPClient: server.Client(),
	})
	matcher := matching.NewService(matching.NewSimpleMatcherParams{Catalog: client})
	result, err := matcher.Match(s.T().Context(), matching.Input{
		Artist: testArtist, Title: "Café", Album: "Album",
	})
	s.Require().NoError(err)
	s.Equal(matching.StatusMatched, result.Status)
	s.Equal(int32(2), requests.Load())
	s.Require().Len(result.Candidates, 1)
	s.InDelta(0.95, result.Candidates[0].Score, 0.0001)
}

func (s *ServiceSuite) TestStructuredArtistMatching() {
	matcher := s.matcher(`{"recordings":[{
		"id":"11111111-1111-4111-8111-111111111111","title":"Song","artist-credit":[
			{"name":"A. Stone","joinphrase":" feat. ",
			 "artist":{"id":"22222222-2222-4222-8222-222222222222","name":"Amelia Stone"}},
			{"name":"JAY","joinphrase":" & ",
			 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Jay Morgan"}},
			{"name":"Northern Lights",
			 "artist":{"id":"44444444-4444-4444-8444-444444444444","name":"The Northern Lights"}}
		]
	}]}`)
	for _, artist := range []string{
		"A. Stone feat. JAY & Northern Lights",
		"Amelia Stone feat. Jay Morgan & The Northern Lights",
	} {
		result, err := matcher.Match(s.T().Context(), matching.Input{Artist: artist, Title: testTitle})
		s.Require().NoError(err)
		s.Equal(matching.StatusMatched, result.Status)
		s.Require().Len(result.Candidates, 1)
		credits := result.Candidates[0].Recording.ArtistCredit
		s.Require().Len(credits, 3)
		s.Equal("A. Stone", credits[0].Name)
		s.Equal("Amelia Stone", credits[0].Artists.Name)
		s.Equal(" feat. ", credits[0].JoinPhrase)
		s.Equal(" & ", credits[1].JoinPhrase)
	}
}

func (s *ServiceSuite) TestPartialAndReorderedArtistsStayAmbiguous() {
	matcher := s.matcher(`{"recordings":[{
		"id":"11111111-1111-4111-8111-111111111111","title":"Song","artist-credit":[
			{"name":"Artist A","joinphrase":" & ",
			 "artist":{"id":"22222222-2222-4222-8222-222222222222","name":"Artist A"}},
			{"name":"Artist B",
			 "artist":{"id":"33333333-3333-4333-8333-333333333333","name":"Artist B"}}
		]
	}]}`)
	for _, artist := range []string{"Artist A", "Artist B & Artist A"} {
		result, err := matcher.Match(s.T().Context(), matching.Input{Artist: artist, Title: testTitle})
		s.Require().NoError(err)
		s.Equal(matching.StatusAmbiguous, result.Status)
	}
}
