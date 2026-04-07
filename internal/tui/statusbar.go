package tui

import (
	"fmt"
	"strings"
	"time"
)

type StatusBarModel struct {
	LastUpdated time.Time
	Error       string
	Hint        string
	Width       int
	PanelCount  int
}

func (s *StatusBarModel) SetWidth(w int) {
	s.Width = w
}

func (s *StatusBarModel) View() string {
	var left string
	if s.Error != "" {
		left = fmt.Sprintf("\033[38;5;203m⚠ %s\033[0m", s.Error)
	} else if s.Hint != "" {
		left = fmt.Sprintf("\033[38;5;222m%s\033[0m", s.Hint)
	} else if !s.LastUpdated.IsZero() {
		left = fmt.Sprintf("\033[38;5;245mupdated %s\033[0m", s.LastUpdated.Format("15:04"))
	} else {
		left = "\033[38;5;245mloading...\033[0m"
	}

	detailKey := s.PanelCount + 1
	if detailKey <= 1 {
		detailKey = 7 // fallback
	}
	right := fmt.Sprintf("\033[38;5;242m1-%d panels  e: edit  n: new win  0: switch win  b: browser  r: refresh  ↑↓  q: quit\033[0m", detailKey)

	leftW := visibleWidth(left)
	rightW := visibleWidth(right)
	gap := s.Width - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	var sb strings.Builder
	sb.WriteString(left)
	sb.WriteString(strings.Repeat(" ", gap))
	sb.WriteString(right)
	return sb.String()
}
