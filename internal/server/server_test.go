package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/server"
)

func TestServerServesFrontendAndAPI(t *testing.T) {
	t.Parallel()

	const filePermissions = 0o600
	const notFound = "404"

	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<h1>Mnemosyne</h1>"), filePermissions); err != nil {
		t.Fatal(err)
	}
	app := server.NewServer(server.NewServerParams{
		Config: &config.Config{
			BindAddress: ":8080",
			FrontendDir: dir,
			APIPrefix:   "/custom/api/",
		},
		Handler: server.NewHandler(server.NewHandlerParams{}),
	})

	for _, tc := range []struct {
		path   string
		body   string
		status int
	}{
		{path: "/", status: http.StatusOK, body: "Mnemosyne"},
		{path: "/listens", status: http.StatusInternalServerError, body: "NOT_IMPLEMENTED"},
		{path: "/custom/api/listens", status: http.StatusInternalServerError, body: "NOT_IMPLEMENTED"},
		{path: "/custom/api/missing", status: http.StatusNotFound, body: notFound},
		{path: "/api/listens", status: http.StatusNotFound, body: notFound},
		{path: "/missing.js", status: http.StatusNotFound, body: notFound},
	} {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			response := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, http.NoBody)
			app.Handler.ServeHTTP(response, request)

			if response.Code != tc.status {
				t.Errorf("status = %d, want %d", response.Code, tc.status)
			}
			if !strings.Contains(response.Body.String(), tc.body) {
				t.Errorf("body = %q, want substring %q", response.Body.String(), tc.body)
			}
		})
	}
}
