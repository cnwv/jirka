package tui

import (
	"buble_jira/internal/config"
	"buble_jira/internal/jira"
	"buble_jira/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type RootModel struct {
	panels      []PanelModel
	panelCount  int // = gridRows * gridCols
	focusDetail int // = panelCount (sentinel for detail view)
	gridRows    int
	gridCols    int
	colWidths   []int
	rowHeights  []int

	focusedPanel int
	detail       TicketDetailModel
	statusBar    StatusBarModel
	width        int
	height       int

	rightW int
	bodyH  int

	jiraClient *jira.Client
	jiraURL    string

	// Window management
	windows    []config.WindowConfig
	activeWin  int
	configPath string // empty = read-only (example config)
	isExample  bool

	sub          chan fetchResult
	pollInterval time.Duration

	// Active popup (nil when none)
	popup any // *winSwitcherPopup | *newWindowPopup | *panelEditPopup
}

func NewRootModel(client *jira.Client, pollInterval time.Duration, jiraURL string, dashboard *config.DashboardConfig) *RootModel {
	m := &RootModel{
		jiraClient:   client,
		jiraURL:      jiraURL,
		windows:      dashboard.Windows,
		configPath:   dashboard.ConfigPath,
		isExample:    dashboard.IsExample,
		sub:          make(chan fetchResult, 1),
		pollInterval: pollInterval,
	}

	m.loadWindow(0)

	if dashboard.IsExample {
		m.statusBar.Hint = "Demo mode: copy config.example.yaml to config.yaml to connect your Jira"
	}

	return m
}

// loadWindow initializes panels for the window at idx and sets it as active.
func (m *RootModel) loadWindow(idx int) {
	if idx < 0 || idx >= len(m.windows) {
		return
	}
	w := m.windows[idx]
	m.activeWin = idx

	rows, cols := w.EffectiveGrid()
	count := rows * cols
	m.panelCount = count
	m.focusDetail = count
	m.gridRows = rows
	m.gridCols = cols
	m.panels = make([]PanelModel, count)
	m.colWidths = make([]int, cols)
	m.rowHeights = make([]int, rows)
	m.focusedPanel = 0
	m.detail = TicketDetailModel{}

	for i, pc := range w.Panels {
		m.panels[i] = NewPanel(pc.Title, i+1, pc.Color)
	}
	m.panels[0].Focused = true

	m.loadStubs()

	if m.width > 0 {
		m.recalcLayout()
	}
}

func stubsToTickets(stubs []config.StubTicket) []model.Ticket {
	tickets := make([]model.Ticket, len(stubs))
	for i, s := range stubs {
		tickets[i] = model.Ticket{
			IssueKey:   s.Key,
			Summary:    s.Summary,
			StatusName: s.Status,
			Priority:   s.Priority,
			IssueType:  s.Type,
		}
	}
	return tickets
}

func (m *RootModel) loadStubs() {
	if m.activeWin >= len(m.windows) {
		return
	}
	for i, pc := range m.windows[m.activeWin].Panels {
		if i >= m.panelCount {
			break
		}
		if len(pc.Sections) > 0 {
			var sections []PanelSection
			hasStubs := false
			for _, sec := range pc.Sections {
				ps := PanelSection{
					Title:   formatSectionTitle(sec.Name, sec.Color),
					Tickets: stubsToTickets(sec.Stubs),
				}
				if len(sec.Stubs) > 0 {
					hasStubs = true
				}
				sections = append(sections, ps)
			}
			if hasStubs {
				m.panels[i].SetSections(sections)
			}
		} else if len(pc.Stubs) > 0 {
			m.panels[i].SetTickets(stubsToTickets(pc.Stubs))
		}
	}
}

func (m *RootModel) Init() tea.Cmd {
	if m.isExample {
		return nil
	}
	return tea.Batch(
		doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
		waitForTickets(m.sub),
		scheduleTick(m.pollInterval),
	)
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()
		return m, nil

	case tea.PasteMsg:
		m.dispatchPaste(msg.Content)
		return m, nil

	case clipboardReadMsg:
		m.dispatchPaste(string(msg))
		return m, nil

	case jqlTestResultMsg:
		if p, ok := m.popup.(*panelEditPopup); ok {
			p.setTestResult(msg.ok, msg.summary)
			if msg.ok {
				p.focusField = 2 // advance to color on success
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		key := msg.String()

		// Popup intercepts keys first
		if m.popup != nil {
			return m.handlePopupKey(key)
		}

		switch key {
		case "ctrl+c", "q", "й":
			return m, tea.Quit
		case "r", "к":
			m.statusBar.Error = ""
			return m, doFetch(m.jiraClient, m.sub, m.windows[m.activeWin])
		case "b", "и":
			m.openInBrowser()
			return m, nil
		case "e":
			if m.focusedPanel < m.panelCount {
				pc := m.windows[m.activeWin].Panels[m.focusedPanel]
				m.popup = newPanelEditPopup(m.focusedPanel, pc)
			}
			return m, nil
		case "n":
			m.popup = newNewWindowPopup()
			return m, nil
		case "0":
			names := make([]string, len(m.windows))
			for i, w := range m.windows {
				names[i] = w.Name
			}
			m.popup = newWinSwitcherPopup(names, m.activeWin)
			return m, nil
		case "tab":
			m.nextPanel()
			m.updateDetail()
			return m, nil
		case "shift+tab":
			m.prevPanel()
			m.updateDetail()
			return m, nil
		case "up", "k", "л":
			if m.focusedPanel == m.focusDetail {
				m.detail.ScrollUp()
			} else {
				m.panels[m.focusedPanel].MoveUp()
				m.updateDetail()
			}
			return m, nil
		case "down", "j", "о":
			if m.focusedPanel == m.focusDetail {
				m.detail.ScrollDown()
			} else {
				m.panels[m.focusedPanel].MoveDown()
				m.updateDetail()
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
				n := int(key[0] - '0')
				if n >= 1 && n <= m.panelCount {
					m.focusPanel(n - 1)
					m.updateDetail()
					return m, nil
				}
				if n == m.panelCount+1 {
					m.focusDetailView()
					return m, nil
				}
			}
		}

	case TicketsRefreshedMsg:
		for panelIdx, tickets := range msg.ByPanel {
			if panelIdx >= 0 && panelIdx < m.panelCount {
				m.panels[panelIdx].SetTickets(tickets)
			}
		}
		for panelIdx, sections := range msg.BySections {
			if panelIdx >= 0 && panelIdx < m.panelCount {
				m.panels[panelIdx].SetSections(sections)
			}
		}
		m.statusBar.LastUpdated = msg.At
		m.statusBar.Error = ""
		m.updateDetail()
		return m, waitForTickets(m.sub)

	case TickMsg:
		return m, doFetch(m.jiraClient, m.sub, m.windows[m.activeWin])

	case ErrorMsg:
		m.statusBar.Error = msg.Err.Error()
		return m, nil
	}

	return m, nil
}

func (m *RootModel) handlePopupKey(key string) (tea.Model, tea.Cmd) {
	if key == "ctrl+v" {
		return m, readSystemClipboard()
	}

	switch p := m.popup.(type) {

	case *winSwitcherPopup:
		action, idx := p.handleKey(key)
		switch action {
		case "switch":
			m.popup = nil
			if idx != m.activeWin {
				m.switchWindow(idx)
				if !m.isExample {
					return m, tea.Batch(
						doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
						waitForTickets(m.sub),
					)
				}
			}
		case "new":
			m.popup = newNewWindowPopup()
		case "delete":
			if len(m.windows) > 1 {
				m.deleteWindow(idx)
				// Update switcher
				names := make([]string, len(m.windows))
				for i, w := range m.windows {
					names[i] = w.Name
				}
				sel := idx
				if sel >= len(m.windows) {
					sel = len(m.windows) - 1
				}
				m.popup = newWinSwitcherPopup(names, sel)
			}
		case "close":
			m.popup = nil
		}

	case *newWindowPopup:
		action, win := p.handleKey(key)
		switch action {
		case "done":
			m.popup = nil
			m.addWindow(win)
			if !m.isExample {
				return m, tea.Batch(
					doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
					waitForTickets(m.sub),
				)
			}
		case "close":
			m.popup = nil
		}

	case *panelEditPopup:
		action, pc := p.handleKey(key)
		switch action {
		case "test_jql":
			return m, testJQL(m.jiraClient, p.currentJQL())
		case "save":
			m.popup = nil
			m.updatePanel(p.panelIdx, pc)
			if !m.isExample && pc.JQL != "" {
				return m, tea.Batch(
					doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
					waitForTickets(m.sub),
				)
			}
		case "close":
			m.popup = nil
		}
	}

	return m, nil
}

func (m *RootModel) switchWindow(idx int) {
	if idx < 0 || idx >= len(m.windows) {
		return
	}
	m.loadWindow(idx)
	m.updateDetail()
}

func (m *RootModel) addWindow(win config.WindowConfig) {
	m.windows = append(m.windows, win)
	m.switchWindow(len(m.windows) - 1)
	m.saveConfig()
}

func (m *RootModel) deleteWindow(idx int) {
	if idx < 0 || idx >= len(m.windows) || len(m.windows) <= 1 {
		return
	}
	m.windows = append(m.windows[:idx], m.windows[idx+1:]...)
	active := m.activeWin
	if active >= len(m.windows) {
		active = len(m.windows) - 1
	}
	m.switchWindow(active)
	m.saveConfig()
}

func (m *RootModel) updatePanel(panelIdx int, pc config.PanelConfig) {
	if m.activeWin >= len(m.windows) || panelIdx >= len(m.windows[m.activeWin].Panels) {
		return
	}
	m.windows[m.activeWin].Panels[panelIdx] = pc
	m.panels[panelIdx] = NewPanel(pc.Title, panelIdx+1, pc.Color)
	if m.width > 0 {
		m.panels[panelIdx].SetSize(m.colWidths[panelIdx%m.gridCols], m.rowHeights[panelIdx/m.gridCols])
	}
	m.panels[panelIdx].Focused = (m.focusedPanel == panelIdx)
	m.updateDetail()
	m.saveConfig()
}

func (m *RootModel) saveConfig() {
	if m.configPath == "" {
		return
	}
	_ = config.SaveConfig(m.configPath, m.windows)
}

func (m *RootModel) recalcLayout() {
	m.rightW = m.width / 2
	leftTotal := m.width - m.rightW

	baseColW := leftTotal / m.gridCols
	for c := range m.colWidths {
		m.colWidths[c] = baseColW
	}
	m.colWidths[m.gridCols-1] += leftTotal - baseColW*m.gridCols

	m.bodyH = m.height - 2
	baseRowH := m.bodyH / m.gridRows
	for r := range m.rowHeights {
		m.rowHeights[r] = baseRowH
	}
	m.rowHeights[m.gridRows-1] += m.bodyH - baseRowH*m.gridRows

	for r := 0; r < m.gridRows; r++ {
		for c := 0; c < m.gridCols; c++ {
			m.panels[r*m.gridCols+c].SetSize(m.colWidths[c], m.rowHeights[r])
		}
	}

	m.detail.KeyNumber = m.panelCount + 1
	m.detail.SetSize(m.rightW, m.bodyH)
	m.statusBar.SetWidth(m.width)
	m.statusBar.PanelCount = m.panelCount
}

func (m *RootModel) View() tea.View {
	if m.width == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	gridW := m.width - m.rightW

	pLines := make([][]string, m.panelCount)
	for i := 0; i < m.panelCount; i++ {
		pLines[i] = strings.Split(m.panels[i].View(), "\n")
	}

	detailLines := strings.Split(m.detail.View(), "\n")

	rowStart := make([]int, m.gridRows+1)
	for r := 0; r < m.gridRows; r++ {
		rowStart[r+1] = rowStart[r] + m.rowHeights[r]
	}

	var sb strings.Builder

	for row := 0; row < m.bodyH; row++ {
		gridRow := m.gridRows - 1
		for r := 0; r < m.gridRows-1; r++ {
			if row < rowStart[r+1] {
				gridRow = r
				break
			}
		}
		localRow := row - rowStart[gridRow]

		var leftLine string
		for c := 0; c < m.gridCols; c++ {
			line := getLine(pLines[gridRow*m.gridCols+c], localRow)
			leftLine += padToWidth(line, m.colWidths[c])
		}

		leftVis := visibleWidth(leftLine)
		if leftVis < gridW {
			leftLine += strings.Repeat(" ", gridW-leftVis)
		}

		rightLine := getLine(detailLines, row)
		rightVis := visibleWidth(rightLine)
		if rightVis > m.rightW {
			rightLine = truncateAnsi(rightLine, m.rightW)
		}

		sb.WriteString(leftLine)
		sb.WriteString(rightLine)
		sb.WriteByte('\n')
	}

	sb.WriteString(m.statusBar.View())

	content := sb.String()

	// Render popup overlay
	if m.popup != nil {
		content = m.renderPopup(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m *RootModel) renderPopup(content string) string {
	switch p := m.popup.(type) {
	case *winSwitcherPopup:
		return p.view(m.width, m.height)
	case *newWindowPopup:
		return p.view(m.width, m.height)
	case *panelEditPopup:
		return p.view(m.width, m.height)
	}
	_ = content
	return content
}

func getLine(lines []string, idx int) string {
	if idx >= 0 && idx < len(lines) {
		return lines[idx]
	}
	return ""
}

func padToWidth(s string, w int) string {
	vis := visibleWidth(s)
	if vis >= w {
		if vis > w {
			return truncateAnsi(s, w)
		}
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

func (m *RootModel) focusPanel(idx int) {
	if idx < 0 || idx >= m.panelCount {
		return
	}
	if m.focusedPanel < m.panelCount {
		m.panels[m.focusedPanel].Focused = false
	}
	m.detail.Focused = false
	m.focusedPanel = idx
	m.panels[m.focusedPanel].Focused = true
}

func (m *RootModel) focusDetailView() {
	if m.focusedPanel < m.panelCount {
		m.panels[m.focusedPanel].Focused = false
	}
	m.focusedPanel = m.focusDetail
	m.detail.Focused = true
}

func (m *RootModel) nextPanel() {
	next := m.focusedPanel + 1
	if next > m.focusDetail {
		next = 0
	}
	if next == m.focusDetail {
		m.focusDetailView()
	} else {
		m.focusPanel(next)
	}
}

func (m *RootModel) prevPanel() {
	prev := m.focusedPanel - 1
	if prev < 0 {
		prev = m.focusDetail
	}
	if prev == m.focusDetail {
		m.focusDetailView()
	} else {
		m.focusPanel(prev)
	}
}

func (m *RootModel) updateDetail() {
	if m.focusedPanel < m.panelCount {
		t := m.panels[m.focusedPanel].SelectedTicket()
		m.detail.SetTicket(t)
	}
}

func (m *RootModel) currentTicket() *model.Ticket {
	if m.focusedPanel == m.focusDetail {
		return m.detail.Ticket
	}
	return m.panels[m.focusedPanel].SelectedTicket()
}

func (m *RootModel) openInBrowser() {
	t := m.currentTicket()
	if t == nil {
		return
	}
	url := fmt.Sprintf("%s/browse/%s", m.jiraURL, t.IssueKey)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("open", url)
	}
	_ = cmd.Start()
}

type fetchResult struct {
	ByPanel    map[int][]model.Ticket
	BySections map[int][]PanelSection
}

type sourceQuery struct {
	panelIdx    int
	sectionName string
	source      string
	jql         string
}

func formatSectionTitle(name, ansiColor string) string {
	return fmt.Sprintf("\033[1;%sm%s\033[0m", ansiColor, name)
}

func buildQueries(win config.WindowConfig) []sourceQuery {
	var queries []sourceQuery
	for i, pc := range win.Panels {
		if len(pc.Sections) > 0 {
			for _, sec := range pc.Sections {
				if sec.JQL == "" {
					continue
				}
				queries = append(queries, sourceQuery{
					panelIdx:    i,
					sectionName: sec.Name,
					source:      fmt.Sprintf("panel%d_%s", i, sec.Name),
					jql:         sec.JQL,
				})
			}
		} else if pc.JQL != "" {
			queries = append(queries, sourceQuery{
				panelIdx: i,
				source:   fmt.Sprintf("panel%d", i),
				jql:      pc.JQL,
			})
		}
	}
	return queries
}

func assembleResults(win config.WindowConfig, fetched map[string][]model.Ticket) fetchResult {
	result := fetchResult{
		ByPanel:    make(map[int][]model.Ticket),
		BySections: make(map[int][]PanelSection),
	}

	for i, pc := range win.Panels {
		if len(pc.Sections) > 0 {
			var sections []PanelSection
			for _, sec := range pc.Sections {
				source := fmt.Sprintf("panel%d_%s", i, sec.Name)
				sections = append(sections, PanelSection{
					Title:   formatSectionTitle(sec.Name, sec.Color),
					Tickets: fetched[source],
				})
			}
			result.BySections[i] = sections
		} else if pc.JQL != "" {
			source := fmt.Sprintf("panel%d", i)
			result.ByPanel[i] = fetched[source]
		}
	}

	return result
}

func fetchSource(ctx context.Context, client *jira.Client, q sourceQuery) ([]model.Ticket, error) {
	issues, err := client.SearchJQL(ctx, q.jql)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", q.source, err)
	}
	return jira.ToTickets(issues), nil
}

func doFetch(client *jira.Client, sub chan fetchResult, win config.WindowConfig) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return ErrorMsg{Err: fmt.Errorf("no Jira client configured")}
		}

		ctx := context.Background()
		queries := buildQueries(win)

		const delay = 150 * time.Millisecond
		fetched := make(map[string][]model.Ticket)

		for i, q := range queries {
			if i > 0 {
				time.Sleep(delay)
			}
			tickets, err := fetchSource(ctx, client, q)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			fetched[q.source] = tickets
		}

		sub <- assembleResults(win, fetched)
		return nil
	}
}

func waitForTickets(sub chan fetchResult) tea.Cmd {
	return func() tea.Msg {
		r := <-sub
		return TicketsRefreshedMsg{ByPanel: r.ByPanel, BySections: r.BySections, At: time.Now()}
	}
}

type clipboardReadMsg string

func readSystemClipboard() tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbpaste")
		case "linux":
			cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		default:
			return nil
		}
		out, err := cmd.Output()
		if err != nil || len(out) == 0 {
			return nil
		}
		return clipboardReadMsg(out)
	}
}

func (m *RootModel) dispatchPaste(content string) {
	if m.popup == nil || content == "" {
		return
	}
	switch p := m.popup.(type) {
	case *panelEditPopup:
		p.handlePaste(content)
	case *newWindowPopup:
		p.handlePaste(content)
	}
}

// compactErr extracts the useful part from Jira errors.
func compactErr(err error) string {
	msg := err.Error()

	// Network error: "jira request: Post "https://host/path": <reason>"
	if strings.HasPrefix(msg, "jira request: ") {
		if i := strings.Index(msg, `": `); i != -1 {
			msg = msg[i+3:]
		} else {
			msg = strings.TrimPrefix(msg, "jira request: ")
		}
		if len(msg) > 0 {
			msg = strings.ToUpper(msg[:1]) + msg[1:]
		}
		return msg
	}

	// HTTP error: "jira returned 400: <json body>"
	if strings.HasPrefix(msg, "jira returned ") {
		rest := strings.TrimPrefix(msg, "jira returned ")
		if colonIdx := strings.Index(rest, ": "); colonIdx != -1 {
			status := rest[:colonIdx]
			body := rest[colonIdx+2:]
			var errResp struct {
				ErrorMessages []string `json:"errorMessages"`
			}
			if json.Unmarshal([]byte(body), &errResp) == nil && len(errResp.ErrorMessages) > 0 {
				return errResp.ErrorMessages[0]
			}
			return "HTTP " + status
		}
	}

	return msg
}

func testJQL(client *jira.Client, jql string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return jqlTestResultMsg{ok: false, summary: "no Jira client (demo mode)"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		total, err := client.CountJQL(ctx, jql)
		if err != nil {
			return jqlTestResultMsg{ok: false, summary: compactErr(err)}
		}
		noun := "tickets"
		if total == 1 {
			noun = "ticket"
		}
		return jqlTestResultMsg{ok: true, summary: fmt.Sprintf("%d %s", total, noun)}
	}
}

func scheduleTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
