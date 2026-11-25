// Package config loads and validates application configuration.
package config

import (
	"fmt"
	"io/fs"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// Config contains the application settings.
type Config struct {
	BindAddress string `koanf:"bindAddress" validate:"required,hostname_port|tcp6_addr"`
	FrontendDir string `koanf:"frontendDir" validate:"required"`
	APIPrefix   string `koanf:"apiPrefix"   validate:"required"`
}

// LoadConfig reads a YAML file from configFiles and returns validated settings.
func LoadConfig(configFiles fs.FS, configLocation string) (*Config, error) {
	data, err := fs.ReadFile(configFiles, configLocation)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", configLocation, err)
	}

	settings := koanf.New(".")
	if err := settings.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load config %q: %w", configLocation, err)
	}

	var cfg Config
	unmarshalConf := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           &cfg,
			TagName:          "koanf",
			ErrorUnused:      true,
			WeaklyTypedInput: false,
		},
	}
	if err := settings.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", configLocation, err)
	}

	validation := validator.New(validator.WithRequiredStructEnabled())
	validation.RegisterStructValidation(validateAPIPrefix, Config{})
	if err := validation.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", configLocation, err)
	}

	return &cfg, nil
}

func validateAPIPrefix(validation validator.StructLevel) {
	prefix := validation.Current().FieldByName("APIPrefix").String()
	if !regexp.MustCompile(`^(/[A-Za-z0-9_-]+)+/$`).MatchString(prefix) {
		validation.ReportError(prefix, "APIPrefix", "apiPrefix", "api_prefix", "")
	}
}
