package tgbot

import "golang.org/x/xerrors"

type Config struct {
	Verbose    bool   `yaml:"verbose"`
	Token      string `yaml:"token"`
	WebhookUrl string `yaml:"webhook_url"`
}

func ValidateConfig(config *Config) error {
	switch {
	case config.Token == "":
		return xerrors.New("\"token\" is required")
	case config.WebhookUrl == "":
		return xerrors.New("\"webhook_url\" is required")
	}

	return nil
}
