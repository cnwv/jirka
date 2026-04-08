package tui

import (
	"fmt"
	"strings"

	"github.com/cnwv/jirka/app/config"
)

type panelEditPopup struct {
	panelIdx      int
	nameInput     textInput
	jqlInput      textInput
	colorIdx      int
	focusField    int    // 0=name, 1=jql(no sections)/color(has sections), 2=color(no sections)/sections(has sections)
	testStatus    string // "" | "testing..." | "✓ ..." | "✗ ..."
	sections      []config.SectionConfig
	sectionCursor int
}

func newPanelEditPopup(panelIdx int, pc config.PanelConfig) *panelEditPopup {
	p := &panelEditPopup{panelIdx: panelIdx}
	p.nameInput.maxLen = 10
	p.nameInput.set(pc.Title)
	p.jqlInput.set(pc.JQL)
	p.sections = append([]config.SectionConfig{}, pc.Sections...)
	for i, name := range config.ColorNames {
		if config.ResolveColorPublic(name) == pc.Color {
			p.colorIdx = i
			break
		}
	}
	return p
}

func (p *panelEditPopup) hasSections() bool {
	return len(p.sections) > 0
}

const panelEditFieldCount = 3 // name + jql/color + color/sections

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

// handleKey returns action and config.
// Actions: actionSave, actionTestJQL, actionClose, actionEditSection, actionAddSection.
func (p *panelEditPopup) handleKey(key string) (action string, pc config.PanelConfig) {
	switch key {
	case "esc":
		return actionClose, pc
	case "ctrl+s":
		return actionSave, p.build()
	case "tab":
		p.focusField = (p.focusField + 1) % panelEditFieldCount
		return "", pc
	case "shift+tab":
		p.focusField = (p.focusField + panelEditFieldCount - 1) % panelEditFieldCount
		return "", pc
	case "enter":
		return p.handleEnter()
	default:
		return p.handleInput(key)
	}
}

func (p *panelEditPopup) handleEnter() (string, config.PanelConfig) {
	var pc config.PanelConfig

	if p.hasSections() {
		switch p.focusField {
		case 0: // name → color
			p.focusField = 1
		case 1: // color → sections
			p.focusField = 2
		case 2: // sections → edit selected section
			if len(p.sections) > 0 {
				return actionEditSection, pc
			}
		}
	} else {
		switch p.focusField {
		case 0: // name → jql
			p.focusField = 1
		case 1: // jql → test
			if strings.TrimSpace(p.jqlInput.Value) != "" {
				p.setTesting()
				return actionTestJQL, pc
			}
		case 2: // color → save
			return actionSave, p.build()
		}
	}
	return "", pc
}

func (p *panelEditPopup) handleSectionListKey(key string) (string, config.PanelConfig) {
	var pc config.PanelConfig
	switch key {
	case "up", "k":
		if p.sectionCursor > 0 {
			p.sectionCursor--
		}
	case "down", "j":
		if p.sectionCursor < len(p.sections)-1 {
			p.sectionCursor++
		}
	case "a", "ф":
		return actionAddSection, pc
	case "d":
		if len(p.sections) > 0 {
			p.sections = append(p.sections[:p.sectionCursor], p.sections[p.sectionCursor+1:]...)
			if p.sectionCursor >= len(p.sections) && p.sectionCursor > 0 {
				p.sectionCursor--
			}
		}
	}
	return "", pc
}

func (p *panelEditPopup) handleInput(key string) (string, config.PanelConfig) {
	var pc config.PanelConfig

	if p.hasSections() {
		switch p.focusField {
		case 0: // name
			p.nameInput.handleKey(key)
		case 1: // color
			p.handleColorKey(key)
		case 2: // sections list
			return p.handleSectionListKey(key)
		}
	} else {
		switch p.focusField {
		case 0: // name
			p.nameInput.handleKey(key)
		case 1: // jql
			p.jqlInput.handleKey(key)
			p.testStatus = ""
		case 2: // color
			switch key {
			case "a", "ф": // add first section
				return actionAddSection, pc
			default:
				p.handleColorKey(key)
			}
		}
	}
	return "", pc
}

func (p *panelEditPopup) handleColorKey(key string) {
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

// handlePaste inserts pasted text into the focused field.
func (p *panelEditPopup) handlePaste(s string) {
	if p.hasSections() {
		if p.focusField == 0 {
			p.nameInput.insertString(s)
		}
	} else {
		switch p.focusField {
		case 0:
			p.nameInput.insertString(s)
		case 1:
			p.jqlInput.insertString(s)
			p.testStatus = ""
		}
	}
}

func (p *panelEditPopup) updateSection(idx int, sec config.SectionConfig) {
	if idx >= 0 && idx < len(p.sections) {
		p.sections[idx] = sec
	}
}

func (p *panelEditPopup) addSection(sec config.SectionConfig) {
	p.sections = append(p.sections, sec)
	p.sectionCursor = len(p.sections) - 1
	// Switch focus to sections list
	p.focusField = 2
}

func (p *panelEditPopup) build() config.PanelConfig {
	color := config.ResolveColorPublic(config.ColorNames[p.colorIdx])
	pc := config.PanelConfig{
		Title:    strings.TrimSpace(p.nameInput.Value),
		Color:    color,
		Sections: p.sections,
	}
	if !p.hasSections() {
		pc.JQL = strings.TrimSpace(p.jqlInput.Value)
	}
	return pc
}

func (p *panelEditPopup) currentJQL() string {
	return strings.TrimSpace(p.jqlInput.Value)
}

const (
	focusActive   = "\033[38;5;75m"
	focusInactive = "\033[38;5;245m"
	ansiReset     = "\033[0m"
)

func (p *panelEditPopup) focusColor(i int) string {
	if p.focusField == i {
		return focusActive
	}
	return focusInactive
}

func (p *panelEditPopup) viewSectionsContent(lines []string, inner int) []string {
	// Color selector (field 1)
	pre := p.focusColor(1)
	colorName := config.ColorNames[p.colorIdx]
	ansiCode := config.ResolveColorPublic(colorName)
	colorSwatch := fmt.Sprintf("\033[%sm●\033[0m %s", ansiCode, colorName)
	lines = append(lines,
		fmt.Sprintf("  %sColor:%s \033[38;5;242m◄\033[0m %s \033[38;5;242m►\033[0m", pre, ansiReset, colorSwatch),
		"",
	)

	// Sections list (field 2)
	secPre := p.focusColor(2)
	lines = append(lines, fmt.Sprintf("  %sSections:%s", secPre, ansiReset))
	for i, sec := range p.sections {
		cursor := "  "
		if p.focusField == 2 && i == p.sectionCursor {
			cursor = "\033[38;5;75m▸\033[0m "
		}
		secColor := sec.Color
		if secColor == "" {
			secColor = "38;5;245"
		}
		name := sec.Name
		if name == "" {
			name = "(unnamed)"
		}
		jqlPreview := sec.JQL
		if len(jqlPreview) > inner-len(name)-10 {
			jqlPreview = jqlPreview[:max(inner-len(name)-13, 0)] + "..."
		}
		lines = append(lines, fmt.Sprintf("  %s\033[1;%sm%s\033[0m  \033[38;5;239m%s\033[0m", cursor, secColor, name, jqlPreview))
	}
	if len(p.sections) == 0 {
		lines = append(lines, "    \033[38;5;242m(empty)\033[0m")
	}
	hint := " \033[38;5;242mTab: next  Enter: edit  a: add  d: del  Ctrl+S: save  Esc: cancel\033[0m"
	lines = append(lines, "", hint, "")
	return lines
}

func (p *panelEditPopup) viewNoSectionsContent(lines []string, inner int) []string {
	// JQL field (field 1)
	pre := p.focusColor(1)
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

	// Color selector (field 2)
	pre = p.focusColor(2)
	colorName := config.ColorNames[p.colorIdx]
	ansiCode := config.ResolveColorPublic(colorName)
	colorSwatch := fmt.Sprintf("\033[%sm●\033[0m %s", ansiCode, colorName)
	lines = append(lines,
		fmt.Sprintf("  %sColor:%s \033[38;5;242m◄\033[0m %s \033[38;5;242m►\033[0m", pre, ansiReset, colorSwatch),
		"",
		" \033[38;5;242mTab: next  Enter: next/test  a: add section  Ctrl+S: save  Esc: cancel\033[0m",
		"",
	)
	return lines
}

func (p *panelEditPopup) view(totalW, totalH int) string {
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
	limit := fmt.Sprintf("\033[38;5;239m%d/%d\033[0m", len([]rune(p.nameInput.Value)), p.nameInput.maxLen)
	lines = append(lines,
		fmt.Sprintf("  %sName:%s  %s %s", pre, ansiReset, nameContent, limit),
		"",
	)

	if p.hasSections() {
		lines = p.viewSectionsContent(lines, inner)
	} else {
		lines = p.viewNoSectionsContent(lines, inner)
	}

	box := popupBox(fmt.Sprintf("Edit Panel %d", p.panelIdx+1), lines, popupW)
	return overlayCenter(strings.Repeat("\n", totalH), box, totalW, totalH)
}
