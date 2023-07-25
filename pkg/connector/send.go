package connector

import (
	"fmt"
	"path/filepath"

	"github.com/bpva/gopm/pkg/config"
)

func SendAndUnpackArchive(arch []byte, sshConfig config.SSHConfig) error {
	client, err := ssh.Connect(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	remoteDir := "gopm_packages"
	err = session.MkdirAll(remoteDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	remoteFile := filepath.Join(remoteDir, "package.tar.gz")
	err = session.WriteFile(remoteFile, arch)
	if err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	err = session.Run(fmt.Sprintf("tar -xf %s -C %s", remoteFile, remoteDir))
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}
