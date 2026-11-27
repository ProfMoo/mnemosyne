package musicbrainz

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/profmoo/mnemosyne/internal/matching"
)

var errInvalidRecording = errors.New("invalid MusicBrainz recording data")

type recordingResponse struct {
	Recordings []recording `json:"recordings"`
}

type recording struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	ArtistCredit []artistCredit `json:"artist-credit"`
	Releases     []release      `json:"releases"`
	Length       int64          `json:"length"`
}

type artistCredit struct {
	Name       string `json:"name"`
	JoinPhrase string `json:"joinphrase"`
	Artist     artist `json:"artist"`
}

type artist struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type release struct {
	Title        string         `json:"title"`
	ID           string         `json:"id"`
	ArtistCredit []artistCredit `json:"artist-credit"`
}

func (r recordingResponse) recordings() ([]matching.Recording, error) {
	const empty = 0

	result := make([]matching.Recording, empty, len(r.Recordings))
	for i := range r.Recordings {
		converted, err := r.Recordings[i].convert()
		if err != nil {
			return nil, err
		}

		result = append(result, converted)
	}

	return result, nil
}

func (r *recording) convert() (matching.Recording, error) {
	const (
		maxDurationMS = int64((1<<63 - 1) / time.Millisecond)
		unknown       = 0
	)

	id, err := parseMBID(r.ID)
	if err != nil {
		return matching.Recording{}, err
	}
	if strings.TrimSpace(r.Title) == "" || len(r.ArtistCredit) == unknown ||
		r.Length < unknown || r.Length > maxDurationMS {
		return matching.Recording{}, errInvalidRecording
	}

	credits, err := convertCredits(r.ArtistCredit)
	if err != nil {
		return matching.Recording{}, err
	}

	var albums []matching.Album
	for i := range r.Releases {
		album, convertErr := r.Releases[i].convert()
		if convertErr != nil {
			return matching.Recording{}, convertErr
		}
		albums = append(albums, album)
	}

	return matching.Recording{
		MBID: id, Title: r.Title, ArtistCredit: credits,
		Duration: time.Duration(r.Length) * time.Millisecond, Albums: albums,
	}, nil
}

func (r *release) convert() (matching.Album, error) {
	id, err := parseMBID(r.ID)
	if err != nil {
		return matching.Album{}, err
	}
	if strings.TrimSpace(r.Title) == "" {
		return matching.Album{}, errInvalidRecording
	}

	credits, err := convertCredits(r.ArtistCredit)
	if err != nil {
		return matching.Album{}, err
	}

	return matching.Album{MBID: id, Title: r.Title, ArtistCredit: credits}, nil
}

func convertCredits(credits []artistCredit) ([]matching.ArtistCredit, error) {
	var result []matching.ArtistCredit
	for _, credit := range credits {
		id, err := parseMBID(credit.Artist.ID)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(credit.Artist.Name) == "" {
			return nil, errInvalidRecording
		}

		name := credit.Name
		if name == "" {
			name = credit.Artist.Name
		}
		result = append(result, matching.ArtistCredit{
			Name: name, JoinPhrase: credit.JoinPhrase,
			Artists: matching.Artist{Name: credit.Artist.Name, MBID: id},
		})
	}

	return result, nil
}

func parseMBID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("decode MusicBrainz entity ID: %w", errInvalidRecording)
	}

	return id, nil
}
