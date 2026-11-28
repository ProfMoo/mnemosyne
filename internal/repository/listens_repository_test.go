package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/internal/listens"
	"github.com/profmoo/mnemosyne/internal/repository"
)

type ListenRepositorySuite struct {
	suite.Suite

	database   *sql.DB
	repository *repository.SQLiteListenRepository
}

func TestListenRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ListenRepositorySuite))
}

func (s *ListenRepositorySuite) SetupTest() {
	database, err := db.Open(s.T().Context(), ":memory:")
	s.Require().NoError(err)
	s.database = database
	s.repository = repository.NewSqliteListenRepository(repository.NewSqliteListenRepositoryParams{DB: database})
	s.T().Cleanup(func() { s.NoError(database.Close()) })
}

func (s *ListenRepositorySuite) TestRoundTrip() {
	expected := repositoryListen(uuid.New())
	s.Require().NoError(s.repository.Create(s.T().Context(), expected))
	actual, err := s.repository.Get(s.T().Context(), *expected.ID)
	s.Require().NoError(err)
	s.Equal(expected, actual)
	s.Require().Error(s.repository.Create(s.T().Context(), expected))
}

func (s *ListenRepositorySuite) TestEmptyListAndMissingListen() {
	items, err := s.repository.List(s.T().Context())
	s.Require().NoError(err)
	s.Empty(items)
	s.NotNil(items)
	_, err = s.repository.Get(s.T().Context(), uuid.New())
	s.Require().ErrorIs(err, listens.ErrNotFound)
}

func (s *ListenRepositorySuite) TestOrderAndNilCollections() {
	const (
		rowCount    = 3
		firstIndex  = 0
		secondIndex = 1
		thirdIndex  = 2
	)
	first := repositoryListen(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	second := repositoryListen(uuid.MustParse("00000000-0000-0000-0000-000000000002"))
	third := repositoryListen(uuid.MustParse("00000000-0000-0000-0000-000000000003"))
	newest := third.CreatedAt.Add(time.Hour)
	third.CreatedAt = &newest
	third.Recording.ArtistCredits = nil
	third.Recording.Albums = nil
	for _, item := range []*listens.Listen{second, first, third} {
		s.Require().NoError(s.repository.Create(s.T().Context(), item))
	}
	items, err := s.repository.List(s.T().Context())
	s.Require().NoError(err)
	s.Require().Len(items, rowCount)
	s.Equal(*third.ID, *items[firstIndex].ID)
	s.Equal(*second.ID, *items[secondIndex].ID)
	s.Equal(*first.ID, *items[thirdIndex].ID)
	s.NotNil(items[firstIndex].Recording.ArtistCredits)
	s.NotNil(items[firstIndex].Recording.Albums)
}

func (s *ListenRepositorySuite) TestStorageFailures() {
	item := repositoryListen(uuid.New())
	s.Require().NoError(s.repository.Create(s.T().Context(), item))
	_, err := s.database.ExecContext(s.T().Context(), "UPDATE listens SET artist_credits = '[1]'")
	s.Require().NoError(err)
	_, err = s.repository.Get(s.T().Context(), *item.ID)
	s.Require().ErrorContains(err, "decode artist credits")
	_, err = s.repository.List(s.T().Context())
	s.Require().ErrorContains(err, "decode artist credits")
}

func (s *ListenRepositorySuite) TestValidationAndCancellation() {
	s.Require().ErrorIs(s.repository.Create(s.T().Context(), nil), listens.ErrInvalidListen)
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err := s.repository.List(ctx)
	s.Require().ErrorIs(err, context.Canceled)
	_, err = s.repository.Get(ctx, uuid.New())
	s.Require().ErrorIs(err, context.Canceled)
	s.Require().ErrorIs(s.repository.Create(ctx, repositoryListen(uuid.New())), context.Canceled)
}

func repositoryListen(id uuid.UUID) *listens.Listen {
	const (
		duration     = 3 * time.Minute
		epochSeconds = 1767225600
		subsecond    = 0
	)
	created := time.Unix(epochSeconds, subsecond).UTC()
	updated := created.Add(time.Hour)
	credit := listens.ArtistCredit{
		Name:       "Credited artist",
		JoinPhrase: " feat. ",
		Artists:    listens.Artist{Name: "Canonical artist", MBID: uuid.New()},
	}
	return &listens.Listen{
		ID: &id, CreatedAt: &created, UpdatedAt: &updated,
		Recording: &listens.Recording{
			Title:         "Song",
			Duration:      duration,
			MBID:          uuid.New(),
			ArtistCredits: []listens.ArtistCredit{credit},
			Albums: []listens.Album{
				{Title: "Album", ArtistCredits: []listens.ArtistCredit{credit}, MBID: uuid.New()},
			},
		},
	}
}
