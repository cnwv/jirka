package tui

import (
	"buble_jira/internal/config"
	"fmt"
	"strings"
)

type gridPreset struct {
	rows, cols int
}

var gridPresets = []gridPreset{
	{1, 1},
	{1, 2},
	{2, 1},
	{2, 2},
	{3, 1},
	{1, 3},
	{2, 3},
	{3, 2},
}

type newWindowPopup struct {
	step      int // 0=name, 1=grid
	nameInput textInput
	gridSel   int
}

func newNewWindowPopup() *newWindowPopup {
	p := &newWindowPopup{}
	p.nameInput.maxLen = 15
	return p
}

// handlePaste inserts pasted text (only on step 0).
func (p *newWindowPopup) handlePaste(s string) {
	if p.step == 0 {
		p.nameInput.insertString(s)
	}
}

// handleKey returns "done" with a built WindowConfig, "back" to go to step 0, or "close".
// Also mutates internal state.
func (p *newWindowPopup) handleKey(key string) (action string, win config.WindowConfig) {
	switch p.step {
	case 0: // name input
		switch key {
		case "enter":
			if strings.TrimSpace(p.nameInput.Value) != "" {
				p.step = 1
			}
		case "esc":
			return "close", win
		default:
			p.nameInput.handleKey(key)
		}
	case 1: // grid picker
		switch key {
		case "up", "k":
			if p.gridSel > 0 {
				p.gridSel--
			}
		case "down", "j":
			if p.gridSel < len(gridPresets)-1 {
				p.gridSel++
			}
		case "enter":
			preset := gridPresets[p.gridSel]
			return "done", p.buildWindow(preset.rows, preset.cols)
		case "esc":
			p.step = 0
			return "back", win
		}
	}
	return "", win
}

func (p *newWindowPopup) buildWindow(rows, cols int) config.WindowConfig {
	count := rows * cols
	panels := make([]config.PanelConfig, count)
	for i := range panels {
		panels[i] = config.PanelConfig{
			Title: fmt.Sprintf("Panel %d", i+1),
			Color: "38;5;245",
		}
	}
	return config.WindowConfig{
		Name:   strings.TrimSpace(p.nameInput.Value),
		Layout: config.GridConfig{Rows: rows, Cols: cols},
		Panels: panels,
	}
}

func (p *newWindowPopup) view(totalW, totalH int) string {
	const popupW = 52

	var lines []string
	lines = append(lines, "")

	switch p.step {
	case 0:
		limit := fmt.Sprintf("\033[38;5;239m%d/%d\033[0m", len([]rune(p.nameInput.Value)), p.nameInput.maxLen)
		lines = append(lines, "  Name: "+p.nameInput.view(popupW-10)+" "+limit)
		lines = append(lines, "")
		lines = append(lines, " \033[38;5;242mEnter: next  Esc: cancel\033[0m")

	case 1:
		lines = append(lines, fmt.Sprintf("  Name: \033[1m%s\033[0m", p.nameInput.Value))
		lines = append(lines, "")
		lines = append(lines, "  Layout:")
		for i, pr := range gridPresets {
			cursor := "  "
			style, reset := "", ""
			if i == p.gridSel {
				cursor = "\033[38;5;75m▸\033[0m "
				style = "\033[1m"
				reset = "\033[0m"
			}
			count := pr.rows * pr.cols
			noun := "panels"
			if count == 1 {
				noun = "panel"
			}
			lines = append(lines, fmt.Sprintf("  %s%s%d×%d  (%d %s)%s", cursor, style, pr.rows, pr.cols, count, noun, reset))
		}
		lines = append(lines, "")
		lines = append(lines, " \033[38;5;242m↑↓ select  Enter: create  Esc: back\033[0m")
	}

	lines = append(lines, "")
	box := popupBox("New Window", lines, popupW)
	return overlayCenter(strings.Repeat("\n", totalH), box, totalW, totalH)
}
