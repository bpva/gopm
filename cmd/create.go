package main

import (
	"fmt"
	"os"

	"github.com/bpva/gopm/pkg/archiver"
	"github.com/bpva/gopm/pkg/config"
	"github.com/bpva/gopm/pkg/connector"
	"github.com/bpva/gopm/pkg/packager"
)

func create(packageFile string, sshConfig config.SSHConfig) {
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
	fmt.Println("size of arch is ", len(arch))
	fmt.Println("package created")
	err = connector.SendAndUnpackArchive(arch, sshConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send and unpack archive: %s\n", err)
		os.Exit(1)
	}
}
