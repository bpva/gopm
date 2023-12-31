package packager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Update struct {
	Name     string `json:"name" yaml:"name"`
	Version  string `json:"ver" yaml:"ver"`
	Operator string `json:"operator" yaml:"operator"`
}

type UpdateConfig struct {
	Updates []Dependency `json:"packages" yaml:"packages"`
}

// Custom unmarshaler for the Dependency struct
func (d *Dependency) UnmarshalJSON(data []byte) error {
	var temp struct {
		Name     string `json:"name"`
		Version  string `json:"ver"`
		Operator string `json:"operator"`
	}

	err := json.Unmarshal(data, &temp)
	if err == nil && temp.Operator != "" {
		d.Name = temp.Name
		d.Version = temp.Version
		d.Operator = temp.Operator
		return nil
	}
	var depMap map[string]interface{}
	err = json.Unmarshal(data, &depMap)
	if err != nil {
		return err
	}

	if name, ok := depMap["name"].(string); ok {
		d.Name = name
	} else {
		return errors.New("missing or invalid name field in dependency")
	}

	if version, ok := depMap["ver"].(string); ok {
		d.Version, d.Operator = extractVersionAndOperator(version)
	} else {
		return errors.New("missing or invalid ver field in dependency")
	}

	return nil
}

func (d *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var depMap map[string]interface{}
	err := unmarshal(&depMap)
	if err != nil {
		return err
	}

	if name, ok := depMap["name"].(string); ok {
		d.Name = name
	} else {
		return errors.New("missing or invalid name field in dependency")
	}

	if version, ok := depMap["ver"].(string); ok {
		d.Version, d.Operator = extractVersionAndOperator(version)
	} else {
		return errors.New("missing or invalid ver field in dependency")
	}

	return nil
}

func extractVersionAndOperator(version string) (string, string) {
	operators := []string{"==", "=", ">=", "<=", ">", "<"}

	for _, op := range operators {
		if strings.HasPrefix(version, op) {
			return strings.TrimPrefix(version, op), op
		}
	}

	return version, ""
}

// Custom unmarshaller for the Target struct
func (t *Target) UnmarshalJSON(data []byte) error {
	var path string
	if err := json.Unmarshal(data, &path); err == nil {
		t.Path = path
		t.Exclude = ""
		return nil
	}

	type targetAlias Target
	return json.Unmarshal(data, (*targetAlias)(t))
}

func (t *Target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var path string
	if err := unmarshal(&path); err == nil {
		t.Path = path
		t.Exclude = ""
		return nil
	}

	type targetAlias Target
	return unmarshal((*targetAlias)(t))
}

func readCreateFile(configFile string) (*Package, error) {
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(configFile)

	var pkg Package
	switch ext {
	case ".json":
		err = json.Unmarshal(fileData, &pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(fileData, &pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return &pkg, nil
}

func GetNameAndVersionFromConfigFile(configFilePath string) (string, string, error) {
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(configFilePath)

	var pkg Package
	switch ext {
	case ".json":
		err = json.Unmarshal(fileData, &pkg)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(fileData, &pkg)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	default:
		return "", "", fmt.Errorf("unsupported config file format: %s", ext)
	}

	return pkg.Name, pkg.Version, nil
}

func ReadUpdateFile(filePath string) (UpdateConfig, error) {
	var config UpdateConfig

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("failed to read update file: %w", err)
	}

	fileExt := strings.ToLower(fileExtension(filePath))

	switch fileExt {
	case ".json":
		if err := json.Unmarshal(fileContent, &config); err != nil {
			return config, fmt.Errorf("failed to parse update file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(fileContent, &config); err != nil {
			return config, fmt.Errorf("failed to parse update file: %w", err)
		}
	default:
		return config, fmt.Errorf("unsupported file format: %s", fileExt)
	}

	return config, nil
}

func fileExtension(filePath string) string {
	filename := strings.ToLower(filePath)
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}
	return filename[lastDot:]
}
