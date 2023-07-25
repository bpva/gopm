package connector

import (
	"fmt"

	"github.com/bpva/gopm/pkg/packager"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func DownloadUpdates(config packager.UpdateConfig, sshClient *ssh.Client) (arch []byte, err error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	return []byte{}, nil
}
