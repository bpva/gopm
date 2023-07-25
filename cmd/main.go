package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bpva/gopm/pkg/config"
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

	envFilePath := flag.String("env", "", "Path to the .env file")
	sshConfig, err := config.Configure(*envFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure SSH connection: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
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
