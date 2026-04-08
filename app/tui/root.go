package tui

import (
	"strings"
	"time"

	"github.com/cnwv/jirka/app/config"
	"github.com/cnwv/jirka/app/jira"
	"github.com/cnwv/jirka/app/model"

	tea "charm.land/bubbletea/v2"
)

// RootModel is the top-level BubbleTea model that owns the dashboard layout.
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

// NewRootModel creates the root model from config.
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
			m.loadSectionStubs(i, pc.Sections)
		} else if len(pc.Stubs) > 0 {
			m.panels[i].SetTickets(stubsToTickets(pc.Stubs))
		}
	}
}

func (m *RootModel) loadSectionStubs(panelIdx int, sections []config.SectionConfig) {
	panelSections := make([]PanelSection, 0, len(sections))
	hasStubs := false
	for _, sec := range sections {
		ps := PanelSection{
			Title:   formatSectionTitle(sec.Name, sec.Color),
			Tickets: stubsToTickets(sec.Stubs),
		}
		if len(sec.Stubs) > 0 {
			hasStubs = true
		}
		panelSections = append(panelSections, ps)
	}
	if hasStubs {
		m.panels[panelIdx].SetSections(panelSections)
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
		switch p := m.popup.(type) {
		case *panelEditPopup:
			p.setTestResult(msg.ok, msg.summary)
			if msg.ok {
				p.focusField = 2 // advance to color on success
			}
		case *sectionEditPopup:
			p.setTestResult(msg.ok, msg.summary)
			if msg.ok {
				p.focusField = 2
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.popup != nil {
			return m.handlePopupKey(msg.String())
		}
		return m.handleKey(msg.String())

	case TicketsRefreshedMsg:
		m.applyTickets(msg)
		return m, waitForTickets(m.sub)

	case TickMsg:
		return m, doFetch(m.jiraClient, m.sub, m.windows[m.activeWin])

	case ErrorMsg:
		m.statusBar.Error = msg.Err.Error()
		return m, nil
	}

	return m, nil
}

func (m *RootModel) handleKey(key string) (tea.Model, tea.Cmd) {
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
		m.popup = m.newWindowSwitcherPopup()
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
		m.scrollOrMove(-1)
		return m, nil
	case "down", "j", "о":
		m.scrollOrMove(1)
		return m, nil
	default:
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			n := int(key[0] - '0')
			if n >= 1 && n <= m.panelCount {
				m.focusPanel(n - 1)
				m.updateDetail()
			} else if n == m.panelCount+1 {
				m.focusDetailView()
			}
		}
		return m, nil
	}
}

func (m *RootModel) scrollOrMove(dir int) {
	if m.focusedPanel == m.focusDetail {
		m.scrollDetail(dir)
		return
	}
	m.moveInPanel(dir)
	m.updateDetail()
}

func (m *RootModel) scrollDetail(dir int) {
	if dir < 0 {
		m.detail.ScrollUp()
	} else {
		m.detail.ScrollDown()
	}
}

func (m *RootModel) moveInPanel(dir int) {
	if dir < 0 {
		m.panels[m.focusedPanel].MoveUp()
	} else {
		m.panels[m.focusedPanel].MoveDown()
	}
}

func (m *RootModel) newWindowSwitcherPopup() *winSwitcherPopup {
	names := make([]string, len(m.windows))
	for i, w := range m.windows {
		names[i] = w.Name
	}
	return newWinSwitcherPopup(names, m.activeWin)
}

func (m *RootModel) applyTickets(msg TicketsRefreshedMsg) {
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
}

func (m *RootModel) handlePopupKey(key string) (tea.Model, tea.Cmd) {
	if key == "ctrl+v" {
		return m, readSystemClipboard()
	}

	switch p := m.popup.(type) {
	case *winSwitcherPopup:
		return m.handleWinSwitcherKey(p, key)
	case *newWindowPopup:
		return m.handleNewWindowKey(p, key)
	case *panelEditPopup:
		return m.handlePanelEditKey(p, key)
	case *sectionEditPopup:
		return m.handleSectionEditKey(p, key)
	}

	return m, nil
}

func (m *RootModel) handleWinSwitcherKey(p *winSwitcherPopup, key string) (tea.Model, tea.Cmd) {
	action, idx := p.handleKey(key)
	switch action {
	case actionSwitch:
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
	case actionNew:
		m.popup = newNewWindowPopup()
	case actionDelete:
		if len(m.windows) > 1 {
			m.deleteWindow(idx)
			m.popup = m.newWindowSwitcherPopup()
		}
	case actionClose:
		m.popup = nil
	}
	return m, nil
}

func (m *RootModel) handleNewWindowKey(p *newWindowPopup, key string) (tea.Model, tea.Cmd) {
	action, win := p.handleKey(key)
	switch action {
	case actionDone:
		m.popup = nil
		m.addWindow(win)
		if !m.isExample {
			return m, tea.Batch(
				doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
				waitForTickets(m.sub),
			)
		}
	case actionClose:
		m.popup = nil
	}
	return m, nil
}

func (m *RootModel) handlePanelEditKey(p *panelEditPopup, key string) (tea.Model, tea.Cmd) {
	action, pc := p.handleKey(key)
	switch action {
	case actionTestJQL:
		return m, testJQL(m.jiraClient, p.currentJQL())
	case actionSave:
		m.popup = nil
		m.updatePanel(p.panelIdx, pc)
		if !m.isExample && (pc.JQL != "" || len(pc.Sections) > 0) {
			return m, tea.Batch(
				doFetch(m.jiraClient, m.sub, m.windows[m.activeWin]),
				waitForTickets(m.sub),
			)
		}
	case actionEditSection:
		sec := p.sections[p.sectionCursor]
		m.popup = newSectionEditPopup(p.panelIdx, p.sectionCursor, sec)
	case actionAddSection:
		m.popup = newSectionEditPopup(p.panelIdx, -1, config.SectionConfig{})
	case actionClose:
		m.popup = nil
	}
	return m, nil
}

func (m *RootModel) handleSectionEditKey(p *sectionEditPopup, key string) (tea.Model, tea.Cmd) {
	action, sec := p.handleKey(key)
	switch action {
	case actionTestJQL:
		return m, testJQL(m.jiraClient, p.currentJQL())
	case actionSave:
		// Update section in the panel config, then reopen panel edit
		pc := &m.windows[m.activeWin].Panels[p.panelIdx]
		if p.sectionIdx >= 0 {
			pc.Sections[p.sectionIdx] = sec
		} else {
			pc.Sections = append(pc.Sections, sec)
		}
		m.popup = newPanelEditPopup(p.panelIdx, *pc)
	case actionClose:
		pc := m.windows[m.activeWin].Panels[p.panelIdx]
		m.popup = newPanelEditPopup(p.panelIdx, pc)
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

	for r := range m.gridRows {
		for c := range m.gridCols {
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
	for i := range m.panelCount {
		pLines[i] = strings.Split(m.panels[i].View(), "\n")
	}

	detailLines := strings.Split(m.detail.View(), "\n")

	rowStart := make([]int, m.gridRows+1)
	for r := range m.gridRows {
		rowStart[r+1] = rowStart[r] + m.rowHeights[r]
	}

	var sb strings.Builder

	for row := range m.bodyH {
		gridRow := m.gridRows - 1
		for r := range m.gridRows - 1 {
			if row < rowStart[r+1] {
				gridRow = r
				break
			}
		}
		localRow := row - rowStart[gridRow]

		var leftParts strings.Builder
		for c := range m.gridCols {
			line := getLine(pLines[gridRow*m.gridCols+c], localRow)
			leftParts.WriteString(padToWidth(line, m.colWidths[c]))
		}
		leftLine := leftParts.String()

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
	case *sectionEditPopup:
		return p.view(m.width, m.height)
	}
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
	if vis > w {
		return truncateAnsi(s, w)
	}
	if vis == w {
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

// SetTestTicketForDetail is a test helper.
func (m *RootModel) SetTestTicketForDetail(t *model.Ticket) {
	m.detail.SetTicket(t)
}
