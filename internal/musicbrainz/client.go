// Package musicbrainz retrieves recording candidates from the MusicBrainz API.
package musicbrainz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/matching"
)

var (
	// ErrInvalidSearch identifies missing search names or an invalid candidate limit.
	ErrInvalidSearch = errors.New("artist, title, and a limit between 1 and 100 are required")
	errResponseLarge = errors.New("MusicBrainz response exceeds size limit")
)

// HTTPError preserves upstream status and retry guidance without response bodies.
type HTTPError struct {
	RetryAfter string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("MusicBrainz returned HTTP %d", e.StatusCode)
}

// Client implements matching.Catalog using one shared public-API request gate.
type Client struct {
	http   *http.Client
	config config.MusicBrainzConfig
}

// NewClientParams supplies upstream settings and HTTP connection management.
type NewClientParams struct {
	fx.In

	HTTPClient *http.Client
	Config     config.MusicBrainzConfig
}

// NewClient constructs a client using configuration validated by its caller.
func NewClient(params NewClientParams) *Client {
	const gateCapacity = 1

	base := params.HTTPClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	transport := &pacedTransport{
		base: base, interval: params.Config.RequestInterval, gate: make(chan struct{}, gateCapacity),
	}
	client := *params.HTTPClient
	client.Transport = transport
	// Redirects must not bypass the configured endpoint or request pacing.
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &Client{config: params.Config, http: &client}
}

// SearchRecordings retrieves candidates without interpreting upstream relevance as confidence.
func (c *Client) SearchRecordings(ctx context.Context, search matching.Search) ([]matching.Recording, error) {
	const (
		minLimit = 1
		maxLimit = 100
	)

	if strings.TrimSpace(search.Artist) == "" || strings.TrimSpace(search.Title) == "" ||
		search.Limit < minLimit || search.Limit > maxLimit {
		return nil, ErrInvalidSearch
	}

	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	endpoint, err := url.Parse(c.config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse MusicBrainz endpoint: %w", err)
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/ws/2/recording/"
	query := endpoint.Query()
	query.Set("query", recordingQuery(search))
	query.Set("fmt", "json")
	query.Set("limit", strconv.Itoa(search.Limit))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create MusicBrainz request: %w", err)
	}
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "application/json")

	return c.fetch(req)
}

func (c *Client) fetch(req *http.Request) ([]matching.Recording, error) {
	const (
		maxResponseBytes = 4 << 20
		overflowByte     = 1
	)

	response, err := c.http.Do(req)
	if err != nil {
		// Transport errors can include the request URL; never include configured credentials.
		var urlError *url.Error
		if errors.As(err, &urlError) {
			err = urlError.Err
		}

		return nil, fmt.Errorf("request MusicBrainz recordings: %w", err)
	}
	defer func() {
		_ = response.Body.Close() //nolint:errcheck // Read-only response close errors cannot change the search result.
	}()

	if response.StatusCode != http.StatusOK {
		return nil, &HTTPError{StatusCode: response.StatusCode, RetryAfter: response.Header.Get("Retry-After")}
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+overflowByte))
	if err != nil {
		return nil, fmt.Errorf("read MusicBrainz recordings: %w", err)
	}
	if len(body) > maxResponseBytes {
		return nil, errResponseLarge
	}

	var payload recordingResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode MusicBrainz recordings: %w", err)
	}
	if payload.Recordings == nil {
		return nil, errInvalidRecording
	}

	return payload.recordings()
}

func recordingQuery(search matching.Search) string {
	query := "artist:" + quote(search.Artist) + " AND recording:" + quote(search.Title)
	if search.Album != "" {
		query += " AND release:" + quote(search.Album)
	}

	return query
}

func quote(value string) string {
	escaped := `"`
	for _, char := range value {
		if strings.ContainsRune(`+-&|!(){}[]^"~*?:\/`, char) {
			escaped += `\`
		}
		escaped += string(char)
	}

	return escaped + `"`
}
