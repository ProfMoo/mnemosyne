//revive:disable:add-constant // Fixture values describe the configuration being loaded.
package config_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
)

const musicBrainzSettings = `
musicBrainz:
  endpoint: https://musicbrainz.org
  userAgent: mnemosyne/test
  requestTimeout: 10s
  requestInterval: 1s
`

type MusicBrainzSuite struct {
	suite.Suite
}

func TestMusicBrainz(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MusicBrainzSuite))
}

func (s *MusicBrainzSuite) load(settings string) (*config.Config, error) {
	s.T().Helper()
	files := fstest.MapFS{
		configPath: {Data: []byte("bindAddress: ':8080'\nfrontendDir: bin/web\napiPrefix: /api/\n" + settings)},
	}

	loaded, err := config.LoadConfig(files, configPath)
	if err != nil {
		return nil, fmt.Errorf("load test configuration: %w", err)
	}

	return loaded, nil
}

func (s *MusicBrainzSuite) TestPublicSettings() {
	settings, err := s.load(musicBrainzSettings)
	s.Require().NoError(err)
	s.Equal("https://musicbrainz.org", settings.MusicBrainz.Endpoint)
	s.Equal("mnemosyne/test", settings.MusicBrainz.UserAgent)
	s.Equal(10*time.Second, settings.MusicBrainz.RequestTimeout)
	s.Equal(time.Second, settings.MusicBrainz.RequestInterval)
}

func (s *MusicBrainzSuite) TestSelfHostedClientUsesLoadedSettings() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/music/ws/2/recording/", r.URL.Path)
		s.Empty(r.Header.Get("Authorization"))
		s.Equal("mnemosyne/test", r.UserAgent())
		_, err := fmt.Fprint(w, `{"recordings":[]}`)
		s.NoError(err)
	}))
	s.T().Cleanup(server.Close)
	yaml := strings.NewReplacer(
		"https://musicbrainz.org", server.URL+"/music",
		"requestInterval: 1s", "requestInterval: 5ms",
	).Replace(musicBrainzSettings)
	settings, err := s.load(yaml)
	s.Require().NoError(err)
	s.Equal(5*time.Millisecond, settings.MusicBrainz.RequestInterval)

	client := musicbrainz.NewClient(musicbrainz.NewClientParams{
		Config: settings.MusicBrainz, HTTPClient: server.Client(),
	})
	recordings, err := client.SearchRecordings(s.T().Context(), matching.Search{
		Artist: "Artist", Title: "Song", Limit: 10,
	})
	s.Require().NoError(err)
	s.Empty(recordings)
}

func (s *MusicBrainzSuite) TestInvalidSettings() {
	for _, invalid := range []string{
		"",
		strings.ReplaceAll(musicBrainzSettings, "https://musicbrainz.org", "not-a-url"),
		strings.ReplaceAll(musicBrainzSettings, "mnemosyne/test", `""`),
		strings.ReplaceAll(musicBrainzSettings, "10s", "0s"),
		strings.ReplaceAll(musicBrainzSettings, "10s", "not-a-duration"),
		strings.ReplaceAll(musicBrainzSettings, "requestInterval: 1s", "requestInterval: -1s"),
		strings.ReplaceAll(musicBrainzSettings, "endpoint:", "endpont:"),
	} {
		_, err := s.load(invalid)
		s.Error(err)
	}
}

func (s *MusicBrainzSuite) TestZeroIntervalForSelfHostedServer() {
	yaml := strings.NewReplacer(
		"https://musicbrainz.org", "http://musicbrainz.internal:5000",
		"requestInterval: 1s", "requestInterval: 0s",
	).Replace(musicBrainzSettings)
	settings, err := s.load(yaml)
	s.Require().NoError(err)
	s.Zero(settings.MusicBrainz.RequestInterval)
}

func (s *MusicBrainzSuite) TestEmbeddedApplicationConfigurations() {
	for _, path := range []string{"config/local.yaml", "config/docker.yaml"} {
		settings, err := config.LoadConfig(os.DirFS("../../cmd/mnemosyne"), path)
		s.Require().NoError(err)
		s.Equal("https://musicbrainz.org", settings.MusicBrainz.Endpoint)
		s.Equal(time.Second, settings.MusicBrainz.RequestInterval)
	}
}
