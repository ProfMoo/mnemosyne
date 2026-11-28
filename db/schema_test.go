package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/db"
)

type SchemaSuite struct {
	suite.Suite
}

func TestSchemaSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SchemaSuite))
}

func (s *SchemaSuite) TestOpenInitializesSchemaDirectory() {
	database, err := db.Open(s.T().Context(), ":memory:")
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(database.Close()) })

	var message string
	err = database.QueryRowContext(s.T().Context(), "SELECT message FROM greetings WHERE id = 1").Scan(&message)
	s.Require().NoError(err)
	s.Equal("Hello, World!", message)

	const (
		uuidBytes        = 16
		expectedDuration = 180
		epochSeconds     = 1767225600
		subsecond        = 0
		title            = "Example track"
		emptyArray       = "[]"
	)
	created := time.Unix(epochSeconds, subsecond).UTC()
	updated := created.Add(time.Hour)
	id := make([]byte, uuidBytes)
	mbid := make([]byte, uuidBytes)
	_, err = database.ExecContext(
		s.T().Context(),
		`INSERT INTO listens (id, created_at, updated_at, duration, title, artist_credits, albums, mbid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		created,
		updated,
		expectedDuration,
		title,
		emptyArray,
		emptyArray,
		mbid,
	)
	s.Require().NoError(err)

	var duration int
	var artistCredits, albums, actualTitle string
	var actualCreated, actualUpdated time.Time
	err = database.QueryRowContext(
		s.T().Context(),
		"SELECT duration, artist_credits, albums, title, created_at, updated_at FROM listens WHERE id = ?",
		id,
	).Scan(&duration, &artistCredits, &albums, &actualTitle, &actualCreated, &actualUpdated)
	s.Require().NoError(err)
	s.Equal(expectedDuration, duration)
	s.Equal(emptyArray, artistCredits)
	s.Equal(emptyArray, albums)
	s.Equal(title, actualTitle)
	s.Equal(created, actualCreated)
	s.Equal(updated, actualUpdated)
}

func (s *SchemaSuite) TestListenSchemaRequiresExplicitValues() {
	database, err := db.Open(s.T().Context(), ":memory:")
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(database.Close()) })

	var defaultCount int
	err = database.QueryRowContext(
		s.T().Context(),
		"SELECT COUNT(*) FROM pragma_table_info('listens') WHERE dflt_value IS NOT NULL",
	).Scan(&defaultCount)
	s.Require().NoError(err)
	s.Zero(defaultCount)

	const uuidBytes = 16
	_, err = database.ExecContext(
		s.T().Context(),
		"INSERT INTO listens (id, duration, mbid) VALUES (?, ?, ?)",
		make([]byte, uuidBytes),
		time.Minute,
		make([]byte, uuidBytes),
	)
	s.Require().Error(err)
	s.Contains(err.Error(), "NOT NULL constraint failed")
}
