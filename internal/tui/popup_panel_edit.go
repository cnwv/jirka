package tui

import (
	"buble_jira/internal/config"
	"fmt"
	"strings"
)

type panelEditPopup struct {
	panelIdx   int
	nameInput  textInput
	jqlInput   textInput
	colorIdx   int
	focusField int // 0=name, 1=jql, 2=color
	testStatus string // "" | "testing..." | "✓ ..." | "✗ ..."
}

func newPanelEditPopup(panelIdx int, pc config.PanelConfig) *panelEditPopup {
	p := &panelEditPopup{panelIdx: panelIdx}
	p.nameInput.maxLen = 10
	p.nameInput.set(pc.Title)
	p.jqlInput.set(pc.JQL)
	for i, name := range config.ColorNames {
		if config.ResolveColorPublic(name) == pc.Color {
			p.colorIdx = i
			break
		}
	}
	return p
}

func (p *panelEditPopup) setTestResult(ok bool, summary string) {
	if ok {
		p.testStatus = "\033[38;5;114m✓ " + summary + "\033[0m"
	} else {
		p.testStatus = "\033[38;5;203m✗ " + summary + "\033[0m"
	}
}

func (p *panelEditPopup) setTesting() {
	p.testStatus = "\033[38;5;245mtesting...\033[0m"
}

// handleKey returns ("save", pc), ("test_jql", {}), or ("close", {}).
func (p *panelEditPopup) handleKey(key string) (action string, pc config.PanelConfig) {
	switch key {
	case "esc":
		return "close", pc
	case "ctrl+s":
		return "save", p.build()
	case "tab":
		p.focusField = (p.focusField + 1) % 3
		return "", pc
	case "shift+tab":
		p.focusField = (p.focusField + 2) % 3
		return "", pc
	case "enter":
		switch p.focusField {
		case 0: // name → move to jql
			p.focusField = 1
		case 1: // jql → test query
			if strings.TrimSpace(p.jqlInput.Value) != "" {
				p.setTesting()
				return "test_jql", pc
			}
		case 2: // color → save
			return "save", p.build()
		}
		return "", pc
	default:
		switch p.focusField {
		case 0:
			p.nameInput.handleKey(key)
		case 1:
			p.jqlInput.handleKey(key)
			p.testStatus = "" // clear stale test result on edit
		case 2:
			switch key {
			case "left", "h":
				if p.colorIdx > 0 {
					p.colorIdx--
				}
			case "right", "l":
				if p.colorIdx < len(config.ColorNames)-1 {
					p.colorIdx++
				}
			}
		}
	}
	return "", pc
}

// handlePaste inserts pasted text into the focused field.
func (p *panelEditPopup) handlePaste(s string) {
	switch p.focusField {
	case 0:
		p.nameInput.insertString(s)
	case 1:
		p.jqlInput.insertString(s)
		p.testStatus = ""
	}
}

func (p *panelEditPopup) build() config.PanelConfig {
	color := config.ResolveColorPublic(config.ColorNames[p.colorIdx])
	return config.PanelConfig{
		Title: strings.TrimSpace(p.nameInput.Value),
		Color: color,
		JQL:   strings.TrimSpace(p.jqlInput.Value),
	}
}

func (p *panelEditPopup) currentJQL() string {
	return strings.TrimSpace(p.jqlInput.Value)
}

func (p *panelEditPopup) view(totalW, totalH int) string {
	const popupW = 62
	inner := popupW - 4

	var lines []string
	lines = append(lines, "")

	focusColor := func(i int) (pre, post string) {
		if p.focusField == i {
			return "\033[38;5;75m", "\033[0m"
		}
		return "\033[38;5;245m", "\033[0m"
	}

	// Name field
	pre, post := focusColor(0)
	var nameContent string
	if p.focusField == 0 {
		nameContent = p.nameInput.view(inner - 8)
	} else {
		nameContent = p.nameInput.viewReadonly(inner - 8)
	}
	limit := fmt.Sprintf("\033[38;5;239m%d/%d\033[0m", len([]rune(p.nameInput.Value)), p.nameInput.maxLen)
	lines = append(lines, fmt.Sprintf("  %sName:%s  %s %s", pre, post, nameContent, limit))
	lines = append(lines, "")

	// JQL field
	pre, post = focusColor(1)
	jqlW := inner - 8
	var jqlContent string
	if p.focusField == 1 {
		jqlContent = p.jqlInput.view(jqlW)
	} else {
		jqlContent = p.jqlInput.viewReadonly(jqlW)
	}
	lines = append(lines, fmt.Sprintf("  %sJQL:%s   %s", pre, post, jqlContent))

	// Test status line
	if p.testStatus != "" {
		lines = append(lines, "         "+p.testStatus)
	} else {
		lines = append(lines, "  \033[38;5;239mEnter on JQL field to test query\033[0m")
	}
	lines = append(lines, "")

	// Color selector
	pre, post = focusColor(2)
	colorName := config.ColorNames[p.colorIdx]
	ansiCode := config.ResolveColorPublic(colorName)
	colorSwatch := fmt.Sprintf("\033[%sm●\033[0m %s", ansiCode, colorName)
	lines = append(lines, fmt.Sprintf("  %sColor:%s \033[38;5;242m◄\033[0m %s \033[38;5;242m►\033[0m", pre, post, colorSwatch))
	lines = append(lines, "")

	lines = append(lines, " \033[38;5;242mTab: next  Enter: next/test  Ctrl+S: save  Esc: cancel\033[0m")
	lines = append(lines, "")

	box := popupBox(fmt.Sprintf("Edit Panel %d", p.panelIdx+1), lines, popupW)
	return overlayCenter(strings.Repeat("\n", totalH), box, totalW, totalH)
}
