package tui

import (
	"fmt"
	"strings"

	"github.com/cnwv/jirka/app/config"
)

type sectionEditPopup struct {
	panelIdx   int
	sectionIdx int // -1 for new section
	nameInput  textInput
	jqlInput   textInput
	colorIdx   int
	focusField int // 0=name, 1=jql, 2=color
	testStatus string
}

func newSectionEditPopup(panelIdx, sectionIdx int, sec config.SectionConfig) *sectionEditPopup {
	p := &sectionEditPopup{panelIdx: panelIdx, sectionIdx: sectionIdx}
	p.nameInput.maxLen = 20
	p.nameInput.set(sec.Name)
	p.jqlInput.set(sec.JQL)
	for i, name := range config.ColorNames {
		if config.ResolveColorPublic(name) == sec.Color {
			p.colorIdx = i
			break
		}
	}
	return p
}

func (p *sectionEditPopup) setTestResult(ok bool, summary string) {
	if ok {
		p.testStatus = "\033[38;5;114m✓ " + summary + "\033[0m"
	} else {
		p.testStatus = "\033[38;5;203m✗ " + summary + "\033[0m"
	}
}

func (p *sectionEditPopup) setTesting() {
	p.testStatus = "\033[38;5;245mtesting...\033[0m"
}

func (p *sectionEditPopup) handleKey(key string) (action string, sec config.SectionConfig) {
	switch key {
	case "esc":
		return actionClose, sec
	case "ctrl+s":
		return actionSave, p.build()
	case "tab":
		p.focusField = (p.focusField + 1) % 3
		return "", sec
	case "shift+tab":
		p.focusField = (p.focusField + 2) % 3
		return "", sec
	case "enter":
		switch p.focusField {
		case 0:
			p.focusField = 1
		case 1:
			if strings.TrimSpace(p.jqlInput.Value) != "" {
				p.setTesting()
				return actionTestJQL, sec
			}
		case 2:
			return actionSave, p.build()
		}
		return "", sec
	default:
		switch p.focusField {
		case 0:
			p.nameInput.handleKey(key)
		case 1:
			p.jqlInput.handleKey(key)
			p.testStatus = ""
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
	return "", sec
}

func (p *sectionEditPopup) handlePaste(s string) {
	switch p.focusField {
	case 0:
		p.nameInput.insertString(s)
	case 1:
		p.jqlInput.insertString(s)
		p.testStatus = ""
	}
}

func (p *sectionEditPopup) build() config.SectionConfig {
	color := config.ResolveColorPublic(config.ColorNames[p.colorIdx])
	return config.SectionConfig{
		Name:  strings.TrimSpace(p.nameInput.Value),
		Color: color,
		JQL:   strings.TrimSpace(p.jqlInput.Value),
	}
}

func (p *sectionEditPopup) currentJQL() string {
	return strings.TrimSpace(p.jqlInput.Value)
}

func (p *sectionEditPopup) focusColor(i int) string {
	if p.focusField == i {
		return focusActive
	}
	return focusInactive
}

func (p *sectionEditPopup) view(totalW, totalH int) string {
	const popupW = 62
	inner := popupW - 4

	lines := []string{""}

	// Name field
	pre := p.focusColor(0)
	var nameContent string
	if p.focusField == 0 {
		nameContent = p.nameInput.view(inner - 8)
	} else {
		nameContent = p.nameInput.viewReadonly(inner - 8)
	}
	lines = append(lines,
		fmt.Sprintf("  %sName:%s  %s", pre, ansiReset, nameContent),
		"",
	)

	// JQL field
	pre = p.focusColor(1)
	jqlW := inner - 8
	var jqlContent string
	if p.focusField == 1 {
		jqlContent = p.jqlInput.view(jqlW)
	} else {
		jqlContent = p.jqlInput.viewReadonly(jqlW)
	}
	lines = append(lines, fmt.Sprintf("  %sJQL:%s   %s", pre, ansiReset, jqlContent))

	if p.testStatus != "" {
		lines = append(lines, "         "+p.testStatus)
	} else {
		lines = append(lines, "  \033[38;5;239mEnter on JQL field to test query\033[0m")
	}
	lines = append(lines, "")

	// Color selector
	pre = p.focusColor(2)
	colorName := config.ColorNames[p.colorIdx]
	ansiCode := config.ResolveColorPublic(colorName)
	colorSwatch := fmt.Sprintf("\033[%sm●\033[0m %s", ansiCode, colorName)
	lines = append(lines,
		fmt.Sprintf("  %sColor:%s \033[38;5;242m◄\033[0m %s \033[38;5;242m►\033[0m", pre, ansiReset, colorSwatch),
		"",
		" \033[38;5;242mTab: next  Enter: next/test  Ctrl+S: save  Esc: back\033[0m",
		"",
	)

	title := "New Section"
	if p.sectionIdx >= 0 {
		title = fmt.Sprintf("Edit Section %d", p.sectionIdx+1)
	}
	box := popupBox(title, lines, popupW)
	return overlayCenter(strings.Repeat("\n", totalH), box, totalW, totalH)
}
