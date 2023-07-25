package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	arch, versions, err := connector.DownloadUpdates(updateConfig, sshClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to download updates: %s\n", err)
		os.Exit(1)
	}
	// delete versions to update
	fmt.Printf("Deleting local versions...\n")
	for packageName, version := range versions {
		packageDir := fmt.Sprintf("gopm_packages/%s/%s", packageName, version)
		err := os.RemoveAll(packageDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete package directory: %s\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Local versions deleted\n")

	// unpack arch
	fmt.Printf("Unpacking...\n")
	archReader := bytes.NewReader(arch)
	err = ExtractTarGz(archReader, "gopm_packages")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unpack archive: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Archive unpacked. Local versions updated\n")

}

func ExtractTarGz(gzipStream io.Reader, destination string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destination, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
		}
	}

	return nil
}
