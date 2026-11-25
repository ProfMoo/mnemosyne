package frontendbuild_test

import (
	"testing"
	"testing/fstest"

	"github.com/profmoo/mnemosyne/internal/frontendbuild"
)

const configPath = "config/config.yaml"

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		configPath: {Data: []byte("outputDir: public\nresourceDir: internal/ui\napiPrefix: /custom/api/\n")},
	}
	cfg, err := frontendbuild.LoadConfig(files, configPath)
	if err != nil {
		t.Fatalf("load frontend config: %v", err)
	}
	if cfg.OutputDir != "public" || cfg.ResourceDir != "internal/ui" || cfg.APIPrefix != "/custom/api/" {
		t.Errorf("unexpected exporter config: %+v", cfg)
	}
}

func TestLoadConfigRejectsInvalidSettings(t *testing.T) {
	t.Parallel()

	for _, contents := range []string{
		"{}",
		"bindAddress: ':8080'\nfrontendDir: bin/web\napiPrefix: /api/",
		"outputDir: public\nresourceDir: internal/ui\napiPrefix: /api",
		"outputDir: public\nresourceDir: internal/ui\napiPrefix: /",
		"outputDir: public\nresourceDir: internal/ui\napiPrefix: /api//",
		"outputDir: 123\nresourceDir: internal/ui\napiPrefix: /api/",
		"outputDir: [",
	} {
		t.Run(contents, func(t *testing.T) {
			t.Parallel()

			files := fstest.MapFS{configPath: {Data: []byte(contents)}}
			if cfg, err := frontendbuild.LoadConfig(files, configPath); err == nil || cfg != nil {
				t.Fatalf("invalid config accepted: %v, %v", cfg, err)
			}
		})
	}
}
