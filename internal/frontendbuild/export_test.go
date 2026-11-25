package frontendbuild_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/profmoo/mnemosyne/internal/frontendbuild"
	"github.com/profmoo/mnemosyne/internal/ui"
)

const (
	directoryPermissions = 0o755
	filePermissions      = 0o644
)

func TestExportServesStandaloneStaticFiles(t *testing.T) {
	t.Parallel()
	output := t.TempDir()
	resources := t.TempDir()
	writeAsset(t, output, "app.wasm", []byte("\x00asm"))
	writeAsset(t, resources, "styles.css", []byte("body { color: black; }"))
	writeAsset(t, resources, "icon.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`))
	if err := frontendbuild.Export(output, "/custom/api/", resources); err != nil {
		t.Fatalf("export website: %v", err)
	}

	static := http.FileServer(http.FS(os.DirFS(output)))
	for _, path := range []string{
		"/", "/app.js", "/app.css", "/wasm_exec.js", "/app-worker.js", "/manifest.webmanifest",
		"/web/app.wasm", "/web/styles.css", "/web/icon.svg", "/web/config.json",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			response := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody)
			static.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Errorf("static asset status = %d, body = %q", response.Code, response.Body.String())
			}
		})
	}
	configBytes, err := fs.ReadFile(os.DirFS(output), "web/config.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg ui.PublicConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.APIPrefix != "/custom/api/" || strings.Contains(string(configBytes), "outputDir") {
		t.Errorf("unexpected public config: %s", configBytes)
	}
}

func writeAsset(t *testing.T, directory, name string, contents []byte) {
	t.Helper()

	web := filepath.Join(directory, "web")
	if err := os.MkdirAll(web, directoryPermissions); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(web, name), contents, filePermissions); err != nil {
		t.Fatal(err)
	}
}
