# jirka

TUI dashboard for Jira tickets. Multi-panel grid layout with configurable JQL queries, live polling, and ticket detail view — all in your terminal.

Built for teams that live in the terminal and want a quick overview of their Jira boards without switching to a browser.

## Features

- **Multi-panel grid** — up to 6 panels arranged in a configurable rows × cols layout
- **JQL-powered** — each panel shows tickets from any JQL query
- **Sectioned panels** — group tickets within a panel by status, priority, or any criteria
- **Multi-window** — create multiple dashboard windows and switch between them
- **Live polling** — refreshes data from Jira every 5 minutes (configurable)
- **Ticket detail** — right-side panel shows full ticket info with scrollable Jira-formatted description
- **In-app editing** — edit panel name, JQL, and color without leaving the TUI
- **JQL testing** — test your JQL queries live before saving
- **Browser integration** — open any ticket in your browser with one key
- **Demo mode** — works out of the box with stub data, no Jira connection needed
- **Keyboard-driven** — Vim-style navigation, numbered panel switching

## Installation

**Homebrew (macOS/Linux):**

```bash
brew install cnwv/apps/jirka
```

**Go:**

```bash
go install github.com/cnwv/jirka/app@latest
```

**Binary releases:** download from [GitHub Releases](https://github.com/cnwv/jirka/releases).

## Quick Start

```bash
# 1. Run in demo mode to see the interface
jirka

# 2. Set up your Jira connection
jirka init

# 3. Start using with real data
jirka
```

`jirka init` creates `~/.config/jirka/.env` (credentials) and `~/.config/jirka/config.yaml` (dashboard layout).

## Usage

```
jirka              Start the dashboard
jirka init         Interactive setup wizard
jirka -v           Show version
jirka -h           Show help
```

## Configuration

### Credentials

Stored in `~/.config/jirka/.env`:

```
JIRA_URL=https://jira.company.com
JIRA_TOKEN=your_bearer_token
```

### Dashboard Layout

Stored in `~/.config/jirka/config.yaml`:

```yaml
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
        jql: "assignee = currentUser() AND status = 'In Progress' ORDER BY priority DESC"

      - title: "Review"
        color: yellow
        jql: "assignee = currentUser() AND status = 'Code Review' ORDER BY updated DESC"

      - title: "Blocked"
        color: red
        sections:
          - name: CRITICAL
            color: red
            jql: "priority = Critical AND status = Blocked"
          - name: OTHER
            color: orange
            jql: "priority != Critical AND status = Blocked"
```

### Layout Rules

- `rows × cols` must equal the number of panels (max 6)
- Supported grids: 1×1, 1×2, 2×1, 2×2, 3×1, 1×3, 2×3, 3×2
- Each panel has either `jql` (flat ticket list) or `sections` (grouped by sub-queries)
- Colors: `red`, `blue`, `green`, `yellow`, `orange`, `purple`, `teal`, `cyan` — or raw ANSI like `38;5;203`

### Multiple Windows

Add more windows to your config and switch between them with `0`:

```yaml
windows:
  - name: My Team
    layout: { rows: 2, cols: 2 }
    panels: [...]

  - name: Support Queue
    layout: { rows: 3, cols: 2 }
    panels: [...]
```

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `1`–`N` | Focus panel N |
| `N+1` | Focus detail view |
| `Tab` / `Shift+Tab` | Cycle focus between panels and detail |
| `↑`/`↓` or `k`/`j` | Navigate tickets or scroll detail |

### Actions

| Key | Action |
|-----|--------|
| `e` | Edit focused panel (name, JQL, color) |
| `n` | Create new window |
| `0` | Window switcher (switch, create, delete) |
| `r` | Refresh data from Jira |
| `b` | Open selected ticket in browser |
| `q` | Quit |

### In Popups

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous field |
| `Enter` | Next step or test JQL query |
| `Ctrl+S` | Save changes |
| `Ctrl+V` | Paste from clipboard |
| `Esc` | Cancel |

## Priority Icons

| Icon | Priority |
|------|----------|
| ⛔ | Blocker / Show-stopper |
| ⏶ | Critical |
| ∧ | Major |
| ∨ | Minor / Trivial |
| ● | Other |

## How It Works

jirka polls the Jira REST API (`POST /rest/api/2/search`) using Bearer token authentication. Each panel's JQL query is executed sequentially with a small delay between requests to avoid rate limiting. Results are displayed in a BubbleTea v2 terminal UI with a grid of panels on the left and a ticket detail view on the right.

No data is stored locally — tickets are fetched fresh on each poll cycle.

## License

[MIT](LICENSE)
