package config

import "fmt"

func validateSSHConfig(config SSHConfig) error {
	if config.Mode == "" {
		return fmt.Errorf("SSH_MODE is not set")
	}

	if config.Mode != "login+password" && config.Mode != "key" {
		return fmt.Errorf("invalid SSH_MODE value")
	}

	if config.Mode == "login+password" && config.Password == "" {
		return fmt.Errorf("SSH_PASSWORD is not set")
	}

	if config.Mode == "key" && config.KeyPath == "" {
		return fmt.Errorf("SSH_KEY_PATH is not set")
	}

	if config.Login == "" {
		return fmt.Errorf("SSH_LOGIN is not set")
	}

	if config.Host == "" {
		return fmt.Errorf("SSH_HOST is not set")
	}

	if config.Port == "" {
		return fmt.Errorf("SSH_PORT is not set")
	}

	return nil
}
