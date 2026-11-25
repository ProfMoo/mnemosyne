// Package frontendbuild exports go-app assets at build time, not server runtime.
package frontendbuild

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"github.com/profmoo/mnemosyne/internal/ui"
)

const (
	filePermissions = 0o644
	webDirectory    = "web"
	iconURL         = "/web/icon.svg"
)

// Export writes a standalone website alongside the previously compiled app.wasm.
func Export(directory, apiPrefix, resourceDirectory string) error {
	cfg := ui.PublicConfig{APIPrefix: apiPrefix}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate public config: %w", err)
	}

	publicConfig, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode public config: %w", err)
	}

	if copyErr := copyWebAssets(directory, resourceDirectory); copyErr != nil {
		return copyErr
	}
	version, err := assetVersion(directory, resourceDirectory, publicConfig)
	if err != nil {
		return err
	}

	handler := staticHandler(version)
	if err := app.GenerateStaticWebsite(directory, handler, "/"); err != nil {
		return fmt.Errorf("export static frontend: %w", err)
	}

	configPath := filepath.Join(directory, webDirectory, "config.json")
	if err := os.WriteFile(configPath, publicConfig, filePermissions); err != nil {
		return fmt.Errorf("write public config: %w", err)
	}

	return nil
}

func staticHandler(version string) *app.Handler {
	ui.RegisterRoutes()
	return &app.Handler{
		Name:        "Mnemosyne",
		Title:       "Mnemosyne — Listens",
		Description: "Your listening history.",
		Styles:      []string{"/web/styles.css"},
		Icon: app.Icon{
			Default: iconURL,
			Large:   iconURL,
			SVG:     iconURL,
		},
		Version: version,
	}
}

func copyWebAssets(directory, resources string) error {
	for _, name := range []string{"styles.css", "icon.svg"} {
		// #nosec G304 -- Resource directory is supplied by the local build tool.
		data, err := os.ReadFile(filepath.Join(resources, webDirectory, name))
		if err != nil {
			return fmt.Errorf("read frontend asset %q: %w", name, err)
		}
		// #nosec G703 -- Local build output directory and fixed asset filenames.
		if err := os.WriteFile(filepath.Join(directory, webDirectory, name), data, filePermissions); err != nil {
			return fmt.Errorf("write frontend asset %q: %w", name, err)
		}
	}
	return nil
}

func assetVersion(directory, resourceDirectory string, publicConfig []byte) (string, error) {
	// #nosec G304 -- Directory comes from local build config, not an HTTP request.
	contents, err := os.ReadFile(filepath.Join(directory, webDirectory, "app.wasm"))
	if err != nil {
		return "", fmt.Errorf("read WebAssembly: %w", err)
	}
	contents = append(contents, publicConfig...)
	for _, name := range []string{"styles.css", "icon.svg"} {
		// #nosec G304 -- Resource directory is supplied by the local build tool.
		data, err := os.ReadFile(filepath.Join(resourceDirectory, webDirectory, name))
		if err != nil {
			return "", fmt.Errorf("read asset for versioning: %w", err)
		}
		contents = append(contents, data...)
	}
	version := sha256.Sum256(contents)
	return fmt.Sprintf("%x", version), nil
}
