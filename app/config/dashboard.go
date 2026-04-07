package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type GridConfig struct {
	Rows int `yaml:"rows"`
	Cols int `yaml:"cols"`
}

// WindowConfig represents a single dashboard window with its own layout and panels.
type WindowConfig struct {
	Name   string        `yaml:"name"`
	Layout GridConfig    `yaml:"layout,omitempty"`
	Panels []PanelConfig `yaml:"panels"`
}

func (w WindowConfig) EffectiveGrid() (rows, cols int) {
	r, c := w.Layout.Rows, w.Layout.Cols
	if r <= 0 || c <= 0 {
		return 3, 2
	}
	return r, c
}

// DashboardConfig supports both legacy (layout+panels) and multi-window formats.
type DashboardConfig struct {
	// Multi-window format
	Windows []WindowConfig `yaml:"windows,omitempty"`
	// Legacy single-window format (backward compat)
	Layout GridConfig    `yaml:"layout,omitempty"`
	Panels []PanelConfig `yaml:"panels,omitempty"`

	IsExample  bool   `yaml:"-"`
	ConfigPath string `yaml:"-"` // path to save changes back to
}

// EffectiveGrid on DashboardConfig is kept for backward compat with tests.
func (dc *DashboardConfig) EffectiveGrid() (rows, cols int) {
	if len(dc.Windows) > 0 {
		return dc.Windows[0].EffectiveGrid()
	}
	r, c := dc.Layout.Rows, dc.Layout.Cols
	if r <= 0 || c <= 0 {
		return 3, 2
	}
	return r, c
}

type PanelConfig struct {
	Title    string          `yaml:"title"`
	Color    string          `yaml:"color"`
	JQL      string          `yaml:"jql,omitempty"`
	Sections []SectionConfig `yaml:"sections,omitempty"`
	Stubs    []StubTicket    `yaml:"stubs,omitempty"`
}

type SectionConfig struct {
	Name  string       `yaml:"name"`
	Color string       `yaml:"color"`
	JQL   string       `yaml:"jql,omitempty"`
	Stubs []StubTicket `yaml:"stubs,omitempty"`
}

type StubTicket struct {
	Key      string `yaml:"key"`
	Summary  string `yaml:"summary"`
	Status   string `yaml:"status"`
	Priority string `yaml:"priority"`
	Type     string `yaml:"type"`
}

var colorMap = map[string]string{
	"red":    "38;5;203",
	"blue":   "38;5;75",
	"green":  "38;5;114",
	"yellow": "38;5;222",
	"orange": "38;5;209",
	"purple": "38;5;177",
	"teal":   "38;5;80",
	"cyan":   "38;5;87",
}

// ColorNames returns named colors available for panel configuration.
var ColorNames = []string{"blue", "green", "yellow", "orange", "red", "purple", "teal", "cyan"}

// ResolveColorPublic is exported for use in TUI color pickers.
func ResolveColorPublic(name string) string {
	return resolveColor(name)
}

func resolveColor(name string) string {
	if c, ok := colorMap[strings.ToLower(name)]; ok {
		return c
	}
	if name != "" {
		return name
	}
	return "38;5;245"
}

func LoadDashboard(path string) (*DashboardConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is from env or hardcoded default
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var dc DashboardConfig
	if err := yaml.Unmarshal(data, &dc); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Normalize legacy format → multi-window
	if len(dc.Windows) == 0 {
		if len(dc.Panels) == 0 {
			return nil, errors.New("config must define either 'windows' or 'panels'")
		}
		dc.Windows = []WindowConfig{{
			Name:   "Main",
			Layout: dc.Layout,
			Panels: dc.Panels,
		}}
	}

	// Validate and resolve colors for each window
	for wi := range dc.Windows {
		w := &dc.Windows[wi]
		rows, cols := w.EffectiveGrid()
		expected := rows * cols
		if rows < 1 || cols < 1 || expected > 6 {
			return nil, fmt.Errorf("window %q: layout %dx%d is invalid: total panels must be 1–6", w.Name, rows, cols)
		}
		if len(w.Panels) != expected {
			return nil, fmt.Errorf(
				"window %q: layout %dx%d (%d panels) but %d panels provided",
				w.Name, rows, cols, expected, len(w.Panels),
			)
		}
		for i := range w.Panels {
			w.Panels[i].Color = resolveColor(w.Panels[i].Color)
			for j := range w.Panels[i].Sections {
				w.Panels[i].Sections[j].Color = resolveColor(w.Panels[i].Sections[j].Color)
			}
		}
	}

	return &dc, nil
}

// SaveConfig writes the windows list to path in multi-window YAML format.
func SaveConfig(path string, windows []WindowConfig) error {
	type saveFormat struct {
		Windows []WindowConfig `yaml:"windows"`
	}
	data, err := yaml.Marshal(saveFormat{Windows: windows})
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}
