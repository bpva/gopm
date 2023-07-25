package connector

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func UploadAndUnpackArchive(arch []byte, sshClient *ssh.Client, packageName, packageVersion string) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Generate a random archive name
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	archiveName := "archive_" + strconv.Itoa(r.Intn(10000)) + ".zip"

	// Check if lock file exists
	lockFileName := archiveName + ".lock"
	for {
		_, err := sftpClient.Stat(lockFileName)
		if err != nil {
			if os.IsNotExist(err) {
				// Create the lock file
				lockFile, err := sftpClient.Create(lockFileName)
				if err != nil {
					return fmt.Errorf("failed to create lock file: %w", err)
				}
				lockFile.Close()
				break
			}
			return fmt.Errorf("failed to check lock file: %w", err)
		}
		time.Sleep(time.Second * 5)
	}

	remoteFile, err := sftpClient.Create(archiveName)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	_, err = remoteFile.Write(arch)
	if err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	targetDir := fmt.Sprintf("gopm_packages/%s/%s", packageName, packageVersion)
	createCmd := fmt.Sprintf("mkdir -p %s && unzip -o %s -d %s", targetDir, archiveName, targetDir)
	err = session.Run(createCmd)
	if err != nil {
		return fmt.Errorf("failed to unpack archive on remote server: %w", err)
	}

	err = sftpClient.Remove(archiveName)
	if err != nil {
		return fmt.Errorf("failed to delete archive file: %w", err)
	}

	// Delete the lock file
	err = sftpClient.Remove(lockFileName)
	if err != nil {
		return fmt.Errorf("failed to delete lock file: %w", err)
	}

	return nil
}
