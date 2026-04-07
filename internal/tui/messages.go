package tui

import (
	"buble_jira/internal/model"
	"time"
)

type TicketsRefreshedMsg struct {
	ByPanel    map[int][]model.Ticket
	BySections map[int][]PanelSection
	At         time.Time
}

type TickMsg time.Time

type ErrorMsg struct {
	Err error
}

type jqlTestResultMsg struct {
	ok      bool
	summary string // "✓ 12 tickets" or "✗ invalid JQL: ..."
}
