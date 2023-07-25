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
		fmt.Println(".env file loaded successfully")
		return config, nil
	}

	// Try to locate the .env file in the current directory and one directory above
	envFileExists, err := findEnvFile()
	if err != nil {
		return config, fmt.Errorf("failed to find .env file: %w", err)
	}

	if envFileExists {
		fmt.Println(".env file found and loaded successfully")
		return config, nil
	}

	// Check if the required environment variables exist
	if hasRequiredEnvVars() {
		fmt.Println("Environment variables found")
		return config, nil
	}

	// If nothing found, print error and exit
	return config, fmt.Errorf("failed to configure SSH connection: no .env file found and required environment variables not set")
}

func loadEnvFile(envFilePath string) error {
	err := godotenv.Load(envFilePath)
	if err != nil {
		return err
	}
	return nil
}

func findEnvFile() (bool, error) {
	// Check in the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return false, err
	}
	envFileExists, err := fileExists(filepath.Join(currentDir, ".env"))
	if err != nil {
		return false, err
	}
	if envFileExists {
		return true, nil
	}

	// Check in the parent directory
	parentDir := filepath.Dir(currentDir)
	envFileExists, err = fileExists(filepath.Join(parentDir, ".env"))
	if err != nil {
		return false, err
	}
	if envFileExists {
		return true, nil
	}

	return false, nil
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
			return true
		}
	}
	return false
}
