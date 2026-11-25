package config_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-playground/validator/v10"

	"github.com/profmoo/mnemosyne/internal/config"
)

const configPath = "config/config.yaml"

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		address string
	}{
		{name: "all interfaces", address: ":8080"},
		{name: "IPv4", address: "127.0.0.1:8080"},
		{name: "IPv6", address: "[::1]:8080"},
		{name: "hostname", address: "localhost:8080"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			files := configFS("bindAddress: \"" + test.address + "\"\nfrontendDir: bin/web\napiPrefix: /api/\n")
			cfg, err := config.LoadConfig(files, configPath)
			if err != nil {
				t.Fatalf("load valid config: %v", err)
			}
			if cfg.BindAddress != test.address {
				t.Errorf("bind address = %q, want %q", cfg.BindAddress, test.address)
			}
			if cfg.FrontendDir != "bin/web" || cfg.APIPrefix != "/api/" {
				t.Errorf("unexpected frontend configuration: %+v", cfg)
			}
		})
	}
}

func TestLoadConfigCustomFrontendSettings(t *testing.T) {
	t.Parallel()

	files := configFS("bindAddress: ':8080'\nfrontendDir: /app/frontend\napiPrefix: /music/v1/\n")
	cfg, err := config.LoadConfig(files, configPath)
	if err != nil {
		t.Fatalf("load valid frontend config: %v", err)
	}
	if cfg.APIPrefix != "/music/v1/" || cfg.FrontendDir != "/app/frontend" {
		t.Errorf("unexpected frontend config: %+v", cfg)
	}
}

func TestLoadConfigInvalidFrontendSettings(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		yaml string
	}{
		{name: "missing directory", yaml: "apiPrefix: /api/"},
		{name: "missing prefix", yaml: "frontendDir: bin/web"},
		{name: "root prefix", yaml: "frontendDir: bin/web\napiPrefix: /"},
		{name: "relative prefix", yaml: "frontendDir: bin/web\napiPrefix: api/"},
		{name: "missing trailing slash", yaml: "frontendDir: bin/web\napiPrefix: /api"},
		{name: "double trailing slash", yaml: "frontendDir: bin/web\napiPrefix: /api//"},
		{name: "empty segment", yaml: "frontendDir: bin/web\napiPrefix: /api//v1/"},
		{name: "route wildcard", yaml: "frontendDir: bin/web\napiPrefix: /{path}/"},
		{name: "query string", yaml: "frontendDir: bin/web\napiPrefix: /api?query/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := config.LoadConfig(configFS("bindAddress: ':8080'\n"+tc.yaml+"\n"), configPath)
			var validationErrors validator.ValidationErrors
			if cfg != nil || !errors.As(err, &validationErrors) {
				t.Fatalf("expected validation error and nil config, got %v, %v", cfg, err)
			}
		})
	}
}

func TestLoadConfigInvalidAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yaml string
	}{
		{name: "missing", yaml: "{}\n"},
		{name: "empty file", yaml: ""},
		{name: "empty address", yaml: "bindAddress: \"\"\n"},
		{name: "no port", yaml: "bindAddress: localhost\n"},
		{name: "invalid port", yaml: "bindAddress: localhost:abc\n"},
		{name: "port out of range", yaml: "bindAddress: localhost:65536\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := config.LoadConfig(configFS(test.yaml), configPath)
			if cfg != nil {
				t.Error("invalid configuration returned a config")
			}
			var validationErrors validator.ValidationErrors
			if !errors.As(err, &validationErrors) {
				t.Fatalf("expected validator error, got %v", err)
			}
		})
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yaml string
	}{
		{name: "malformed", yaml: "bindAddress: [\n"},
		{name: "unknown field", yaml: "bindAdress: \":8080\"\n"},
		{name: "wrong type", yaml: "bindAddress: 8080\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := config.LoadConfig(configFS(test.yaml), configPath)
			if err == nil || cfg != nil {
				t.Fatalf("expected nil config and error, got %v, %v", cfg, err)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadConfig(fstest.MapFS{}, "config/missing.yaml")
	if !errors.Is(err, fs.ErrNotExist) || cfg != nil {
		t.Fatalf("expected nil config and missing-file error, got %v, %v", cfg, err)
	}
}

func configFS(contents string) fs.FS {
	return fstest.MapFS{
		configPath: {Data: []byte(contents)},
	}
}
