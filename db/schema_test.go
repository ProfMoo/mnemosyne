package db_test

import (
	"testing"

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
	)
	id := make([]byte, uuidBytes)
	mbid := make([]byte, uuidBytes)
	_, err = database.ExecContext(
		s.T().Context(),
		"INSERT INTO listens (id, duration, mbid) VALUES (?, ?, ?)",
		id,
		expectedDuration,
		mbid,
	)
	s.Require().NoError(err)

	var duration int
	var artistCredits, albums string
	err = database.QueryRowContext(
		s.T().Context(),
		"SELECT duration, artist_credits, albums FROM listens WHERE id = ?",
		id,
	).Scan(&duration, &artistCredits, &albums)
	s.Require().NoError(err)
	s.Equal(expectedDuration, duration)
	s.Equal("[]", artistCredits)
	s.Equal("[]", albums)
}
