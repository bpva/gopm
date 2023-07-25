package main

import (
	"fmt"
	"os"

	"github.com/bpva/gopm/pkg/packager"
)

func create(packageFile string) {
	err := packager.CreatePackage(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create package: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("package created")
}
