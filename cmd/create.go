package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bpva/gopm/pkg/archiver"
	"github.com/bpva/gopm/pkg/config"
	"github.com/bpva/gopm/pkg/connector"
	"github.com/bpva/gopm/pkg/packager"
)

func create(packageFile string, sshConfig config.SSHConfig) {
	name, version, err := packager.GetNameAndVersionFromConfigFile(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get name and version from config file: %s\n", err)
		os.Exit(1)
	}
	supposedPath := fmt.Sprintf("gopm_packages/%s/%s", name, version)
	// Check if the package directory already exists
	if _, err := os.Stat(supposedPath); err == nil {
		fmt.Printf("Directory '%s' already exists. Do you want to force rewrite it? (yes/no): ", supposedPath)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "yes" && answer != "y" {
			fmt.Println("Skipping package upload and unpack...")
			return
		}
	}

	if _, err := os.Stat(supposedPath); err == nil {
		// Delete the package directory
		err = os.RemoveAll(supposedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete package directory: %s\n", err)
			os.Exit(1)
		}
	}

	packageDir, err := packager.CreatePackage(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create package: %s\n", err)
		os.Exit(1)
	}

	arch, err := archiver.Archive(packageDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create archive: %s\n", err)
		os.Exit(1)
	}
	err = connector.CheckSSHConnection(sshConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to SSH server: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("SSH connection successful")
	}
	fmt.Printf("Package %s v%s created localy\n", name, version)
	sshClient, err := connector.CreateSSHClient(sshConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create SSH client: %s\n", err)
		os.Exit(1)
	}

	err = connector.UploadAndUnpackArchive(arch, sshClient, name, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to upload and unpack archive: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Package %s v%s uploaded and unpacked on remote server %s@%s\n", name, version, sshConfig.Login, sshConfig.Host)
	}

}
