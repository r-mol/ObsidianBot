package configs

import (
	"fmt"
	"os"

	"github.com/r-mol/ObsidianBot/pkg/tgbot"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server *ServerConfig `yaml:"server"`
	TgBot  *tgbot.Config `yaml:"tg_bot"`
}

func validateConfig(config *Config) error {
	switch {
	case config.Server == nil:
		return xerrors.New("\"server\" is required")
	case config.TgBot == nil:
		return xerrors.New("\"tg_bot\" is required")
	}

	if err := validateServerConfig(config.Server); err != nil {
		return fmt.Errorf("validate server config: %w", err)
	}

	if err := tgbot.ValidateConfig(config.TgBot); err != nil {
		return fmt.Errorf("validate telegram config: %w", err)
	}

	return nil
}

func ParseConfig(path string) (*Config, error) {
	config := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read file: %w", err)
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("umarshal config: %w", err)
	}

	if err = validateConfig(config); err != nil {
		return config, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}
