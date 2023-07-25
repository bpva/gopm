package main

import (
	"fmt"
	"os"

	"github.com/bpva/gopm/pkg/config"
	"github.com/bpva/gopm/pkg/connector"
	"github.com/bpva/gopm/pkg/packager"
)

func update(packageFile string, sshConfig config.SSHConfig) {
	updateConfig, err := packager.ReadUpdateFile(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read update file: %v\n", err)
		os.Exit(1)
	}

	sshClient, err := connector.CreateSSHClient(sshConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create SSH client: %s\n", err)
		os.Exit(1)
	}
	err = packager.CollectDependencies(&updateConfig, sshClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to collect dependencies: %s\n", err)
		os.Exit(1)
	}
	for _, update := range updateConfig.Updates {
		fmt.Println(update)
	}

	arch, err := connector.DownloadUpdates(updateConfig, sshClient)
	_ = arch
}
