package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Create a temp dir with config.example.yaml
	dir := t.TempDir()
	example := filepath.Join(dir, "config.example.yaml")
	err := os.WriteFile(example, []byte(`
windows:
  - name: Test
    layout:
      rows: 1
      cols: 1
    panels:
      - title: P1
        color: blue
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Run from temp dir so it picks up example config
	origDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	_ = os.Chdir(dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.JiraURL != "https://jira.example.com" {
		t.Errorf("expected default JiraURL, got %q", cfg.JiraURL)
	}
	if cfg.JiraToken != "" {
		t.Errorf("expected empty JiraToken, got %q", cfg.JiraToken)
	}
	if cfg.Dashboard == nil {
		t.Fatal("expected dashboard to be loaded")
	}
	if !cfg.Dashboard.IsExample {
		t.Error("expected IsExample to be true")
	}
}

func TestLoad_WithConfigYaml(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgFile, []byte(`
windows:
  - name: Main
    layout:
      rows: 1
      cols: 2
    panels:
      - title: A
        color: red
      - title: B
        color: green
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	_ = os.Chdir(dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Dashboard.IsExample {
		t.Error("expected IsExample to be false when config.yaml exists")
	}
	if len(cfg.Dashboard.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(cfg.Dashboard.Windows))
	}
}
