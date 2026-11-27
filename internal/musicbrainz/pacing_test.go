//revive:disable:add-constant // Fixture limits and time steps are explicit test inputs.
package musicbrainz_test

import (
	"context"
	"net/http"
	"testing"
	"testing/fstest"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/matching"
	"github.com/profmoo/mnemosyne/internal/musicbrainz"
)

type PacingSuite struct {
	suite.Suite
}

func TestPacing(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PacingSuite))
}

func (s *PacingSuite) TestSharedGateAndCanceledWait() {
	synctest.Test(s.T(), func(t *testing.T) {
		files := fstest.MapFS{
			"ws/2/recording/index.html": &fstest.MapFile{Data: []byte(`{"recordings":[]}`)},
		}
		client := musicbrainz.NewClient(musicbrainz.NewClientParams{
			Config: config.MusicBrainzConfig{
				Endpoint: "https://musicbrainz.example", UserAgent: "pacing-test",
				RequestInterval: time.Second, RequestTimeout: 10 * time.Second,
			},
			HTTPClient: &http.Client{Transport: http.NewFileTransportFS(files)},
		})
		query := matching.Search{Artist: "Artist", Title: "Song", Limit: 10}
		_, err := client.SearchRecordings(t.Context(), query)
		s.Require().NoError(err)

		completed := make(chan error, 1)
		go func() {
			_, searchErr := client.SearchRecordings(t.Context(), query)
			completed <- searchErr
		}()
		synctest.Wait()
		s.Empty(completed, "second request must wait for its slot")

		ctx, cancel := context.WithCancel(t.Context())
		canceled := make(chan error, 1)
		go func() {
			_, searchErr := client.SearchRecordings(ctx, query)
			canceled <- searchErr
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		s.Require().ErrorIs(<-canceled, context.Canceled)
		s.Empty(completed)

		time.Sleep(time.Second)
		synctest.Wait()
		s.NoError(<-completed)
	})
}
