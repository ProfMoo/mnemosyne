package frontendbuild

import (
	"fmt"
	"io/fs"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"

	"github.com/profmoo/mnemosyne/internal/ui"
)

// Config contains settings owned by the frontend export command.
type Config struct {
	OutputDir   string `koanf:"outputDir"   validate:"required"`
	ResourceDir string `koanf:"resourceDir" validate:"required"`
	APIPrefix   string `koanf:"apiPrefix"   validate:"required"`
}

// LoadConfig reads exporter configuration without loading backend settings.
func LoadConfig(files fs.FS, path string) (*Config, error) {
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return nil, fmt.Errorf("read frontend build config: %w", err)
	}
	settings := koanf.New(".")
	if err := settings.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load frontend build config: %w", err)
	}
	var cfg Config
	options := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           &cfg,
			TagName:          "koanf",
			ErrorUnused:      true,
			WeaklyTypedInput: false,
		},
	}
	if err := settings.UnmarshalWithConf("", &cfg, options); err != nil {
		return nil, fmt.Errorf("decode frontend build config: %w", err)
	}
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate frontend build config: %w", err)
	}
	if err := (ui.PublicConfig{APIPrefix: cfg.APIPrefix}).Validate(); err != nil {
		return nil, fmt.Errorf("validate frontend API prefix: %w", err)
	}
	return &cfg, nil
}
