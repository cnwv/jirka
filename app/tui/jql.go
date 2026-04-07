package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cnwv/jirka/app/jira"

	tea "charm.land/bubbletea/v2"
)

const jqlTestTimeout = 10 * time.Second

// compactErr extracts the useful part from Jira errors.
func compactErr(err error) string {
	msg := err.Error()

	// Network error: "jira request: Post "https://host/path": <reason>"
	if rest, found := strings.CutPrefix(msg, "jira request: "); found {
		if i := strings.Index(rest, `": `); i != -1 {
			rest = rest[i+3:]
		}
		if rest != "" {
			rest = strings.ToUpper(rest[:1]) + rest[1:]
		}
		return rest
	}

	// HTTP error: "jira returned 400: <json body>"
	if rest, found := strings.CutPrefix(msg, "jira returned "); found {
		if status, body, ok := strings.Cut(rest, ": "); ok {
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
		ctx, cancel := context.WithTimeout(context.Background(), jqlTestTimeout)
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

// errNoClient is returned when no Jira client is configured.
var errNoClient = errors.New("no Jira client configured")
