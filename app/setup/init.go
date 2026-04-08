package setup

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnwv/jirka/app/config"
)

const defaultConfig = `# Jirka Dashboard Configuration
# Edit JQL queries below to match your Jira filters.
# Colors: red, blue, green, yellow, orange, purple, teal, cyan

windows:
  - name: Main
    layout:
      rows: 2
      cols: 2
    panels:
      - title: "To Do"
        color: blue
        jql: "assignee = currentUser() AND status = Open ORDER BY priority DESC"
      - title: "In Progress"
        color: green
        jql: "assignee = currentUser() AND status = \"In Progress\" ORDER BY priority DESC"
      - title: "Review"
        color: yellow
        jql: "assignee = currentUser() AND status = \"Code Review\" ORDER BY updated DESC"
      - title: "Done"
        color: teal
        jql: "assignee = currentUser() AND status = Done AND updated >= -7d ORDER BY updated DESC"
`

// Run executes the interactive init setup.
func Run() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to jirka setup!")
	fmt.Println()

	// Check if already configured
	dir := config.ConfigDir()
	envPath := filepath.Join(dir, ".env")
	cfgPath := filepath.Join(dir, "config.yaml")

	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Printf("Config already exists at %s\n", cfgPath)
		fmt.Print("Overwrite? [y/N] ")
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Ask for Jira type
	fmt.Println("Jira type:")
	fmt.Println("  [1] Jira Cloud (*.atlassian.net)")
	fmt.Println("  [2] Jira Server / Data Center")
	fmt.Print("Choose [1/2]: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	isCloud := choice == "1"

	// Ask for Jira URL
	if isCloud {
		fmt.Print("Jira URL (e.g. https://mycompany.atlassian.net): ")
	} else {
		fmt.Print("Jira URL (e.g. https://jira.company.com): ")
	}
	jiraURL, _ := reader.ReadString('\n')
	jiraURL = strings.TrimSpace(jiraURL)
	if jiraURL == "" {
		return errors.New("jira URL is required")
	}

	var email, token string

	if isCloud {
		fmt.Print("Email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email == "" {
			return errors.New("email is required for Jira Cloud")
		}

		fmt.Print("API token (https://id.atlassian.net/manage-profile/security/api-tokens): ")
		token, _ = reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return errors.New("API token is required")
		}
	} else {
		fmt.Print("Personal access token: ")
		token, _ = reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return errors.New("token is required")
		}
	}

	// Create config directory
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Write .env
	var envContent string
	if isCloud {
		envContent = fmt.Sprintf("JIRA_URL=%s\nJIRA_AUTH_TYPE=basic\nJIRA_EMAIL=%s\nJIRA_TOKEN=%s\n", jiraURL, email, token)
	} else {
		envContent = fmt.Sprintf("JIRA_URL=%s\nJIRA_AUTH_TYPE=bearer\nJIRA_TOKEN=%s\n", jiraURL, token)
	}
	if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}
	fmt.Printf("Wrote %s\n", envPath)

	// Write config.yaml
	if err := os.WriteFile(cfgPath, []byte(defaultConfig), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Printf("Wrote %s\n", cfgPath)

	fmt.Println()
	fmt.Println("Done! Run `jirka` to start.")
	fmt.Printf("Edit %s to customize your dashboard panels.\n", cfgPath)

	return nil
}
