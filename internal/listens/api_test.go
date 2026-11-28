package listens_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/listens"
)

type APISuite struct {
	suite.Suite
}

const creditedName = "Credited name"

func TestAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APISuite))
}

func (s *APISuite) TestRoundTrip() {
	const firstCredit = 0
	value := exampleAPIListen()
	domain, err := listens.FromAPI(&value)
	s.Require().NoError(err)
	s.Equal(time.Duration(value.Recording.Duration), domain.Recording.Duration)
	s.Equal(" feat. ", domain.Recording.ArtistCredits[firstCredit].JoinPhrase)
	s.Equal(
		value.Recording.ArtistCredits[firstCredit].Artist.Mbid,
		domain.Recording.ArtistCredits[firstCredit].Artists.MBID,
	)

	result, err := listens.ToAPI(domain)
	s.Require().NoError(err)
	s.Equal(value, result)
	domain.Recording.ArtistCredits[firstCredit].Name = "Changed"
	s.Equal(creditedName, value.Recording.ArtistCredits[firstCredit].Name)
	s.Equal(creditedName, result.Recording.ArtistCredits[firstCredit].Name)
}

func (s *APISuite) TestEmptyCollectionsAreArrays() {
	value := exampleAPIListen()
	value.Recording.ArtistCredits = nil
	value.Recording.Albums = nil
	domain, err := listens.FromAPI(&value)
	s.Require().NoError(err)
	result, err := listens.ToAPI(domain)
	s.Require().NoError(err)
	encoded, err := json.Marshal(result)
	s.Require().NoError(err)
	s.Contains(string(encoded), `"artistCredits":[]`)
	s.Contains(string(encoded), `"albums":[]`)
}

func (s *APISuite) TestInvalidListen() {
	_, err := listens.ToAPI(nil)
	s.Require().ErrorIs(err, listens.ErrInvalidListen)
	_, err = listens.FromAPI(nil)
	s.Require().ErrorIs(err, listens.ErrInvalidListen)
	_, err = listens.ToAPI(&listens.Listen{})
	s.Require().ErrorIs(err, listens.ErrInvalidListen)
	const invalidDuration = -1
	value := exampleAPIListen()
	value.Recording.Duration = invalidDuration
	_, err = listens.FromAPI(&value)
	s.Require().ErrorIs(err, listens.ErrInvalidListen)
}

func exampleAPIListen() api.Listen {
	const (
		duration     = 3 * time.Minute
		epochSeconds = 1767225600
		subsecond    = 0
	)
	created := time.Unix(epochSeconds, subsecond).UTC()
	artist := api.Artist{Name: "Canonical name", Mbid: uuid.MustParse("00000000-0000-0000-0000-000000000002")}
	credits := []api.ArtistCredit{{Name: creditedName, JoinPhrase: " feat. ", Artist: artist}}
	return api.Listen{
		Id:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		CreatedAt: created,
		UpdatedAt: created.Add(time.Hour),
		Recording: api.Recording{
			Title:         "Track title",
			Duration:      int64(duration),
			Mbid:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			ArtistCredits: credits,
			Albums: []api.Album{
				{Title: "Album one", ArtistCredits: credits, Mbid: artist.Mbid},
				{Title: "Album two", ArtistCredits: []api.ArtistCredit{}, Mbid: artist.Mbid},
			},
		},
	}
}
