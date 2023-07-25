package connector

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bpva/gopm/pkg/packager"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func DownloadUpdates(config packager.UpdateConfig, sshClient *ssh.Client) (arch []byte, varsions map[string]string, err error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Create a map to store the last suitable versions for each package
	lastSuitableVersions := map[string]string{}

	// Iterate over the updates in the config
	for _, update := range config.Updates {
		packageName := update.Name
		packageDir := fmt.Sprintf("gopm_packages/%s", packageName)

		versions, err := packager.FindSuitableVersions(packageDir, update.Version, update.Operator, sshClient)
		if err != nil {
			return []byte{}, nil, fmt.Errorf("failed to find suitable versions for package %s: %w", packageName, err)
		}

		if len(versions) > 0 {
			lastSuitableVersions[packageName] = versions[0]
		} else {
			return []byte{}, nil, fmt.Errorf("no suitable versions found for package %s", packageName)
		}
	}

	// Create a temporary directory on the remote server
	tempDir := fmt.Sprintf("/tmp/%d", time.Now().UnixNano())
	err = session.Run(fmt.Sprintf("mkdir %s", tempDir))
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create temporary directory on the remote server: %w", err)
	}

	// Reopen the session after creating the temporary directory
	session, err = sshClient.NewSession()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	session.Close()

	// Copy the needed packages to the temporary directory
	for packageName, version := range lastSuitableVersions {
		sourceDir := fmt.Sprintf("gopm_packages/%s/%s", packageName, version)
		destinationDir := fmt.Sprintf("%s/%s/%s", tempDir, packageName, version)
		session, err = sshClient.NewSession()
		if err != nil {
			return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
		}
		err = session.Run(fmt.Sprintf("mkdir -p %s", destinationDir))
		session.Close()
		if err != nil {
			return []byte{}, nil, fmt.Errorf("failed to create directory %s on the remote server: %w", destinationDir, err)
		}
		session, err = sshClient.NewSession()
		if err != nil {
			return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
		}
		err = session.Run(fmt.Sprintf("cp -r %s/. %s", sourceDir, destinationDir))
		session.Close()
		if err != nil {
			return []byte{}, nil, fmt.Errorf("failed to copy package %s/%s to the remote server: %w", packageName, version, err)
		}
	}

	// Archive the temporary directory
	tempFilePath := fmt.Sprintf("%s.tar.gz", tempDir)
	session, err = sshClient.NewSession()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	err = session.Run(fmt.Sprintf("tar -czvf %s -C %s .", tempFilePath, tempDir))
	session.Close()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create archive on the remote server: %w", err)
	}

	// Read the archive using SFTP
	remoteFile, err := sftpClient.Open(tempFilePath)
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to open the remote file: %w", err)
	}
	defer remoteFile.Close()

	// Read the file into a buffer
	arch, err = ioutil.ReadAll(remoteFile)
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to read the remote file: %w", err)
	}

	// Remove the temporary directory and archive file
	session, err = sshClient.NewSession()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	err = session.Run(fmt.Sprintf("rm -rf %s", tempDir))
	session.Close()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to remove the temporary directory: %w", err)
	}
	session, err = sshClient.NewSession()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	err = session.Run(fmt.Sprintf("rm %s", tempFilePath))
	session.Close()
	if err != nil {
		return []byte{}, nil, fmt.Errorf("failed to remove the archive file: %w", err)
	}

	return arch, lastSuitableVersions, nil
}
