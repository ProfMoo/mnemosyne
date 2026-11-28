package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/internal/config"
	"github.com/profmoo/mnemosyne/internal/repository"
	"github.com/profmoo/mnemosyne/internal/server"
)

func TestServerServesFrontendAndAPI(t *testing.T) {
	t.Parallel()

	const filePermissions = 0o600
	const notFound = "404"

	dir := t.TempDir()
	database, err := db.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := database.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})
	listenRepository := repository.NewSqliteListenRepository(repository.NewSqliteListenRepositoryParams{DB: database})
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
		Handler: server.NewHandler(server.NewHandlerParams{
			GetListensHandler: server.NewGetListensHandler(
				server.GetListensHandlerParams{Repository: listenRepository},
			),
		}),
	})

	for _, tc := range []struct {
		path   string
		body   string
		status int
	}{
		{path: "/", status: http.StatusOK, body: "Mnemosyne"},
		{path: "/listens", status: http.StatusOK, body: `"listens":[]`},
		{path: "/custom/api/listens", status: http.StatusOK, body: `"listens":[]`},
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
