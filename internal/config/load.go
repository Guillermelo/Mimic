package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadFile(path string) (Config, error) {
	ext := filepath.Ext(path)
	if ext != ".yml" && ext != ".yaml" {
		return Config{}, fmt.Errorf("config file must use .yml or .yaml extension")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
