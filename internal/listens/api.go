// Package listens owns listen metadata, persistence contracts, and API mappings.
package listens

import (
	"time"

	"github.com/profmoo/mnemosyne/gen/api"
)

// ToAPI maps a domain listen to the public response without exposing domain types.
func ToAPI(listen *Listen) (api.Listen, error) {
	if err := listen.Validate(); err != nil {
		return api.Listen{}, err
	}
	recording := listen.Recording
	return api.Listen{
		Id:        *listen.ID,
		CreatedAt: *listen.CreatedAt,
		UpdatedAt: *listen.UpdatedAt,
		Recording: api.Recording{
			Title:         recording.Title,
			ArtistCredits: creditsToAPI(recording.ArtistCredits),
			Albums:        albumsToAPI(recording.Albums),
			Mbid:          recording.MBID,
			Duration:      int64(recording.Duration),
		},
	}, nil
}

func creditsToAPI(credits []ArtistCredit) []api.ArtistCredit {
	result := make([]api.ArtistCredit, len(credits))
	for index, credit := range credits {
		result[index] = api.ArtistCredit{
			Name:       credit.Name,
			JoinPhrase: credit.JoinPhrase,
			Artist:     api.Artist{Name: credit.Artists.Name, Mbid: credit.Artists.MBID},
		}
	}
	return result
}

func albumsToAPI(albums []Album) []api.Album {
	result := make([]api.Album, len(albums))
	for index, album := range albums {
		result[index] = api.Album{
			Title:         album.Title,
			ArtistCredits: creditsToAPI(album.ArtistCredits),
			Mbid:          album.MBID,
		}
	}
	return result
}

// FromAPI maps the public representation to independently owned domain values.
func FromAPI(value *api.Listen) (*Listen, error) {
	if value == nil {
		return nil, ErrInvalidListen
	}
	id, created, updated := value.Id, value.CreatedAt, value.UpdatedAt
	listen := &Listen{
		ID:        &id,
		CreatedAt: &created,
		UpdatedAt: &updated,
		Recording: &Recording{
			Title:         value.Recording.Title,
			ArtistCredits: creditsFromAPI(value.Recording.ArtistCredits),
			Albums:        albumsFromAPI(value.Recording.Albums),
			MBID:          value.Recording.Mbid,
			Duration:      time.Duration(value.Recording.Duration),
		},
	}
	if err := listen.Validate(); err != nil {
		return nil, err
	}
	return listen, nil
}

func creditsFromAPI(credits []api.ArtistCredit) []ArtistCredit {
	result := make([]ArtistCredit, len(credits))
	for index, credit := range credits {
		result[index] = ArtistCredit{
			Name:       credit.Name,
			JoinPhrase: credit.JoinPhrase,
			Artists:    Artist{Name: credit.Artist.Name, MBID: credit.Artist.Mbid},
		}
	}
	return result
}

func albumsFromAPI(albums []api.Album) []Album {
	result := make([]Album, len(albums))
	for index, album := range albums {
		result[index] = Album{
			Title:         album.Title,
			ArtistCredits: creditsFromAPI(album.ArtistCredits),
			MBID:          album.Mbid,
		}
	}
	return result
}
