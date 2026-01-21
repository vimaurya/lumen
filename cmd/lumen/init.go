package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vimaurya/lumen/internal/config"
	"gopkg.in/yaml.v3"
)

func runInit() {
	var target string
	prompt := &survey.Input{
		Message: "What is your target backend URL? (e.g., http://localhost:8081)",
		Default: "http://localhost:8081",
	}
	survey.AskOne(prompt, &target)

	var secret string
	secretPrompt := &survey.Password{
		Message: "Set a secret token for proxy-to-backend security:",
	}
	survey.AskOne(secretPrompt, &secret)

	cfg := config.Config{}

	cfg.Server.Port = 8080
	cfg.Server.AdminPath = "/admin"

	cfg.Security.IgnoredExtensions = map[string]bool{
		".js":    true,
		".css":   true,
		".map":   true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".ico":   true,
		".svg":   true,
		".woff":  true,
		".woff2": true,
		".json":  true,
	}
	cfg.Security.LumenSecret = secret

	cfg.Proxy = append(cfg.Proxy, config.Proxy{
		Name:         "nimbus",
		Target:       target,
		PreservePath: true,
	})

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		log.Printf("error encoding config : %v", err)
	}

	os.WriteFile("config.yaml", []byte(yamlData), 0o644)
	fmt.Println("Configuration generated successfully!")
}
