package tui

import (
	"fmt"
	"strconv"
	"strings"
)

type winSwitcherPopup struct {
	names    []string
	selected int
}

func newWinSwitcherPopup(names []string, active int) *winSwitcherPopup {
	return &winSwitcherPopup{names: names, selected: active}
}

// handleKey returns (action, data).
// action: "switch"→selected idx, "delete"→selected idx, "new"→open new window, "close"→nil
func (p *winSwitcherPopup) handleKey(key string) (action string, idx int) {
	switch key {
	case "up", "k":
		if p.selected > 0 {
			p.selected--
		}
	case "down", "j":
		if p.selected < len(p.names)-1 {
			p.selected++
		}
	case "enter":
		return "switch", p.selected
	case "n":
		return "new", 0
	case "d", "backspace":
		return "delete", p.selected
	case "esc", "0":
		return "close", 0
	}
	return "", 0
}

func (p *winSwitcherPopup) view(totalW, totalH int) string {
	const popupW = 50

	lines := make([]string, 0, len(p.names)+2)

	for i, name := range p.names {
		cursor := "  "
		style := ""
		reset := ""
		if i == p.selected {
			cursor = "\033[38;5;75m▸\033[0m "
			style = "\033[1m"
			reset = "\033[0m"
		}
		num := strconv.Itoa(i + 1)
		lines = append(lines, fmt.Sprintf(" %s%s%s. %s%s", cursor, style, num, name, reset))
	}

	lines = append(lines,
		"",
		" \033[38;5;242m↑↓ navigate  Enter: switch  n: new  d: delete  Esc: close\033[0m",
	)

	box := popupBox("Windows", lines, popupW)
	return overlayCenter(strings.Repeat("\n", totalH), box, totalW, totalH)
}
