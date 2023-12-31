package connector

import (
	"fmt"
	"os"

	"github.com/bpva/gopm/pkg/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func CheckSSHConnection(config config.SSHConfig) error {
	// Create the SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            config.Login,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if config.Mode == "login+password" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(config.Password),
		}
	} else if config.Mode == "key" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		return fmt.Errorf("unknown SSH authentication mode: %s", config.Mode)
	}

	// Connect to the SSH server
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer client.Close()

	// Open an SFTP session
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to open SFTP session: %w", err)
	}
	defer sftpClient.Close()

	// Perform a simple operation
	// To fix
	homeDir, err := sftpClient.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	_, err = sftpClient.ReadDir(homeDir)
	if err != nil {
		return fmt.Errorf("failed to list directory on SSH server: %w", err)
	}

	return nil
}

func CreateSSHClient(sshConfig config.SSHConfig) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if sshConfig.Mode == "login+password" {
		authMethods = []ssh.AuthMethod{
			ssh.Password(sshConfig.Password),
		}
	} else if sshConfig.Mode == "key" {
		key, err := os.ReadFile(sshConfig.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		return nil, fmt.Errorf("unsupported SSH authentication mode: %s", sshConfig.Mode)
	}

	config := &ssh.ClientConfig{
		User:            sshConfig.Login,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%s", sshConfig.Host, sshConfig.Port)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	return client, nil
}
