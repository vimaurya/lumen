package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Name         string   `yaml:"name"`
	Prefix       string   `yaml:"prefix"`
	Targets      []string `yaml:"target"`
	PreservePath bool     `yaml:"preserve_path"`
}

type Security struct {
	LumenSecret       string          `yaml:"lumen_secret"`
	IgnoredExtensions map[string]bool `yaml:"ignored_extensions"`
}

type Server struct {
	Port      int    `yaml:"port"`
	AdminPath string `yaml:"admin_path"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Security Security `yaml:"security"`
	Proxy    []Proxy  `yaml:"proxy"`
}

func LoadConfig(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	return &cfg, err
}
