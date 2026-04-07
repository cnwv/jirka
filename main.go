package main

import (
	"buble_jira/internal/config"
	"buble_jira/internal/jira"
	"buble_jira/internal/tui"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	var client *jira.Client
	if cfg.JiraToken != "" {
		client = jira.NewClient(cfg.JiraURL, cfg.JiraToken)
	}

	root := tui.NewRootModel(client, cfg.PollInterval, cfg.JiraURL, cfg.Dashboard)
	p := tea.NewProgram(root)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
