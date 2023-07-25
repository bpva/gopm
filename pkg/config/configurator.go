package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type SSHConfig struct {
	Mode     string
	Login    string
	KeyPath  string
	Password string
	Host     string
	Port     string
}

func Configure(envFilePath string) (SSHConfig, error) {
	config := SSHConfig{}

	// Check if the .env file is specified
	if envFilePath != "" {
		err := loadEnvFile(envFilePath)
		if err != nil {
			return config, fmt.Errorf("failed to load .env file: %w", err)
		}
		err = fillSSHConfigFields(&config)
		if err != nil {
			return config, fmt.Errorf("failed to configure SSH connection: %w", err)
		}
		return config, nil
	}

	// Try to locate the .env file in the current directory and one directory above
	envFilePath, envFileExists, err := findEnvFile()
	if err != nil {
		return config, fmt.Errorf("failed to find .env file: %w", err)
	}

	if envFileExists {
		err := loadEnvFile(envFilePath)
		if err != nil {
			return config, fmt.Errorf("failed to load .env file: %w", err)
		}
		if hasRequiredEnvVars() {
			err := fillSSHConfigFields(&config)
			if err != nil {
				return config, fmt.Errorf("failed to configure SSH connection: %w", err)
			}
			return config, nil
		} else {
			return config, fmt.Errorf("failed to configure SSH connection: no .env file found and required environment variables not set")
		}
	}

	// If nothing found, print error and exit
	return config, fmt.Errorf("failed to configure SSH connection: no .env file found and required environment variables not set")
}
func fillSSHConfigFields(config *SSHConfig) error {
	config.Mode = os.Getenv("GOPM_SSH_MODE")
	config.Login = os.Getenv("GOPM_SSH_LOGIN")
	config.KeyPath = os.Getenv("GOPM_SSH_KEY_PATH")
	config.Password = os.Getenv("GOPM_SSH_PASSWORD")
	config.Host = os.Getenv("GOPM_SSH_HOST")
	config.Port = os.Getenv("GOPM_SSH_PORT")

	if err := validateSSHConfig(*config); err != nil {
		return err
	}

	return nil
}
func loadEnvFile(envFilePath string) error {
	err := godotenv.Load(envFilePath)
	if err != nil {
		return err
	}
	if !hasRequiredEnvVars() {
		return fmt.Errorf("required environment variables missing in .env file or key file not exists")
	}
	return nil
}

func findEnvFile() (string, bool, error) {
	// Check in the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", false, err
	}
	envFilePath := filepath.Join(currentDir, ".env")
	envFileExists, err := fileExists(envFilePath)
	if err != nil {
		return "", false, err
	}
	if envFileExists {
		return envFilePath, true, nil
	}

	// Check in the parent directory
	parentDir := filepath.Dir(currentDir)
	envFilePath = filepath.Join(parentDir, ".env")
	envFileExists, err = fileExists(envFilePath)
	if err != nil {
		return "", false, err
	}
	if envFileExists {
		return envFilePath, true, nil
	}

	return "", false, nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func hasRequiredEnvVars() bool {
	mode := os.Getenv("GOPM_SSH_MODE")
	if mode == "login+password" {
		login := os.Getenv("GOPM_SSH_LOGIN")
		password := os.Getenv("GOPM_SSH_PASSWORD")
		host := os.Getenv("GOPM_SSH_HOST")
		port := os.Getenv("GOPM_SSH_PORT")
		if login != "" && password != "" && host != "" && port != "" {
			return true
		}
	} else if mode == "key" {
		login := os.Getenv("GOPM_SSH_LOGIN")
		keyPath := os.Getenv("SSH_KEY_PATH")
		host := os.Getenv("GOPM_SSH_HOST")
		port := os.Getenv("GOPM_SSH_PORT")
		if login != "" && keyPath != "" && host != "" && port != "" {
			keyFileExists, err := fileExists(keyPath)
			if err != nil {
				fmt.Println("failed to check if key file exists: ", err)
				return false
			}
			if keyFileExists {
				return true
			} else {
				fmt.Println("key file does not exist")
				return false
			}
		}
	}
	return false
}
