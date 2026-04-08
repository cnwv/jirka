package main

import (
	"fmt"
	"os"

	"github.com/cnwv/jirka/app/config"
	"github.com/cnwv/jirka/app/jira"
	"github.com/cnwv/jirka/app/setup"
	"github.com/cnwv/jirka/app/tui"

	tea "charm.land/bubbletea/v2"
)

var revision = "unknown"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("jirka %s\n", revision)
			return
		case "-h", "--help":
			printHelp()
			return
		case "init":
			if err := setup.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Show hint if not configured
	if !config.Configured() {
		fmt.Println("No configuration found. Starting in demo mode.")
		fmt.Println("Run `jirka init` to connect your Jira.")
		fmt.Println()
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	var client *jira.Client
	if cfg.JiraToken != "" {
		switch cfg.JiraAuthType {
		case "basic":
			client = jira.NewClientBasic(cfg.JiraURL, cfg.JiraEmail, cfg.JiraToken)
		default:
			client = jira.NewClient(cfg.JiraURL, cfg.JiraToken)
		}
	}

	root := tui.NewRootModel(client, cfg.PollInterval, cfg.JiraURL, cfg.Dashboard)
	p := tea.NewProgram(root)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`jirka %s — TUI dashboard for Jira tickets

Usage:
  jirka              Start the dashboard
  jirka init         Interactive setup (creates ~/.config/jirka/)
  jirka -v           Show version
  jirka -h           Show this help

Config files:
  ~/.config/jirka/.env          JIRA_URL, JIRA_TOKEN, JIRA_AUTH_TYPE, JIRA_EMAIL
  ~/.config/jirka/config.yaml   Dashboard layout and JQL queries

Environment variables:
  JIRA_URL         Jira instance URL
  JIRA_TOKEN       API token or personal access token
  JIRA_AUTH_TYPE   "bearer" (default, Server/DC) or "basic" (Cloud)
  JIRA_EMAIL       Email for basic auth (Jira Cloud)

Keyboard shortcuts:
  1-N         Focus panel N
  Tab         Cycle panels
  ↑/↓ or k/j Navigate tickets
  e           Edit panel (name, JQL, color)
  n           New window
  0           Switch window
  r           Refresh data
  b           Open ticket in browser
  q           Quit
`, revision)
}
