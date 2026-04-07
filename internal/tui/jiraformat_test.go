package tui

import (
	"strings"
	"testing"
)

func TestFormatJiraMarkup(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // substring that should be present (after stripping ANSI)
		notWant string // substring that should NOT be present
	}{
		{
			name:    "bold markup {*}",
			input:   "{*}Severity{*}: major",
			want:    "Severity: major",
			notWant: "{*}",
		},
		{
			name:    "italic markup {_}",
			input:   "{_}For example{_}, test",
			want:    "For example, test",
			notWant: "{_}",
		},
		{
			name:    "monospace {{}}",
			input:   "table {{game_rounds}}",
			want:    "table game_rounds",
			notWant: "{{",
		},
		{
			name:    "link with text",
			input:   "[report|https://example.com/report]",
			want:    "report (https://example.com/report)",
			notWant: "|",
		},
		{
			name:    "bare link",
			input:   "[https://example.com]",
			want:    "https://example.com",
			notWant: "[",
		},
		{
			name:    "image",
			input:   "!screenshot.png|width=100!",
			want:    "[image: screenshot.png]",
			notWant: "width=",
		},
		{
			name:    "color markup",
			input:   "{color:#ff0000}red text{color}",
			want:    "red text",
			notWant: "{color",
		},
		{
			name:    "noformat",
			input:   "{noformat}plain text{noformat}",
			want:    "plain text",
			notWant: "{noformat}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJiraMarkup(tt.input)
			plain := stripAnsi(result)
			if !strings.Contains(plain, tt.want) {
				t.Errorf("expected %q in result, got %q", tt.want, plain)
			}
			if tt.notWant != "" && strings.Contains(plain, tt.notWant) {
				t.Errorf("did not expect %q in result, got %q", tt.notWant, plain)
			}
		})
	}
}
