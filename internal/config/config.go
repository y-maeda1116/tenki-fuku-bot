package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Categories struct {
	Men   bool `yaml:"men"`
	Women bool `yaml:"women"`
	Kids  bool `yaml:"kids"`
}

type Config struct {
	City       string     `yaml:"city"`
	Categories Categories `yaml:"categories"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return &cfg, nil
}
