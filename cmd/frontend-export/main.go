// Command frontend-export builds WASM and exports a standalone static website.
package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/profmoo/mnemosyne/internal/frontendbuild"
)

const directoryPermissions = 0o755

const failureExitCode = 1

//go:embed config/*.yaml
var configFiles embed.FS

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("build frontend", "error", err)
		os.Exit(failureExitCode)
	}
}

func run(ctx context.Context) error {
	cfg, err := frontendbuild.LoadConfig(configFiles, "config/config.yaml")
	if err != nil {
		return fmt.Errorf("load build configuration: %w", err)
	}
	output := filepath.Join(cfg.OutputDir, "web", "app.wasm")
	if err := os.MkdirAll(filepath.Dir(output), directoryPermissions); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	// #nosec G204 -- Output comes from local build config; no shell is invoked.
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", output, "./cmd/frontend")
	// These are compiler target settings, not application configuration.
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("compile WebAssembly: %w", err)
	}
	if err := frontendbuild.Export(cfg.OutputDir, cfg.APIPrefix, cfg.ResourceDir); err != nil {
		return fmt.Errorf("export frontend: %w", err)
	}
	return nil
}
