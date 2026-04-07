package tui

import (
	"fmt"
	"strings"
	"time"
)

// StatusBarModel renders the bottom status bar.
type StatusBarModel struct {
	LastUpdated time.Time
	Error       string
	Hint        string
	Width       int
	PanelCount  int
}

// SetWidth sets the total available width.
func (s *StatusBarModel) SetWidth(w int) {
	s.Width = w
}

// View renders the status bar.
func (s *StatusBarModel) View() string {
	var left string
	switch {
	case s.Error != "":
		left = "\033[38;5;203m⚠ " + s.Error + "\033[0m"
	case s.Hint != "":
		left = "\033[38;5;222m" + s.Hint + "\033[0m"
	case !s.LastUpdated.IsZero():
		left = "\033[38;5;245mupdated " + s.LastUpdated.Format("15:04") + "\033[0m"
	default:
		left = "\033[38;5;245mloading...\033[0m"
	}

	detailKey := s.PanelCount + 1
	if detailKey <= 1 {
		detailKey = 7 // fallback
	}
	right := fmt.Sprintf("\033[38;5;242m1-%d panels  e: edit  n: new win  0: switch win  b: browser  r: refresh  ↑↓  q: quit\033[0m", detailKey)

	gap := max(s.Width-visibleWidth(left)-visibleWidth(right), 1)

	var sb strings.Builder
	sb.WriteString(left)
	sb.WriteString(strings.Repeat(" ", gap))
	sb.WriteString(right)
	return sb.String()
}
