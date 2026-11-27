package musicbrainz

import (
	"fmt"
	"net/http"
	"time"
)

type pacedTransport struct {
	base     http.RoundTripper
	gate     chan struct{}
	next     time.Time
	interval time.Duration
}

func (t *pacedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	select {
	case t.gate <- struct{}{}:
	case <-req.Context().Done():
		return nil, fmt.Errorf("wait for MusicBrainz request slot: %w", req.Context().Err())
	}

	err := t.wait(req)
	<-t.gate
	if err != nil {
		return nil, err
	}

	response, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("send MusicBrainz request: %w", err)
	}

	return response, nil
}

func (t *pacedTransport) wait(req *http.Request) error {
	timer := time.NewTimer(time.Until(t.next))
	defer timer.Stop()

	select {
	case <-req.Context().Done():
		return fmt.Errorf("wait for MusicBrainz request interval: %w", req.Context().Err())
	case <-timer.C:
	}

	if err := req.Context().Err(); err != nil {
		return fmt.Errorf("start MusicBrainz request: %w", err)
	}

	t.next = time.Now().Add(t.interval)

	return nil
}
