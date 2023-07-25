package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  create  Create a package\n")
		fmt.Fprintf(os.Stderr, "  update  Update packages\n")
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
		create(flag.Arg(1))
	case "update":
		update()
	default:
		fmt.Fprintln(os.Stderr, "Unknown command. Available commands:")
		flag.Usage()
		os.Exit(1)
	}
}
