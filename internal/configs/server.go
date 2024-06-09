package configs

import "golang.org/x/xerrors"

type ServerConfig struct {
	Host                 string `yaml:"host"`
	Port                 string `yaml:"port"`
	UserID               int64  `yaml:"user_id"`
	ObsidianAbsolutePath string `yaml:"obsidian_absolute_path"`
}

func validateServerConfig(config *ServerConfig) error {
	switch {
	case config.Port == "":
		return xerrors.New("\"port\" is required")
	case config.UserID == 0:
		return xerrors.New("\"user_id\" is required")
	case config.ObsidianAbsolutePath == "":
		return xerrors.New("\"obsidian_absolute_path\" is required")
	}

	return nil
}
