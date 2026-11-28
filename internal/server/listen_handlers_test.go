package server_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/listens"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
	"github.com/profmoo/mnemosyne/internal/repository"
	"github.com/profmoo/mnemosyne/internal/server"
)

const (
	listensPath    = "/api/listens"
	listenPath     = "/api/listens/"
	submitPath     = "/api/1/submit-listens"
	submissionBody = `{"listen_type":"single","payload":[{"listened_at":1767225600,
		"track_metadata":{"artist_name":"A feat. B","track_name":"Song"}}]}`
	matchedRecording = `{"recordings":[{
		"id":"11111111-1111-4111-8111-111111111111","title":"Song","length":180000,
		"artist-credit":[
			{"name":"A","joinphrase":" feat. ","artist":{"id":"22222222-2222-4222-8222-222222222222","name":"A"}},
			{"name":"B","artist":{"id":"33333333-3333-4333-8333-333333333333","name":"B"}}
		],
		"releases":[{"id":"44444444-4444-4444-8444-444444444444","title":"Album","artist-credit":[
			{"name":"A","artist":{"id":"22222222-2222-4222-8222-222222222222","name":"A"}}
		]}]
	}]}`
)

type ListenHandlersSuite struct {
	suite.Suite

	database     *sql.DB
	repository   *repository.SQLiteListenRepository
	handler      http.Handler
	upstreamBody string
}

func TestListenHandlersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ListenHandlersSuite))
}

func (s *ListenHandlersSuite) SetupTest() {
	database, err := db.Open(s.T().Context(), ":memory:")
	s.Require().NoError(err)
	s.database = database
	s.T().Cleanup(func() { s.NoError(database.Close()) })
	s.repository = repository.NewSqliteListenRepository(repository.NewSqliteListenRepositoryParams{DB: database})
	s.upstreamBody = matchedRecording
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, writeErr := fmt.Fprint(w, s.upstreamBody)
		s.NoError(writeErr)
	}))
	s.T().Cleanup(upstream.Close)
	client := musicbrainz.NewClient(musicbrainz.NewClientParams{
		Config: config.MusicBrainzConfig{
			Endpoint: upstream.URL, UserAgent: "mnemosyne/test", RequestTimeout: time.Second,
		},
		HTTPClient: upstream.Client(),
	})
	submit := server.NewSubmitHandler(server.SubmitHandlerParams{
		Matcher:          matching.NewService(matching.NewSimpleMatcherParams{Catalog: client}),
		ListenRepository: s.repository,
		TimeSource:       handlerTime,
	})
	app := server.NewServer(server.NewServerParams{
		Config: &config.Config{APIPrefix: "/api/", FrontendDir: s.T().TempDir()},
		Handler: server.NewHandler(server.NewHandlerParams{
			SubmitHandler:     submit,
			GetListenHandler:  server.NewGetListenHandler(server.GetListenHandlerParams{Repository: s.repository}),
			GetListensHandler: server.NewGetListensHandler(server.GetListensHandlerParams{Repository: s.repository}),
		}),
	})
	s.handler = app.Handler
}

func (s *ListenHandlersSuite) TestEmptyList() {
	response := s.request(http.MethodGet, listensPath, "")
	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(`{"listens":[]}`, response.Body.String())
}

func (s *ListenHandlersSuite) TestSubmitThenGetAndList() {
	const (
		duration    = 3 * time.Minute
		oneListen   = 1
		firstIndex  = 0
		creditCount = 2
		albumCount  = 1
	)
	submitted := s.request(http.MethodPost, submitPath, submissionBody)
	s.Require().Equal(http.StatusCreated, submitted.Code, submitted.Body.String())
	var result api.ListListensResponse
	response := s.request(http.MethodGet, listensPath, "")
	s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
	s.Require().NoError(json.Unmarshal(response.Body.Bytes(), &result))
	s.Require().Len(result.Listens, oneListen)
	listen := result.Listens[firstIndex]
	s.Equal(handlerTime(), listen.CreatedAt)
	s.Equal(handlerTime(), listen.UpdatedAt)
	s.Equal("Song", listen.Recording.Title)
	s.Equal(int64(duration), listen.Recording.Duration)
	s.Require().Len(listen.Recording.ArtistCredits, creditCount)
	s.Equal(" feat. ", listen.Recording.ArtistCredits[firstIndex].JoinPhrase)
	s.Require().Len(listen.Recording.Albums, albumCount)
	s.Equal("Album", listen.Recording.Albums[firstIndex].Title)
	s.NotEqual(uuid.Nil, listen.Id)

	var single api.Listen
	response = s.request(http.MethodGet, listenPath+listen.Id.String(), "")
	s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
	s.Require().NoError(json.Unmarshal(response.Body.Bytes(), &single))
	s.Equal(listen, single)
}

func (s *ListenHandlersSuite) TestNewestFirst() {
	const (
		count      = 3
		firstIndex = 0
		lastIndex  = 2
	)
	for index := range count {
		id := uuid.New()
		created := handlerTime().Add(time.Duration(index) * time.Hour)
		listen := &listens.Listen{
			ID: &id, CreatedAt: &created, UpdatedAt: &created,
			Recording: &listens.Recording{Title: fmt.Sprintf("Song %d", index), MBID: uuid.New()},
		}
		s.Require().NoError(s.repository.Create(s.T().Context(), listen))
	}
	response := s.request(http.MethodGet, listensPath, "")
	var result api.ListListensResponse
	s.Require().Equal(http.StatusOK, response.Code)
	s.Require().NoError(json.Unmarshal(response.Body.Bytes(), &result))
	s.Require().Len(result.Listens, count)
	s.Equal("Song 2", result.Listens[firstIndex].Recording.Title)
	s.Equal("Song 0", result.Listens[lastIndex].Recording.Title)
}

func (s *ListenHandlersSuite) TestMissingAndInvalidID() {
	response := s.request(http.MethodGet, listenPath+uuid.NewString(), "")
	s.Equal(http.StatusNotFound, response.Code)
	s.Contains(response.Body.String(), "NOT_FOUND")
	response = s.request(http.MethodGet, listenPath+"not-a-uuid", "")
	s.Equal(http.StatusBadRequest, response.Code)
}

func (s *ListenHandlersSuite) TestRepositoryFailure() {
	s.Require().NoError(s.database.Close())
	for _, path := range []string{listensPath, listenPath + uuid.NewString()} {
		response := s.request(http.MethodGet, path, "")
		s.Equal(http.StatusInternalServerError, response.Code)
		s.Contains(response.Body.String(), "INTERNAL_ERROR")
		s.NotContains(response.Body.String(), "database is closed")
	}
}

func (s *ListenHandlersSuite) TestInvalidAndUnmatchedSubmissions() {
	for _, body := range []string{"{}", `{"listen_type":"single","payload":[]}`, "null"} {
		response := s.request(http.MethodPost, submitPath, body)
		s.Equal(http.StatusBadRequest, response.Code, response.Body.String())
	}
	s.upstreamBody = `{"recordings":[]}`
	response := s.request(http.MethodPost, submitPath, submissionBody)
	s.Equal(http.StatusBadRequest, response.Code, response.Body.String())
	items, err := s.repository.List(s.T().Context())
	s.Require().NoError(err)
	s.Empty(items)
}

func (s *ListenHandlersSuite) request(method, path, body string) *httptest.ResponseRecorder {
	s.T().Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(s.T().Context(), method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	s.handler.ServeHTTP(response, request)
	return response
}

func handlerTime() time.Time {
	const (
		epochSeconds = 1767225600
		subsecond    = 0
	)
	return time.Unix(epochSeconds, subsecond).UTC()
}
