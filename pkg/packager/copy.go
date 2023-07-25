package packager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	Name         string       `json:"name" yaml:"name"`
	Version      string       `json:"ver" yaml:"ver"`
	Targets      []Target     `json:"targets,omitempty" yaml:"targets,omitempty"`
	Dependencies []Dependency `json:"packets,omitempty" yaml:"packets,omitempty"`
}

type Target struct {
	Path    string `json:"path" yaml:"path"`
	Exclude string `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}

type Dependency struct {
	Name     string `json:"name" yaml:"name"`
	Version  string `json:"ver" yaml:"ver"`
	Operator string `json:"operator" yaml:"operator"`
}

func copyTargets(targets []Target, packageDir string) error {
	for _, target := range targets {
		matches, err := filepath.Glob(target.Path)
		if err != nil {
			return fmt.Errorf("failed to match pattern '%s': %w", target.Path, err)
		}
		excludes := strings.Split(target.Exclude, ",")
		for _, match := range matches {
			fileInfo, err := os.Stat(match)
			if err != nil {
				return fmt.Errorf("failed to access path '%s': %w", match, err)
			}

			if shouldExclude(fileInfo.Name(), excludes) {
				continue
			}
			// pakageDir: /gopm_packages/<package-name>/<package-version>/
			destPath := getDestinationPath(packageDir, match)
			if fileInfo.IsDir() {
				err = copyDir(match, destPath, excludes)
				if err != nil {
					return fmt.Errorf("failed to copy directory '%s' to '%s': %w", match, destPath, err)
				}
			} else {
				err = copyFile(match, destPath, excludes)
				if err != nil {
					return fmt.Errorf("failed to copy file '%s' to '%s': %w", match, destPath, err)
				}
			}
		}
	}

	return nil
}
func getDestinationPath(packageDir, filePath string) string {
	// to fix
	baseDir := filepath.Base(packageDir)
	fileName := filepath.Base(filePath)

	destinationPath := filepath.Join(filepath.Dir(packageDir), baseDir, fileName)

	return destinationPath
}
func shouldExclude(fileName string, exclude []string) bool {
	for _, excludedPattern := range exclude {
		matched, err := filepath.Match(excludedPattern, fileName)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func copyFile(srcPath, destPath string, excludes []string) error {
	srcFileInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if srcFileInfo.IsDir() {
		return copyDir(srcPath, destPath, excludes)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destDir := filepath.Dir(destPath)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func copyDir(srcDir, destDir string, excludes []string) error {
	// Create the destination directory
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", destDir, err)
		}
	}

	// Copy the contents of the source directory to the package
	fileInfos, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s': %w", srcDir, err)
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(srcDir, fileInfo.Name())
		destPath := filepath.Join(destDir, fileInfo.Name())

		if fileInfo.IsDir() {
			err = copyDir(srcPath, destPath, excludes)
			if err != nil {
				return err
			}
		} else {
			if shouldExclude(fileInfo.Name(), excludes) {
				continue
			}
			err = copyFile(srcPath, destPath, excludes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
