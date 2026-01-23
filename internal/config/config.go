package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Name         string `yaml:"name"`
	Prefix       string `yaml:"prefix"`
	Target       string `yaml:"target"`
	PreservePath bool   `yaml:"preserve_path"`
}

type Config struct {
	Server struct {
		Port      int    `yaml:"port"`
		AdminPath string `yaml:"admin_path"`
	} `yaml:"server"`
	Security struct {
		LumenSecret       string          `yaml:"lumen_secret"`
		IgnoredExtensions map[string]bool `yaml:"ignored_extensions"`
	} `yaml:"security"`
	Proxy []Proxy `yaml:"proxy"`
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
