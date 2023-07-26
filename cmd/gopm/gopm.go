package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bpva/gopm/pkg/archiver"
	"github.com/bpva/gopm/pkg/config"
	"github.com/bpva/gopm/pkg/connector"
	"github.com/bpva/gopm/pkg/packager"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  create  Create a package\n")
		fmt.Fprintf(os.Stderr, "  update  Update packages\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -env  Path to the .env file\n")
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	envFilePath := flag.String("env", "", "Path to the .env file")
	sshConfig, err := config.Configure(*envFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure SSH connection: %v\n", err)
		os.Exit(1)
	}

	command := flag.Arg(0)
	switch command {
	case "create":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: %s create <package.json>\n", os.Args[0])
			os.Exit(1)
		}
		create(flag.Arg(1), sshConfig)
	case "update":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: %s create <package.json>\n", os.Args[0])
			os.Exit(1)
		}
		update(flag.Arg(1), sshConfig)
	default:
		fmt.Fprintln(os.Stderr, "Unknown command. Available commands:")
		flag.Usage()
		os.Exit(1)
	}
}

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
