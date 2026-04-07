package tui

import (
	"buble_jira/internal/model"
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
	d.ScrollOffset -= 10
	if d.ScrollOffset < 0 {
		d.ScrollOffset = 0
	}
}

func (d *TicketDetailModel) ScrollDown() {
	maxOffset := len(d.lines) - d.visibleHeight()
	if maxOffset < 0 {
		maxOffset = 0
	}
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

	// Issue key + summary, wrap if needed
	wrapW := d.Width - 2
	if wrapW < 10 {
		wrapW = 10
	}
	titlePlain := t.IssueKey + " " + t.Summary
	titleWrapped := wrapLine(titlePlain, wrapW)
	for i, wl := range titleWrapped {
		if i == 0 {
			// First line: key in orange, rest in white
			if len(wl) > len(t.IssueKey) {
				lines = append(lines, fmt.Sprintf("\033[1;38;5;209m%s\033[0m\033[1;37m%s\033[0m", t.IssueKey, wl[len(t.IssueKey):]))
			} else {
				lines = append(lines, fmt.Sprintf("\033[1;38;5;209m%s\033[0m", wl))
			}
		} else {
			lines = append(lines, fmt.Sprintf("\033[1;37m%s\033[0m", wl))
		}
	}
	lines = append(lines, "")
	// Fields: labels in yellow, values in white
	lines = append(lines, fmt.Sprintf("\033[38;5;222mStatus:\033[0m %s", t.StatusName))
	lines = append(lines, fmt.Sprintf("\033[38;5;222mPriority:\033[0m %s", t.Priority))
	lines = append(lines, fmt.Sprintf("\033[38;5;222mType:\033[0m %s", t.IssueType))
	if t.Components != "" {
		lines = append(lines, fmt.Sprintf("\033[38;5;222mComponents:\033[0m %s", t.Components))
	}

	assignee := t.AssigneeName
	if !t.IsAssigned {
		assignee = "Unassigned"
	}
	lines = append(lines, fmt.Sprintf("\033[38;5;222mAssignee:\033[0m %s", assignee))
	lines = append(lines, fmt.Sprintf("\033[38;5;222mReporter:\033[0m %s", t.ReporterName))
	lines = append(lines, fmt.Sprintf("\033[38;5;222mCreated:\033[0m %s", t.Created.Format("2006-01-02 15:04")))
	lines = append(lines, "")

	if t.Description != "" {
		lines = append(lines, "\033[38;5;222mDescription:\033[0m")
		desc := strings.ReplaceAll(t.Description, "\r", "")
		desc = formatJiraMarkup(desc)
		descLines := strings.Split(desc, "\n")
		wrapW := d.Width - 2 // inner width (minus border chars)
		if wrapW < 10 {
			wrapW = 10
		}
		for _, dl := range descLines {
			wrapped := wrapLine(dl, wrapW)
			for _, wl := range wrapped {
				lines = append(lines, fmt.Sprintf("\033[38;5;252m%s\033[0m", wl))
			}
		}
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

	innerWidth := d.Width - 2
	// content lines = total height - top border - bottom border
	vis := d.Height - 2
	if vis < 1 {
		vis = 1
	}

	// Reserve 1 line for scroll indicator when focused and content overflows
	contentVis := vis
	if d.Focused && len(d.lines) > vis {
		contentVis = vis - 1
	}

	start := d.ScrollOffset
	end := start + contentVis
	if end > len(d.lines) {
		end = len(d.lines)
	}
	if start > len(d.lines) {
		start = len(d.lines)
	}

	// Build content lines
	var contentBuf strings.Builder
	linesWritten := 0
	for i := start; i < end; i++ {
		if linesWritten > 0 {
			contentBuf.WriteByte('\n')
		}
		line := d.lines[i]
		if visibleWidth(line) > innerWidth {
			line = truncateAnsi(line, innerWidth)
		}
		contentBuf.WriteString(line)
		linesWritten++
	}

	if d.Focused && len(d.lines) > vis {
		contentBuf.WriteByte('\n')
		contentBuf.WriteString(fmt.Sprintf("\033[38;5;242m[%d/%d]\033[0m", d.ScrollOffset+1, len(d.lines)-vis+1))
		linesWritten++
	}

	// Pad remaining lines
	for linesWritten < vis {
		contentBuf.WriteByte('\n')
		linesWritten++
	}

	// Build bordered output
	borderColor := "38;5;240"
	if d.Focused {
		borderColor = "38;5;245"
	}
	titleColor := "38;5;147" // soft lavender for detail panel
	keyNum := d.KeyNumber
	if keyNum <= 0 {
		keyNum = 7
	}
	titleText := fmt.Sprintf("%d:Ticket", keyNum)
	titleRendered := fmt.Sprintf("\033[1;%sm%s\033[0m", titleColor, titleText)

	var out strings.Builder

	// Top border
	out.WriteString(fmt.Sprintf("\033[%sm╭─\033[0m %s \033[%sm", borderColor, titleRendered, borderColor))
	topUsed := 1 + 1 + len(titleText) + 1
	for topUsed < innerWidth {
		out.WriteString("─")
		topUsed++
	}
	out.WriteString("╮\033[0m\n")

	// Content lines with side borders
	for _, line := range strings.Split(contentBuf.String(), "\n") {
		out.WriteString(fmt.Sprintf("\033[%sm│\033[0m", borderColor))
		lineW := visibleWidth(line)
		if lineW > innerWidth {
			out.WriteString(truncateAnsi(line, innerWidth))
		} else {
			out.WriteString(line)
			for i := lineW; i < innerWidth; i++ {
				out.WriteByte(' ')
			}
		}
		out.WriteString(fmt.Sprintf("\033[%sm│\033[0m\n", borderColor))
	}

	// Bottom border
	out.WriteString(fmt.Sprintf("\033[%sm╰", borderColor))
	for i := 0; i < innerWidth; i++ {
		out.WriteString("─")
	}
	out.WriteString("╯\033[0m")

	return out.String()
}
