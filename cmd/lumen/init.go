package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/vimaurya/lumen/internal/config"
	"gopkg.in/yaml.v3"
)

func runInit() {
	var targetsInput string
	targetPrompt := &survey.Input{
		Message: "Enter target backend URLs (comma-separated):",
		Default: "http://localhost:8081, http://localhost:8082",
	}
	survey.AskOne(targetPrompt, &targetsInput)

	rawTargets := strings.Split(targetsInput, ",")
	var targets []string
	for _, t := range rawTargets {
		targets = append(targets, strings.TrimSpace(t))
	}

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
		Targets:      targets,
		PreservePath: true,
	})

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		log.Printf("error encoding config : %v", err)
	}

	os.WriteFile("config.yaml", []byte(yamlData), 0o644)
	fmt.Println("Configuration generated successfully!")
}
