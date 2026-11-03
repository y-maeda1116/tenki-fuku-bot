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
	// AreaCode は JMA の地方コード (例: 130000 = 東京)。
	// WeatherArea / TempArea は JMA レスポンス内のエリア名。
	AreaCode    string     `yaml:"areaCode"`
	WeatherArea string     `yaml:"weatherArea"`
	TempArea    string     `yaml:"tempArea"`
	Categories  Categories `yaml:"categories"`
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
