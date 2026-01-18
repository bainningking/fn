package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Agent  AgentConfig  `yaml:"agent"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	TLS     bool   `yaml:"tls"`
}

type AgentConfig struct {
	ID              string `yaml:"id"`
	CollectInterval int    `yaml:"collect_interval"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
