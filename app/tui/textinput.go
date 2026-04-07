package tui

import (
	"strings"
	"unicode/utf8"
)

// textInput is a simple single-line text input widget.
type textInput struct {
	Value  string
	cursor int
	maxLen int // 0 = unlimited
}

func (t *textInput) insert(ch rune) {
	if t.maxLen > 0 && len([]rune(t.Value)) >= t.maxLen {
		return
	}
	runes := []rune(t.Value)
	runes = append(runes[:t.cursor], append([]rune{ch}, runes[t.cursor:]...)...)
	t.Value = string(runes)
	t.cursor++
}

// insertString inserts a multi-character string (e.g. from paste), respecting maxLen.
func (t *textInput) insertString(s string) {
	for _, r := range s {
		if r < 32 || r == 127 {
			continue // skip control chars
		}
		if t.maxLen > 0 && len([]rune(t.Value)) >= t.maxLen {
			break
		}
		t.insert(r)
	}
}

func (t *textInput) backspace() {
	if t.cursor == 0 {
		return
	}
	runes := []rune(t.Value)
	t.Value = string(append(runes[:t.cursor-1], runes[t.cursor:]...))
	t.cursor--
}

func (t *textInput) deleteForward() {
	runes := []rune(t.Value)
	if t.cursor >= len(runes) {
		return
	}
	t.Value = string(append(runes[:t.cursor], runes[t.cursor+1:]...))
}

func (t *textInput) moveLeft() {
	if t.cursor > 0 {
		t.cursor--
	}
}

func (t *textInput) moveRight() {
	if t.cursor < len([]rune(t.Value)) {
		t.cursor++
	}
}

func (t *textInput) home() { t.cursor = 0 }
func (t *textInput) end()  { t.cursor = len([]rune(t.Value)) }

// view renders the input with a block cursor, truncated to width if needed.
// width=0 means no truncation.
func (t *textInput) view(width int) string {
	runes := []rune(t.Value)
	prefix := string(runes[:t.cursor])
	var cur, suffix string
	if t.cursor < len(runes) {
		cur = string(runes[t.cursor])
		suffix = string(runes[t.cursor+1:])
	} else {
		cur = " "
	}
	raw := prefix + "\033[7m" + cur + "\033[0m" + suffix
	if width > 0 && visibleWidth(raw) > width {
		raw = t.viewTruncated(runes, width)
	}
	return raw
}

func (t *textInput) viewTruncated(runes []rune, width int) string {
	visible := width - 1
	start := t.cursor - visible/2
	start = max(start, 0)
	end := start + visible
	if end > len(runes) {
		end = len(runes)
		start = max(end-visible, 0)
	}
	p := string(runes[start:t.cursor])
	var c, s string
	if t.cursor < end {
		c = string(runes[t.cursor])
		s = string(runes[t.cursor+1 : end])
	} else {
		c = " "
	}
	return p + "\033[7m" + c + "\033[0m" + s
}

// viewReadonly renders the value without a cursor (when field is not focused).
func (t *textInput) viewReadonly(width int) string {
	runes := []rune(t.Value)
	if width > 0 && len(runes) > width {
		return string(runes[:width])
	}
	return t.Value
}

// set sets the value and moves cursor to end.
func (t *textInput) set(v string) {
	if t.maxLen > 0 {
		runes := []rune(v)
		if len(runes) > t.maxLen {
			v = string(runes[:t.maxLen])
		}
	}
	t.Value = v
	t.cursor = len([]rune(v))
}

// handleKey processes a key string from tea.KeyPressMsg.String().
// Returns true if the key was consumed.
func (t *textInput) handleKey(key string) {
	switch key {
	case "left", "ctrl+b":
		t.moveLeft()
	case "right", "ctrl+f":
		t.moveRight()
	case "home", "ctrl+a":
		t.home()
	case "end", "ctrl+e":
		t.end()
	case "backspace", "ctrl+h":
		t.backspace()
	case "delete", "ctrl+d":
		t.deleteForward()
	case "ctrl+k":
		runes := []rune(t.Value)
		t.Value = string(runes[:t.cursor])
	case "ctrl+u":
		runes := []rune(t.Value)
		t.Value = string(runes[t.cursor:])
		t.cursor = 0
	default:
		t.handleCharKey(key)
	}
}

func (t *textInput) handleCharKey(key string) {
	if len(key) == 1 {
		r := rune(key[0])
		if r >= 32 && r != 127 {
			t.insert(r)
			return
		}
	}
	r, _ := utf8.DecodeRuneInString(key)
	if r != utf8.RuneError && r >= 32 && key != "esc" && key != "tab" && key != "enter" {
		t.insert(r)
		return
	}
	if strings.HasPrefix(key, "shift+") && len(key) == 7 {
		r := rune(key[6])
		if r >= 'a' && r <= 'z' {
			t.insert(r - 32)
		}
	}
}
