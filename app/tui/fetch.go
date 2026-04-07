package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/cnwv/jirka/app/config"
	"github.com/cnwv/jirka/app/jira"
	"github.com/cnwv/jirka/app/model"

	tea "charm.land/bubbletea/v2"
)

const fetchDelay = 150 * time.Millisecond

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
			sections := make([]PanelSection, 0, len(pc.Sections))
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
			return ErrorMsg{Err: errNoClient}
		}

		ctx := context.Background()
		queries := buildQueries(win)
		fetched := make(map[string][]model.Ticket)

		for i, q := range queries {
			if i > 0 {
				time.Sleep(fetchDelay)
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

func scheduleTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
