package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	JiraURL      string
	JiraToken    string
	PollInterval time.Duration
	Dashboard    *DashboardConfig
}


func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		JiraURL:      os.Getenv("JIRA_URL"),
		JiraToken:    os.Getenv("JIRA_TOKEN"),
		PollInterval: 5 * time.Minute,
	}

	if cfg.JiraURL == "" {
		cfg.JiraURL = "https://jira.example.com"
	}

	configPath := "config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		configPath = p
	}

	isExample := false
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.example.yaml"
		isExample = true
	}

	dashboard, err := LoadDashboard(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config: %w\nCopy config.example.yaml to config.yaml and customize it", err)
	}
	dashboard.IsExample = isExample
	if !isExample {
		dashboard.ConfigPath = configPath
	}
	cfg.Dashboard = dashboard

	return cfg, nil
}
