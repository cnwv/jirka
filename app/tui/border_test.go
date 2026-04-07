package tui

import (
	"strings"
	"testing"
)

func TestBorderedBox_Dimensions(t *testing.T) {
	sizes := [][2]int{
		{20, 5},
		{40, 10},
		{60, 15},
		{80, 20},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		content := []string{"line 1", "line 2", "line 3"}
		result := borderedBox("Test", "38;5;75", content, w, h, false)

		lines := strings.Split(result, "\n")
		if len(lines) != h {
			t.Errorf("%dx%d: expected %d lines, got %d", w, h, h, len(lines))
		}
		for i, line := range lines {
			vw := visibleWidth(line)
			if vw != w {
				t.Errorf("%dx%d line %d: visible width = %d, want %d", w, h, i, vw, w)
			}
		}
	}
}

func TestBorderedBox_Empty(t *testing.T) {
	result := borderedBox("T", "0", nil, 2, 2, false)
	if result != "" {
		t.Errorf("expected empty string for too-small box, got %q", result)
	}
}

func TestBorderedBox_FocusedColor(t *testing.T) {
	focused := borderedBox("T", "0", nil, 20, 5, true)
	unfocused := borderedBox("T", "0", nil, 20, 5, false)
	// Focused should use brighter border
	if !strings.Contains(focused, borderColorFocused) {
		t.Error("focused box should contain focused border color")
	}
	if !strings.Contains(unfocused, borderColorDefault) {
		t.Error("unfocused box should contain default border color")
	}
}
