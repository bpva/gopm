package packager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/crypto/ssh"
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
	dependenciesJSON, err := json.Marshal(dependencies)
	if err != nil {
		return fmt.Errorf("failed to marshal dependencies to JSON: %w", err)
	}

	err = os.WriteFile(dependenciesFile, dependenciesJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write dependencies file: %w", err)
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

func CollectDependencies(updateConfig *UpdateConfig, sshClient *ssh.Client) error {
	for _, update := range updateConfig.Updates {
		session, err := sshClient.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create SSH session: %w", err)
		}

		dependencyDir := filepath.Join("gopm_packages", update.Name)
		versions, err := findSuitableVersions(dependencyDir, update.Version, update.Operator)
		if err != nil {
			return fmt.Errorf("failed to find suitable versions: %w", err)
		}

		if len(versions) > 0 {
			// Use the first suitable version
			suitableVersion := versions[0]

			dependenciesFilePath := filepath.Join(dependencyDir, suitableVersion, "dependencies.json")

			command := fmt.Sprintf("cat %s", dependenciesFilePath)
			output, err := session.CombinedOutput(command)
			session.Close()
			if err != nil {
				return fmt.Errorf("failed to execute SSH command %s: %w. Please ensure that package exists", command, err)
			}

			var dependencies []Dependency
			err = json.Unmarshal([]byte(strings.TrimSpace(string(output))), &dependencies)
			if err != nil {
				return fmt.Errorf("failed to parse dependencies JSON: %w", err)
			}

			err = addDependencies(updateConfig, dependencies)
			if err != nil {
				return fmt.Errorf("failed to add dependencies: %w", err)
			}
		}
	}

	return nil
}

func addDependencies(updateConfig *UpdateConfig, dependencies []Dependency) error {
	for len(dependencies) > 0 {
		dependency := dependencies[0]
		dependencies = dependencies[1:]

		found := false

		for _, update := range updateConfig.Updates {
			if update.Name == dependency.Name {
				if satisfiesOperator(update.Version, update.Operator, dependency.Version) {
					found = true
					break
				}
			}
		}

		if !found {
			newDependency := Dependency{
				Name:     dependency.Name,
				Version:  dependency.Version,
				Operator: dependency.Operator,
			}
			updateConfig.Updates = append(updateConfig.Updates, newDependency)

			dependencyDir := filepath.Join("gopm_packages", dependency.Name)
			versions, err := findSuitableVersions(dependencyDir, dependency.Version, dependency.Operator)
			if err != nil {
				return fmt.Errorf("failed to find suitable versions: %w", err)
			}

			if len(versions) > 0 {
				// Use the greatest version
				suitableVersion := versions[0]

				dependenciesFilePath := filepath.Join(dependencyDir, suitableVersion, "dependencies.json")

				fileContent, err := os.ReadFile(dependenciesFilePath)
				if err != nil {
					return fmt.Errorf("failed to read dependencies file: %w", err)
				}

				if len(fileContent) > 0 && string(fileContent) != "[]" && strings.TrimSpace(string(fileContent)) != "null" {
					var nestedDependencies []Dependency
					fmt.Printf("File content: %s\n", string(fileContent))
					err = json.Unmarshal(fileContent, &nestedDependencies)
					if err != nil {
						return fmt.Errorf("failed to parse dependencies JSON: %w", err)
					}

					// Check if all nested dependencies are already found
					allFound := true
					for _, nestedDependency := range nestedDependencies {
						nestedFound := false
						for _, update := range updateConfig.Updates {
							if update.Name == nestedDependency.Name && satisfiesOperator(update.Version, update.Operator, nestedDependency.Version) {
								nestedFound = true
								break
							}
						}
						if !nestedFound {
							allFound = false
							dependencies = append(dependencies, nestedDependency)
						}
					}

					if !allFound {
						continue
					}
				}
			}
		}
	}

	return nil
}

func findSuitableVersions(dir, targetVersion, operator string) ([]string, error) {
	versions := []string{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return versions, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			version := file.Name()

			if satisfiesOperator(version, operator, targetVersion) {
				versions = append(versions, version)
			}
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		v1, _ := semver.NewVersion(versions[i])
		v2, _ := semver.NewVersion(versions[j])
		return v1.GreaterThan(v2)
	})

	return versions, nil
}

func satisfiesOperator(version, operator, targetVersion string) bool {
	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	t, err := semver.NewVersion(targetVersion)
	if err != nil {
		return false
	}

	switch operator {
	case "==":
		return v.Equal(t)
	case "<":
		return v.LessThan(t)
	case "<=":
		return v.LessThan(t) || v.Equal(t)
	case ">":
		return v.GreaterThan(t)
	case ">=":
		return v.GreaterThan(t) || v.Equal(t)
	default:
		return false
	}
}
