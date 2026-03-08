// Package config handles .helm-snoop.yaml config file parsing and resolution.
package config

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

const configVersion = 0

// fileConfig is the raw YAML structure of .helm-snoop.yaml.
type fileConfig struct {
	Version *int                       `yaml:"version"`
	Global  fileChartConfig            `yaml:"global"`
	Charts  map[string]fileChartConfig `yaml:"charts"`
}

// fileChartConfig holds per-chart (or global) settings from the config file.
type fileChartConfig struct {
	Ignore      []string       `yaml:"ignore"`
	ValuesFiles []string       `yaml:"valuesFiles"`
	ExtraValues map[string]any `yaml:"extraValues"`
}

func parse(data []byte) (*fileConfig, error) {
	var cfg fileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	if cfg.Version == nil {
		return nil, errors.New("config file missing required 'version' field")
	}
	if *cfg.Version != configVersion {
		return nil, fmt.Errorf("unsupported config version %d (expected %d)", *cfg.Version, configVersion)
	}
	return &cfg, nil
}
