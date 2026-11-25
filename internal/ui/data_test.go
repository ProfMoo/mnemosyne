package ui_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/profmoo/mnemosyne/internal/ui"
)

const (
	publicConfig = `{"apiPrefix":"/custom/api/"}`
	emptyListens = `{"listens":[]}`
)

func TestLoadListens(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		body   string
		status int
		empty  bool
		failed bool
	}{
		{name: "empty", body: emptyListens, status: http.StatusOK, empty: true},
		{
			name:   "response",
			body:   `{"listens":[{"id":"00000000-0000-0000-0000-000000000001"}]}`,
			status: http.StatusOK,
		},
		{
			name:   "HTTP failure",
			body:   `{"code":"NOT_IMPLEMENTED","message":"Not implemented"}`,
			status: http.StatusInternalServerError,
			failed: true,
		},
		{name: "invalid JSON", body: "{", status: http.StatusOK, failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := responseServer(t, publicConfig, tc.body, tc.status)
			result, err := ui.LoadListens(t.Context(), server.URL)
			if (err != nil) != tc.failed {
				t.Fatalf("error = %v, want failure %t", err, tc.failed)
			}
			if result.Empty != tc.empty {
				t.Errorf("empty = %t, want %t", result.Empty, tc.empty)
			}
			if tc.body != "{" && !strings.Contains(result.JSON, "\n") {
				t.Errorf("response was not formatted: %q", result.JSON)
			}
		})
	}
}

func TestLoadListensInvalidConfig(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`{}`,
		`{"apiPrefix":"/api"}`,
		`{"apiPrefix":"/api//"}`,
		`{"apiPrefix":"https://untrusted.example/"}`,
		`{`,
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()

			server := responseServer(t, body, emptyListens, http.StatusOK)
			if _, err := ui.LoadListens(t.Context(), server.URL); err == nil {
				t.Fatal("invalid public config was accepted")
			}
		})
	}
}

func TestLoadListensCancellation(t *testing.T) {
	t.Parallel()

	server := responseServer(t, publicConfig, emptyListens, http.StatusOK)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ui.LoadListens(ctx, server.URL); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func responseServer(t *testing.T, configBody, apiBody string, apiStatus int) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/web/config.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, configBody); err != nil {
			t.Errorf("write config response: %v", err)
		}
	})
	mux.HandleFunc("/custom/api/listens", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiStatus)
		if _, err := io.WriteString(w, apiBody); err != nil {
			t.Errorf("write API response: %v", err)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}
