package tui

import (
	"buble_jira/internal/model"
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

type PanelSection struct {
	Title   string
	Tickets []model.Ticket
}

type PanelModel struct {
	Title      string
	Number     int    // 1-6 hotkey number
	TitleColor string // ANSI color code for title (e.g. "36" for cyan)
	Tickets    []model.Ticket
	Sections   []PanelSection // if set, renders with section headers
	Cursor     int
	Focused    bool
	Width      int
	Height     int
	ScrollOffset int
}

func NewPanel(title string, number int, titleColor string) PanelModel {
	return PanelModel{Title: title, Number: number, TitleColor: titleColor}
}

func (p *PanelModel) SetSize(w, h int) {
	p.Width = w
	p.Height = h
}

func (p *PanelModel) SetTickets(tickets []model.Ticket) {
	p.Sections = nil
	p.Tickets = tickets
	if p.Cursor >= len(tickets) {
		if len(tickets) > 0 {
			p.Cursor = len(tickets) - 1
		} else {
			p.Cursor = 0
		}
	}
	p.clampScroll()
}

func (p *PanelModel) SetSections(sections []PanelSection) {
	p.Sections = sections
	var all []model.Ticket
	for _, s := range sections {
		all = append(all, s.Tickets...)
	}
	p.Tickets = all
	if p.Cursor >= len(all) {
		if len(all) > 0 {
			p.Cursor = len(all) - 1
		} else {
			p.Cursor = 0
		}
	}
	p.clampScroll()
}

func (p *PanelModel) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
		p.clampScroll()
	}
}

func (p *PanelModel) MoveDown() {
	if p.Cursor < len(p.Tickets)-1 {
		p.Cursor++
		p.clampScroll()
	}
}

func (p *PanelModel) SelectedTicket() *model.Ticket {
	if len(p.Tickets) == 0 {
		return nil
	}
	return &p.Tickets[p.Cursor]
}

// displayItems builds the list of renderable rows: section headers, tickets, empty markers.
// Returns items and the display index of the cursor.
type displayItem struct {
	isHeader bool
	header   string
	isEmpty  bool // "EMPTY" placeholder
	ticket   *model.Ticket
	ticketIdx int // index in p.Tickets (-1 for non-ticket items)
}

func (p *PanelModel) buildDisplayItems() ([]displayItem, int) {
	if len(p.Sections) == 0 {
		// No sections — flat ticket list
		items := make([]displayItem, len(p.Tickets))
		cursorDisplay := 0
		for i := range p.Tickets {
			items[i] = displayItem{ticket: &p.Tickets[i], ticketIdx: i}
			if i == p.Cursor {
				cursorDisplay = i
			}
		}
		if len(items) == 0 {
			items = []displayItem{{isEmpty: true, ticketIdx: -1}}
		}
		return items, cursorDisplay
	}

	var items []displayItem
	cursorDisplay := 0
	ticketIdx := 0

	for _, sec := range p.Sections {
		items = append(items, displayItem{isHeader: true, header: sec.Title, ticketIdx: -1})
		if len(sec.Tickets) == 0 {
			items = append(items, displayItem{isEmpty: true, ticketIdx: -1})
		} else {
			for i := range sec.Tickets {
				if ticketIdx == p.Cursor {
					cursorDisplay = len(items)
				}
				items = append(items, displayItem{ticket: &sec.Tickets[i], ticketIdx: ticketIdx})
				ticketIdx++
			}
		}
	}

	return items, cursorDisplay
}

func (p *PanelModel) clampScroll() {
	vis := p.visibleLines()
	if vis <= 0 {
		return
	}
	_, cursorDisplay := p.buildDisplayItems()
	if cursorDisplay < p.ScrollOffset {
		p.ScrollOffset = cursorDisplay
	}
	if cursorDisplay >= p.ScrollOffset+vis {
		p.ScrollOffset = cursorDisplay - vis + 1
	}
}

func (p *PanelModel) visibleLines() int {
	lines := p.Height - 2
	if lines < 1 {
		return 1
	}
	return lines
}

func (p *PanelModel) View() string {
	innerWidth := p.Width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	vis := p.visibleLines()

	items, _ := p.buildDisplayItems()

	var sb strings.Builder
	sb.Grow(vis * (innerWidth + 10))

	end := p.ScrollOffset + vis
	if end > len(items) {
		end = len(items)
	}

	maxSummaryW := innerWidth - 20
	if maxSummaryW < 5 {
		maxSummaryW = 5
	}

	linesWritten := 0
	for i := p.ScrollOffset; i < end; i++ {
		if linesWritten > 0 {
			sb.WriteByte('\n')
		}
		item := items[i]
		switch {
		case item.isHeader:
			sb.WriteString(fmt.Sprintf(" %s", item.header))
		case item.isEmpty:
			sb.WriteString("   \033[38;5;242mEMPTY\033[0m")
		default:
			t := item.ticket
			pIcon := priorityIcon(t.Priority)
			summary := truncate(t.Summary, maxSummaryW)
			if item.ticketIdx == p.Cursor && p.Focused {
				sb.WriteString(fmt.Sprintf(" \033[36m▸\033[0m %s \033[1;37m%s\033[0m \033[38;5;245m%s\033[0m", pIcon, t.IssueKey, summary))
			} else {
				sb.WriteString(fmt.Sprintf("   %s \033[38;5;252m%s\033[0m \033[38;5;242m%s\033[0m", pIcon, t.IssueKey, summary))
			}
		}
		linesWritten++
	}

	// Pad to fill height
	for linesWritten < vis {
		sb.WriteByte('\n')
		linesWritten++
	}

	// Build border
	borderColor := "38;5;240"
	if p.Focused {
		borderColor = "38;5;245"
	}

	titleText := fmt.Sprintf("%d:%s", p.Number, p.Title)
	titleRendered := fmt.Sprintf("\033[1;%sm%s\033[0m", p.TitleColor, titleText)

	var out strings.Builder
	out.Grow(sb.Len() + 300)

	// Top border
	out.WriteString(fmt.Sprintf("\033[%sm╭─\033[0m %s \033[%sm", borderColor, titleRendered, borderColor))
	topUsed := 1 + 1 + len(titleText) + 1
	for topUsed < innerWidth {
		out.WriteString("─")
		topUsed++
	}
	out.WriteString("╮\033[0m\n")

	// Content lines
	content := sb.String()
	for _, line := range strings.Split(content, "\n") {
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

func priorityIcon(priority string) string {
	switch strings.ToLower(priority) {
	case "show-stopper", "blocker":
		return "⛔"
	case "critical":
		return "\033[38;5;209m⏶\033[0m"
	case "major":
		return "\033[38;5;209m∧\033[0m"
	case "minor":
		return "\033[38;5;75m∨\033[0m"
	case "trivial":
		return "\033[38;5;242m∨\033[0m"
	default:
		return "\033[38;5;245m●\033[0m"
	}
}

func truncate(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(runes[:maxW-1]) + "…"
}

func truncateAnsi(s string, maxW int) string {
	var out strings.Builder
	w := 0
	inEsc := false
	for _, r := range s {
		if inEsc {
			out.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			out.WriteRune(r)
			continue
		}
		rw := runewidth.RuneWidth(r)
		if w+rw > maxW {
			break
		}
		out.WriteRune(r)
		w += rw
	}
	out.WriteString("\033[0m")
	return out.String()
}

func visibleWidth(s string) int {
	w := 0
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
		w += runewidth.RuneWidth(r)
	}
	return w
}
