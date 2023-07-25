package packager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
)

func checkDependency(dependency Dependency) error {
	// Check if the dependency exists in gopm_packages
	dependencyDir := filepath.Join("gopm_packages", dependency.Name)
	_, err := os.Stat(dependencyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("dependency %s not found in gopm_packages", dependency.Name)
		}
		return fmt.Errorf("failed to check dependency: %w", err)
	}

	// Check if the dependency version is suitable
	installedVersions, err := getInstalledVersions(dependencyDir)
	if err != nil {
		return fmt.Errorf("failed to get installed versions for dependency %s: %w", dependency.Name, err)
	}

	requiredVersion, err := semver.NewVersion(dependency.Version)
	if err != nil {
		return fmt.Errorf("invalid version for dependency %s: %w", dependency.Name, err)
	}

	for _, installedVersion := range installedVersions {
		// Use semver.Compare to compare the installed version with the required version based on the operator
		switch dependency.Operator {
		case "==", "=":
			if installedVersion.Equal(requiredVersion) {
				return nil
			}
		case ">=":
			if installedVersion.Compare(requiredVersion) >= 0 {
				return nil
			}
		case "<=":
			if installedVersion.Compare(requiredVersion) <= 0 {
				return nil
			}
		case ">":
			if installedVersion.Compare(requiredVersion) > 0 {
				return nil
			}
		case "<":
			if installedVersion.Compare(requiredVersion) < 0 {
				return nil
			}
		default:
			return fmt.Errorf("invalid operator for dependency %s: %s", dependency.Name, dependency.Operator)
		}
	}

	return fmt.Errorf("no suitable version found for dependency %s", dependency.Name)
}

func createDependenciesFile(dependencies []Dependency, dependenciesFile string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(dependenciesFile)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create dependencies directory: %w", err)
	}

	file, err := os.Create(dependenciesFile)
	if err != nil {
		return fmt.Errorf("failed to create dependencies file: %w", err)
	}
	defer file.Close()

	for _, dependency := range dependencies {
		line := fmt.Sprintf("%s %s\n", dependency.Name, dependency.Version)
		_, err := file.WriteString(line)
		if err != nil {
			return fmt.Errorf("failed to write to dependencies file: %w", err)
		}
	}

	return nil
}

func getInstalledVersions(dependencyDir string) ([]*semver.Version, error) {
	installedVersions := []*semver.Version{}
	files, err := os.ReadDir(dependencyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependency directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			version, err := semver.NewVersion(file.Name())
			if err == nil {
				installedVersions = append(installedVersions, version)
			}
		}
	}

	return installedVersions, nil
}
