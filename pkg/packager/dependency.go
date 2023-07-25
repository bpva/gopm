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
		versions, err := FindSuitableVersions(dependencyDir, update.Version, update.Operator, sshClient)
		if err != nil {
			session.Close()
			return fmt.Errorf("failed to find suitable versions: %w", err)
		}

		if len(versions) > 0 {
			// Use the greatest version suitable
			suitableVersion := versions[0]

			dependenciesFilePath := filepath.Join(dependencyDir, suitableVersion, "dependencies.json")

			command := fmt.Sprintf("cat %s", dependenciesFilePath)

			execSession, err := sshClient.NewSession()
			if err != nil {
				session.Close()
				return fmt.Errorf("failed to create SSH session for command execution: %w", err)
			}

			output, err := execSession.CombinedOutput(command)
			execSession.Close() // Close the execSession
			session.Close()

			if err != nil {
				return fmt.Errorf("failed to execute SSH command %s: %w. Please ensure that the package exists", command, err)
			}

			var dependencies []Dependency
			err = json.Unmarshal([]byte(strings.TrimSpace(string(output))), &dependencies)
			if err != nil {
				return fmt.Errorf("failed to parse dependencies JSON: %w", err)
			}

			err = addDependencies(updateConfig, dependencies, sshClient)
			if err != nil {
				return fmt.Errorf("failed to add dependencies: %w", err)
			}
		} else {
			session.Close()
		}
	}

	return nil
}

func addDependencies(updateConfig *UpdateConfig, dependencies []Dependency, sshClient *ssh.Client) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

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
			versions, err := FindSuitableVersions(dependencyDir, dependency.Version, dependency.Operator, sshClient)
			if err != nil {
				return fmt.Errorf("failed to find suitable versions: %w", err)
			}

			if len(versions) > 0 {
				// Use the greatest version
				suitableVersion := versions[0]

				dependenciesFilePath := filepath.Join(dependencyDir, suitableVersion, "dependencies.json")

				command := fmt.Sprintf("cat %s", dependenciesFilePath)
				execSession, err := sshClient.NewSession()
				if err != nil {
					session.Close()
					return fmt.Errorf("failed to create SSH session for command execution: %w", err)
				}
				output, err := execSession.CombinedOutput(command)
				execSession.Close()
				session.Close()
				if err != nil {
					return fmt.Errorf("failed to execute SSH command %s: %w. Please ensure that package exists", command, err)
				}

				var nestedDependencies []Dependency
				err = json.Unmarshal([]byte(strings.TrimSpace(string(output))), &nestedDependencies)
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

	return nil
}

func FindSuitableVersions(dir, targetVersion, operator string, sshClient *ssh.Client) ([]string, error) {
	versions := []string{}

	if sshClient != nil {
		session, err := sshClient.NewSession()
		if err != nil {
			return versions, fmt.Errorf("failed to create SSH session: %w", err)
		}
		defer session.Close()

		// Construct the remote command to list the directory contents
		command := fmt.Sprintf("ls -d %s/*/ | xargs -n 1 basename", dir)

		// Execute the remote command
		output, err := session.CombinedOutput(command)
		if err != nil {
			return versions, fmt.Errorf("failed to execute SSH command %s: %w", command, err)
		}

		// Split the output into individual lines
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")

		// Iterate over each line and check if it satisfies the version requirements
		for _, line := range lines {
			if satisfiesOperator(line, operator, targetVersion) {
				versions = append(versions, line)
			}
		}
	} else {
		// Read the directory contents locally
		files, err := os.ReadDir(dir)
		if err != nil {
			return versions, fmt.Errorf("failed to read directory: %w", err)
		}

		// Iterate over each file and check if it is a directory
		for _, file := range files {
			if file.IsDir() {
				version := file.Name()

				if satisfiesOperator(version, operator, targetVersion) {
					versions = append(versions, version)
				}
			}
		}
	}

	// Sort the versions in descending order
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
