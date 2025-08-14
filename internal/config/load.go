package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var kayPattern = regexp.MustCompile("^[A-Za-z0-9_]{3,}$")

func Load(fileName string) (TouchConfig, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return TouchConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TouchConfig
	ext := filepath.Ext(fileName)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config.Resources); err != nil {
			return TouchConfig{}, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config.Resources); err != nil {
			return TouchConfig{}, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return TouchConfig{}, fmt.Errorf("unsupported file format: %s", ext)
	}

	return config, nil
}
