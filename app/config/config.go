package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// ConfigDir returns the path to ~/.config/jirka/.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "jirka")
}

// Config holds the application configuration.
type Config struct {
	JiraURL      string
	JiraToken    string
	JiraAuthType string // "bearer" (default) or "basic"
	JiraEmail    string // required for basic auth (Jira Cloud)
	PollInterval time.Duration
	Dashboard    *DashboardConfig
}

// Load reads .env and config.yaml from standard locations.
// Search order for .env: ~/.config/jirka/.env, then ./.env
// Search order for config: $CONFIG_PATH, ~/.config/jirka/config.yaml, ./config.yaml, embedded example
func Load() (Config, error) {
	// Load .env from ~/.config/jirka/.env first, then local .env
	configDir := ConfigDir()
	if configDir != "" {
		_ = godotenv.Load(filepath.Join(configDir, ".env"))
	}
	_ = godotenv.Load() // local .env (overrides if set)

	authType := os.Getenv("JIRA_AUTH_TYPE")
	if authType == "" {
		authType = "bearer"
	}

	cfg := Config{
		JiraURL:      os.Getenv("JIRA_URL"),
		JiraToken:    os.Getenv("JIRA_TOKEN"),
		JiraAuthType: authType,
		JiraEmail:    os.Getenv("JIRA_EMAIL"),
		PollInterval: 5 * time.Minute,
	}

	if cfg.JiraURL == "" {
		cfg.JiraURL = "https://jira.example.com"
	}

	configPath, isExample := findConfig(configDir)

	dashboard, err := LoadDashboard(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config: %w\nRun `jirka init` to set up your configuration", err)
	}
	dashboard.IsExample = isExample
	if !isExample {
		dashboard.ConfigPath = configPath
	}
	cfg.Dashboard = dashboard

	return cfg, nil
}

// Configured returns true if a real config exists (not demo mode).
func Configured() bool {
	configDir := ConfigDir()
	_, isExample := findConfig(configDir)
	return !isExample
}

// findConfig returns the config path and whether it's the embedded example.
func findConfig(configDir string) (path string, isExample bool) {
	// 1. Explicit env override
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil { //nolint:gosec // config path from env
			return p, false
		}
	}

	// 2. ~/.config/jirka/config.yaml
	if configDir != "" {
		p := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p, false
		}
	}

	// 3. ./config.yaml (backward compat)
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml", false
	}

	// 4. ./config.example.yaml (local dev)
	if _, err := os.Stat("config.example.yaml"); err == nil {
		return "config.example.yaml", true
	}

	// 5. Embedded example (written to temp file)
	return writeEmbeddedExample(), true
}
