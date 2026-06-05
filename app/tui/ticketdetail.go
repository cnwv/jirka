package tui

import (
	"github.com/cnwv/jirka/app/model"
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

type TicketDetailModel struct {
	Ticket       *model.Ticket
	Focused      bool
	Width        int
	Height       int
	ScrollOffset int
	KeyNumber    int    // key to focus detail view (panelCount + 1)
	lines        []string // precomputed on SetTicket
}

func (d *TicketDetailModel) SetSize(w, h int) {
	d.Width = w
	d.Height = h
}

func (d *TicketDetailModel) SetTicket(t *model.Ticket) {
	d.Ticket = t
	d.ScrollOffset = 0
	d.buildLines()
}

func (d *TicketDetailModel) ScrollUp() {
	d.ScrollOffset = max(d.ScrollOffset-10, 0)
}

func (d *TicketDetailModel) ScrollDown() {
	maxOffset := max(len(d.lines)-d.visibleHeight(), 0)
	d.ScrollOffset += 10
	if d.ScrollOffset > maxOffset {
		d.ScrollOffset = maxOffset
	}
}

func (d *TicketDetailModel) visibleHeight() int {
	h := d.Height - 2
	if h < 1 {
		return 1
	}
	return h
}

func (d *TicketDetailModel) buildLines() {
	if d.Ticket == nil {
		d.lines = []string{"\033[90mSelect a ticket to view details\033[0m"}
		return
	}

	t := d.Ticket
	var lines []string

	wrapW := max(d.Width-2, 10)

	// Title: KEY — Summary
	titlePlain := t.IssueKey + " \u2014 " + t.Summary
	titleWrapped := wrapLine(titlePlain, wrapW)
	keyLen := len(t.IssueKey)
	dashPrefix := " \u2014 "
	for i, wl := range titleWrapped {
		if i == 0 {
			if len(wl) > keyLen+len(dashPrefix) {
				lines = append(lines, fmt.Sprintf("\033[1;38;5;209m%s\033[0;38;5;242m%s\033[0m\033[1;37m%s\033[0m",
					t.IssueKey, dashPrefix, wl[keyLen+len(dashPrefix):]))
			} else {
				lines = append(lines, fmt.Sprintf("\033[1;38;5;209m%s\033[0m", wl))
			}
		} else {
			lines = append(lines, fmt.Sprintf("\033[1;37m%s\033[0m", wl))
		}
	}
	lines = append(lines, "")

	// Fields — aligned key-value pairs
	assignee := t.AssigneeName
	if !t.IsAssigned {
		assignee = "Unassigned"
	}

	fields := [][2]string{
		{"Status", t.StatusName},
		{"Priority", t.Priority},
		{"Type", t.IssueType},
		{"Assignee", assignee},
		{"Reporter", t.ReporterName},
		{"Created", t.Created.Format("2006-01-02 15:04")},
	}
	if t.Components != "" {
		fields = append(fields, [2]string{"Components", t.Components})
	}

	// Find max label width for alignment
	labelW := 0
	for _, f := range fields {
		if len(f[0]) > labelW {
			labelW = len(f[0])
		}
	}

	for _, f := range fields {
		padding := strings.Repeat(" ", labelW-len(f[0]))
		lines = append(lines, fmt.Sprintf("\033[38;5;110m%s%s\033[0m  \033[38;5;255m%s\033[0m", f[0], padding, f[1]))
	}

	// Separator line
	lines = append(lines, "", "\033[38;5;238m"+strings.Repeat("\u2500", wrapW)+"\033[0m", "")

	// Description
	if t.Description != "" {
		lines = append(lines, "\033[1;37mDescription\033[0m", "")
		desc := strings.ReplaceAll(t.Description, "\r", "")
		desc = formatJiraMarkup(desc)
		for dl := range strings.SplitSeq(desc, "\n") {
			wrapped := wrapLine(dl, wrapW)
			for _, wl := range wrapped {
				lines = append(lines, fmt.Sprintf("\033[38;5;252m%s\033[0m", wl))
			}
		}
	} else {
		lines = append(lines, "\033[38;5;242mNo description\033[0m")
	}

	d.lines = lines
}

func wrapLine(s string, maxW int) []string {
	if visibleWidth(s) <= maxW {
		return []string{s}
	}
	var result []string
	runes := []rune(s)
	for len(runes) > 0 {
		w := 0
		cut := 0
		lastSpace := -1
		for i, r := range runes {
			rw := 1
			if r > 127 {
				rw = runewidth.RuneWidth(r)
			}
			if w+rw > maxW {
				break
			}
			if r == ' ' {
				lastSpace = i
			}
			w += rw
			cut = i + 1
		}
		if cut < len(runes) && lastSpace > 0 {
			cut = lastSpace + 1
		}
		result = append(result, string(runes[:cut]))
		runes = runes[cut:]
	}
	return result
}

func (d *TicketDetailModel) View() string {
	if d.Width < 3 || d.Height < 3 {
		return ""
	}

	vis := max(d.Height-2, 1) // content lines = total height - top/bottom borders

	// Reserve 1 line for scroll indicator when focused and content overflows
	contentVis := vis
	if d.Focused && len(d.lines) > vis {
		contentVis = vis - 1
	}

	start := d.ScrollOffset
	end := min(start+contentVis, len(d.lines))
	if start > len(d.lines) {
		start = len(d.lines)
	}

	contentLines := make([]string, 0, vis)
	for i := start; i < end; i++ {
		contentLines = append(contentLines, d.lines[i])
	}

	if d.Focused && len(d.lines) > vis {
		contentLines = append(contentLines, fmt.Sprintf("\033[38;5;242m[%d/%d]\033[0m", d.ScrollOffset+1, len(d.lines)-vis+1))
	}

	const detailTitleColor = "38;5;147" // soft lavender
	keyNum := d.KeyNumber
	if keyNum <= 0 {
		keyNum = 7
	}
	title := fmt.Sprintf("%d:Ticket", keyNum)
	return borderedBox(title, detailTitleColor, contentLines, d.Width, d.Height, d.Focused)
}
