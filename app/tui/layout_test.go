package tui

import (
	"github.com/cnwv/jirka/app/config"
	"github.com/cnwv/jirka/app/model"
	"fmt"
	"strings"
	"testing"
	"time"
)

func testDashboard() *config.DashboardConfig {
	return &config.DashboardConfig{
		Windows: []config.WindowConfig{
			{
				Name:   "Test",
				Layout: config.GridConfig{Rows: 3, Cols: 2},
				Panels: []config.PanelConfig{
					{Title: "Panel 1", Color: "38;5;75"},
					{Title: "Panel 2", Color: "38;5;222"},
					{Title: "Panel 3", Color: "38;5;114"},
					{Title: "Panel 4", Color: "38;5;209"},
					{Title: "Panel 5", Color: "38;5;177"},
					{Title: "Panel 6", Color: "38;5;80"},
				},
			},
		},
	}
}

func newTestRoot(width, height int) *RootModel {
	m := NewRootModel(nil, 5*time.Minute, "https://jira.example.com", testDashboard())

	m.panels[0].SetTickets([]model.Ticket{
		{IssueKey: "TEST-1", Summary: "First ticket", StatusName: "Open", Priority: "High", IssueType: "Bug"},
		{IssueKey: "TEST-2", Summary: "Second ticket", StatusName: "Open", Priority: "Medium", IssueType: "Task"},
	})

	m.width = width
	m.height = height
	m.recalcLayout()
	m.updateDetail()
	return m
}

func TestLayoutLineCount(t *testing.T) {
	sizes := [][2]int{
		{120, 40},
		{180, 50},
		{200, 60},
		{100, 30},
		{160, 45},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			m := newTestRoot(w, h)
			view := m.View()
			content := view.Content

			lines := strings.Split(content, "\n")
			// View should produce bodyH lines of grid + 1 status bar line.
			// bodyH = h - 2, so total lines in content = bodyH + 1 = h - 1.
			// strings.Split on trailing \n adds an extra empty element.
			// Each grid row ends with \n, status bar has no trailing \n.
			expectedLines := h - 1

			// Remove trailing empty string from split
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}

			if len(lines) != expectedLines {
				t.Errorf("expected %d lines, got %d", expectedLines, len(lines))
			}
		})
	}
}

func TestLayoutLineWidth(t *testing.T) {
	sizes := [][2]int{
		{120, 40},
		{180, 50},
		{200, 60},
		{100, 30},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			m := newTestRoot(w, h)
			view := m.View()
			content := view.Content

			lines := strings.Split(content, "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}

			bodyH := h - 2
			for i, line := range lines {
				vw := visibleWidth(line)
				if i < bodyH {
					if vw != w {
						t.Errorf("grid line %d: visible width = %d, want %d\n  line: %q", i, vw, w, stripAnsi(line))
					}
				}
			}
		})
	}
}

func TestPanelBorderLineCount(t *testing.T) {
	widths := []int{40, 60, 80}
	heights := []int{10, 15, 20}

	for _, pw := range widths {
		for _, ph := range heights {
			t.Run(fmt.Sprintf("panel_%dx%d", pw, ph), func(t *testing.T) {
				p := NewPanel("Test", 1, "38;5;75")
				p.SetSize(pw, ph)
				p.SetTickets([]model.Ticket{
					{IssueKey: "T-1", Summary: "Ticket one", Priority: "High"},
				})
				out := p.View()
				lines := strings.Split(out, "\n")
				if len(lines) != ph {
					t.Errorf("panel: expected %d lines, got %d", ph, len(lines))
				}
				for i, line := range lines {
					vw := visibleWidth(line)
					if vw != pw {
						t.Errorf("panel line %d: visible width = %d, want %d", i, vw, pw)
					}
				}
			})
		}
	}
}

func TestDetailBorderLineCount(t *testing.T) {
	widths := []int{40, 60, 90}
	heights := []int{10, 20, 38}

	for _, dw := range widths {
		for _, dh := range heights {
			t.Run(fmt.Sprintf("detail_%dx%d", dw, dh), func(t *testing.T) {
				d := TicketDetailModel{}
				d.SetSize(dw, dh)
				d.SetTicket(&model.Ticket{
					IssueKey: "T-1", Summary: "Test ticket", StatusName: "Open",
					Priority: "High", IssueType: "Bug",
					AssigneeName: "User", ReporterName: "Reporter",
					Created: time.Now(), Updated: time.Now(),
				})
				out := d.View()
				lines := strings.Split(out, "\n")
				if len(lines) != dh {
					t.Errorf("detail: expected %d lines, got %d", dh, len(lines))
				}
				for i, line := range lines {
					vw := visibleWidth(line)
					if vw != dw {
						t.Errorf("detail line %d: visible width = %d, want %d", i, vw, dw)
					}
				}
			})
		}
	}
}

func TestLayoutWithRealContent(t *testing.T) {
	realTicket := model.Ticket{
		IssueKey:     "CLPS-26587",
		Summary:      "KingBilly: referral bonus issue",
		StatusName:   "Open",
		Priority:     "Major",
		IssueType:    "Problem",
		Components:   "Referral system",
		AssigneeName: "",
		IsAssigned:   false,
		ReporterName: "Ekaterina Ignatovets",
		Created:      time.Date(2026, 4, 3, 18, 20, 0, 0, time.UTC),
		Updated:      time.Date(2026, 4, 3, 18, 20, 0, 0, time.UTC),
		Description: `{*}Description{*}:

We have an issue with referral bonus issuing.

The inviter didn't receive a bonus, though both players fulfilled requirements and the bonus was supposed to
Inviter: [https://kingbilly.casino-backend.com/en-AU/backend/players/1684328]
Invited player: [https://kingbilly.casino-backend.com/en-AU/backend/players/1686484]
Kindly assist us with finding the reason of this issue and how can we fix it.

{*}Actual Result{*}:

referral bonus wasn't received

{*}Expected Result{*}:

referral bonus is received`,
	}

	sizes := [][2]int{
		{120, 40},
		{180, 50},
		{200, 60},
		{100, 30},
		{160, 45},
		{80, 24},
	}

	for _, sz := range sizes {
		w, h := sz[0], sz[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			m := newTestRoot(w, h)
			m.SetTestTicketForDetail(&realTicket)
			view := m.View()
			content := view.Content

			lines := strings.Split(content, "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}

			bodyH := h - 2
			for i, line := range lines {
				vw := visibleWidth(line)
				if i < bodyH && vw != w {
					t.Errorf("line %d: visible width = %d, want %d\n  stripped: %q", i, vw, w, stripAnsi(line))
				}
			}
		})
	}
}

func stripAnsi(s string) string {
	var out strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}
