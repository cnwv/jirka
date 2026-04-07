package tui

import (
	"strings"
	"testing"
)

func TestPopupBox_Dimensions(t *testing.T) {
	lines := []string{"line 1", "line 2", "line 3"}
	box := popupBox("Title", lines, 20)
	result := strings.Split(box, "\n")

	expectedLines := len(lines) + 2 // content + top + bottom border
	if len(result) != expectedLines {
		t.Errorf("expected %d lines, got %d", expectedLines, len(result))
	}

	// All lines should have same visible width
	expectedW := 20
	for i, line := range result {
		w := visibleWidth(line)
		if w != expectedW {
			t.Errorf("line %d: visible width = %d, want %d", i, w, expectedW)
		}
	}
}

func TestPopupBox_MinWidth(t *testing.T) {
	box := popupBox("T", []string{"a"}, 5)
	// Should be clamped to 10
	lines := strings.Split(box, "\n")
	for i, line := range lines {
		w := visibleWidth(line)
		if w != 10 {
			t.Errorf("line %d: visible width = %d, want 10 (min)", i, w)
		}
	}
}

func TestStripAnsiStr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello", "hello"},
		{"bold", "\033[1mhello\033[0m", "hello"},
		{"color", "\033[38;5;75mblue\033[0m", "blue"},
		{"mixed", "a\033[1mb\033[0mc", "abc"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsiStr(tt.input)
			if got != tt.want {
				t.Errorf("stripAnsiStr(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain", "hello", 5},
		{"with ansi", "\033[1mhello\033[0m", 5},
		{"empty", "", 0},
		{"mixed", "a\033[38;5;75mb\033[0mc", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleWidth(tt.input)
			if got != tt.want {
				t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateAnsi(t *testing.T) {
	input := "\033[1mhello world\033[0m"
	result := truncateAnsi(input, 5)
	w := visibleWidth(result)
	if w != 5 {
		t.Errorf("truncateAnsi width = %d, want 5", w)
	}
	plain := stripAnsiStr(result)
	if plain != "hello" {
		t.Errorf("truncated text = %q, want %q", plain, "hello")
	}
}

func TestOverlayCenter(t *testing.T) {
	main := strings.Repeat("........\n", 5)
	popup := "XX\nXX"
	result := overlayCenter(main, popup, 8, 5)

	lines := strings.Split(result, "\n")
	// The popup should appear somewhere in the middle
	found := false
	for _, line := range lines {
		if strings.Contains(line, "XX") {
			found = true
			break
		}
	}
	if !found {
		t.Error("popup content 'XX' not found in overlay result")
	}
}

func TestCenterTitle(t *testing.T) {
	result := centerTitle(" Test ", 20)
	w := visibleWidth(result)
	if w != 20 {
		t.Errorf("centerTitle width = %d, want 20", w)
	}
	if !strings.Contains(result, "Test") {
		t.Error("title not found in result")
	}
}
