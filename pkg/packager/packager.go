package packager

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreatePackage(packageFile string) (string, error) {
	// Read the package file
	mainPackage, err := readCreateFile(packageFile)
	if err != nil {
		return "", fmt.Errorf("failed to read package file: %v", err)
	}

	// Check if all dependencies exist and have suitable versions
	for _, dependency := range mainPackage.Dependencies {
		err := checkDependency(dependency)
		if err != nil {
			return "", fmt.Errorf("failed to check dependency: %v", err)
		}
	}

	// Create the package directory
	packageDir := filepath.Join("gopm_packages", mainPackage.Name, mainPackage.Version)
	err = os.MkdirAll(packageDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create package directory: %v", err)
	}

	// Copy targets to the package directory
	err = copyTargets(mainPackage.Targets, packageDir)
	if err != nil {
		_ = os.RemoveAll(packageDir)
		return "", fmt.Errorf("failed to copy targets: %v", err)
	}

	// Create dependencies.json file in the package directory
	dependenciesFile := filepath.Join(packageDir, "dependencies.json")
	err = createDependenciesFile(mainPackage.Dependencies, dependenciesFile)
	if err != nil {
		_ = os.RemoveAll(packageDir)
		return "", fmt.Errorf("failed to create dependencies file: %v", err)
	}

	return packageDir, nil
}
