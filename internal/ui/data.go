// Package ui implements the browser UI and its API data loading.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/profmoo/mnemosyne/gen/client"
)

const requestTimeout = 15 * time.Second

const emptyListLength = 0

var errHTTPStatus = errors.New("request failed")

var errAPIPrefix = errors.New("API prefix must have leading and trailing slashes and nonempty path segments")

// PublicConfig contains only settings safe to publish to browsers.
type PublicConfig struct {
	APIPrefix string `json:"apiPrefix"`
}

// Validate rejects malformed prefixes rather than adding missing slashes.
func (cfg PublicConfig) Validate() error {
	if !regexp.MustCompile(`^(/[A-Za-z0-9_-]+)+/$`).MatchString(cfg.APIPrefix) {
		return errAPIPrefix
	}
	return nil
}

// ListenResult holds the formatted response and its empty-state indicator.
type ListenResult struct {
	JSON  string
	Empty bool
}

// LoadListens reads exported public configuration and calls the generated client.
func LoadListens(ctx context.Context, origin string) (ListenResult, error) {
	httpClient := &http.Client{Timeout: requestTimeout}
	cfg, err := loadConfig(ctx, httpClient, origin)
	if err != nil {
		return ListenResult{}, err
	}
	apiClient, err := client.NewClientWithResponses(origin+cfg.APIPrefix, client.WithHTTPClient(httpClient))
	if err != nil {
		return ListenResult{}, fmt.Errorf("create API client: %w", err)
	}
	response, err := apiClient.GetListensWithResponse(ctx)
	if err != nil {
		return ListenResult{}, fmt.Errorf("load listens: %w", err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, response.Body, "", "  "); err != nil {
		return ListenResult{}, fmt.Errorf("format listens response: %w", err)
	}
	result := ListenResult{JSON: pretty.String()}
	if response.JSON200 == nil {
		return result, fmt.Errorf("%w: HTTP %d", errHTTPStatus, response.StatusCode())
	}
	result.Empty = len(response.JSON200.Listens) == emptyListLength
	return result, nil
}

func loadConfig(ctx context.Context, httpClient *http.Client, origin string) (PublicConfig, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/web/config.json", http.NoBody)
	if err != nil {
		return PublicConfig{}, fmt.Errorf("create config request: %w", err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return PublicConfig{}, fmt.Errorf("load public config: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			slog.Warn("close public config response", "error", err)
		}
	}()
	if response.StatusCode != http.StatusOK {
		return PublicConfig{}, fmt.Errorf("load public config: %w: HTTP %d", errHTTPStatus, response.StatusCode)
	}
	var cfg PublicConfig
	if err := json.NewDecoder(response.Body).Decode(&cfg); err != nil {
		return PublicConfig{}, fmt.Errorf("decode public config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return PublicConfig{}, fmt.Errorf("validate public config: %w", err)
	}
	return cfg, nil
}
