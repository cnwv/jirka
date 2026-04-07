package config

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed config.example.yaml
var ExampleConfig []byte

// writeEmbeddedExample writes the embedded example config to a temp file and returns its path.
func writeEmbeddedExample() string {
	dir := os.TempDir()
	path := filepath.Join(dir, "jirka-example-config.yaml")
	_ = os.WriteFile(path, ExampleConfig, 0o600)
	return path
}
