package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveColor(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"red", "38;5;203"},
		{"blue", "38;5;75"},
		{"green", "38;5;114"},
		{"yellow", "38;5;222"},
		{"orange", "38;5;209"},
		{"purple", "38;5;177"},
		{"teal", "38;5;80"},
		{"cyan", "38;5;87"},
		{"RED", "38;5;203"},   // case insensitive
		{"38;5;42", "38;5;42"}, // raw ANSI passthrough
		{"", "38;5;245"},       // default
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveColor(tt.name)
			if got != tt.want {
				t.Errorf("resolveColor(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestLoadDashboard_MultiWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	err := os.WriteFile(path, []byte(`
windows:
  - name: Win1
    layout:
      rows: 2
      cols: 1
    panels:
      - title: A
        color: blue
      - title: B
        color: red
  - name: Win2
    layout:
      rows: 1
      cols: 1
    panels:
      - title: C
        color: green
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	dc, err := LoadDashboard(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(dc.Windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(dc.Windows))
	}
	if dc.Windows[0].Name != "Win1" {
		t.Errorf("window 0 name = %q, want Win1", dc.Windows[0].Name)
	}
	if dc.Windows[1].Name != "Win2" {
		t.Errorf("window 1 name = %q, want Win2", dc.Windows[1].Name)
	}
}

func TestLoadDashboard_LegacyFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.yaml")
	err := os.WriteFile(path, []byte(`
layout:
  rows: 1
  cols: 1
panels:
  - title: Legacy
    color: yellow
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	dc, err := LoadDashboard(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(dc.Windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(dc.Windows))
	}
	if dc.Windows[0].Name != "Main" {
		t.Errorf("legacy window name = %q, want Main", dc.Windows[0].Name)
	}
}

func TestLoadDashboard_InvalidGrid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	err := os.WriteFile(path, []byte(`
windows:
  - name: Bad
    layout:
      rows: 3
      cols: 3
    panels:
      - title: A
        color: blue
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadDashboard(path)
	if err == nil {
		t.Error("expected error for invalid grid (3x3 > 6 panels)")
	}
}

func TestLoadDashboard_PanelCountMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mismatch.yaml")
	err := os.WriteFile(path, []byte(`
windows:
  - name: X
    layout:
      rows: 2
      cols: 2
    panels:
      - title: A
        color: blue
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadDashboard(path)
	if err == nil {
		t.Error("expected error for panel count mismatch (2x2 but 1 panel)")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save_test.yaml")

	windows := []WindowConfig{
		{
			Name:   "Saved",
			Layout: GridConfig{Rows: 1, Cols: 1},
			Panels: []PanelConfig{{Title: "P1", Color: "blue"}},
		},
	}

	if err := SaveConfig(path, windows); err != nil {
		t.Fatal(err)
	}

	// Read it back
	dc, err := LoadDashboard(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(dc.Windows) != 1 || dc.Windows[0].Name != "Saved" {
		t.Errorf("save/load roundtrip failed: %+v", dc.Windows)
	}
}

func TestEffectiveGrid_Defaults(t *testing.T) {
	w := WindowConfig{} // no layout set
	rows, cols := w.EffectiveGrid()
	if rows != 3 || cols != 2 {
		t.Errorf("expected 3x2 default, got %dx%d", rows, cols)
	}
}
