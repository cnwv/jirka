# jirka

TUI dashboard for Jira tickets, built with BubbleTea v2.

## Commands
- Build: `make build` (output: `.bin/jirka`)
- Test: `make test` (race detector + coverage)
- Lint: `make lint` or `golangci-lint run`
- Run: `go run ./app` (requires `.env` with JIRA_URL and JIRA_TOKEN)

## Project Structure
- `app/` — entry point (`main.go`), CLI flags (`-v`, `-h`, `init`), wiring
- `app/config/` — `.env` + YAML config loading, dashboard types, color resolution, embedded example config
- `app/jira/` — REST client (`POST /rest/api/2/search`), Bearer token auth, 50-ticket limit per query
- `app/model/` — shared `Ticket` struct
- `app/tui/` — BubbleTea TUI model, views, popups, rendering
- `app/setup/` — interactive `jirka init` setup wizard

## Key Types
- `config.Config` — top-level config: JiraURL, JiraToken, PollInterval, Dashboard
- `config.DashboardConfig` — multi-window layout with panels, sections, stubs
- `config.WindowConfig` — grid layout (rows × cols) + panel array
- `jira.Client` — HTTP client with `SearchJQL()`, `CountJQL()`, shared `doSearch()` method
- `model.Ticket` — IssueKey, Summary, Status, Priority, Assignee, Description, etc.
- `tui.RootModel` — top-level `tea.Model`, owns panels + detail + status bar + popups

## TUI File Map
- `root.go` — `RootModel`: keyboard dispatch, focus management, window switching, layout, popup routing
- `fetch.go` — background Jira polling: `doFetch`, `waitForTickets`, `buildQueries`, `assembleResults`, `scheduleTick`
- `browser.go` — open ticket in browser (`open`/`xdg-open`), clipboard read/paste dispatch
- `jql.go` — JQL test command, Jira error formatting (`compactErr`)
- `border.go` — shared `borderedBox()` renderer (panels + detail use the same function)
- `panel.go` — `PanelModel`: bordered ticket list with priority icons, sections, cursor, scroll
- `ticketdetail.go` — `TicketDetailModel`: ticket fields + scrollable Jira-formatted description
- `statusbar.go` — bottom bar: last update time, error, hint, key bindings
- `messages.go` — custom `tea.Msg` types: `TicketsRefreshedMsg`, `TickMsg`, `ErrorMsg`, `jqlTestResultMsg`
- `textinput.go` — single-line text input widget: cursor, maxLen, paste, truncation
- `overlay.go` — popup overlay helpers: `overlayCenter`, `overlayLine`, `popupBox`, `stripAnsiStr`
- `jiraformat.go` — Jira wiki markup → terminal ANSI converter (bold, links, images, headers, code blocks)
- `popup_window_switcher.go` — window list popup (key `0`)
- `popup_new_window.go` — new window creation: name input + grid preset picker
- `popup_panel_edit.go` — panel editor: name, JQL (with live test), color selector

## Data Flow
```
Init() → doFetch() + waitForTickets() + scheduleTick()
  doFetch() goroutine: buildQueries(window) → sequential SearchJQL per query (150ms delay)
    → assembleResults() → fetchResult{ByPanel, BySections} → channel
  waitForTickets() blocks on channel → emits TicketsRefreshedMsg
  Update() receives msg → distributes tickets to panels → re-renders
  scheduleTick() → TickMsg after pollInterval → triggers next doFetch()
```

## Config
- Config dir: `~/.config/jirka/` (`.env` + `config.yaml`)
- Search order: `$CONFIG_PATH` → `~/.config/jirka/config.yaml` → `./config.yaml` → embedded example
- `.env` search: `~/.config/jirka/.env` then `./.env`
- `config.example.yaml` embedded via `go:embed` in `app/config/embed.go` for demo mode
- Legacy single-window format (layout + panels at root) auto-normalized to `windows[0]` on load
- Config saved back to disk only when `configPath != ""` (not in demo mode)
- Colors: named (`red`, `blue`, `green`, `yellow`, `orange`, `purple`, `teal`, `cyan`) or raw ANSI (`38;5;203`)

## Layout System
- Left half: configurable grid (rows × cols, max 6 panels). Right half: always detail view.
- `recalcLayout()` distributes widths/heights evenly, remainder to last column/row
- `View()` renders line-by-line: for each screen row, find grid row via `rowStart[]`, compose panel lines left-to-right, append detail line
- All rendering is raw ANSI — no lipgloss, direct `\033[...m` sequences for performance
- `borderedBox()` handles top/bottom borders + content padding to exact width/height — used by both panels and detail

## Gotchas
- `visibleWidth()` and `truncateAnsi()` are the ANSI-aware string measurement functions — all rendering depends on them matching terminal behavior. Unicode box-drawing chars and emoji can cause width mismatches.
- `doFetch()` runs queries sequentially with 150ms delay between requests to avoid Jira rate limiting
- Popup state is `any` type (`*winSwitcherPopup | *newWindowPopup | *panelEditPopup`) — nil means no popup active
- `clipboardReadMsg` uses platform-specific commands (`pbpaste` on macOS, `xclip` on Linux) via `exec.CommandContext`
- `jiraformat.go` regex patterns are compiled once at package level (`var reJiraBold = ...`)
- `textinput.handleKey()` has special handling for multi-byte runes via `utf8.DecodeRuneInString`
- `config.example.yaml` exists in two places: repo root (for local dev) and `app/config/` (for `go:embed`) — keep them in sync
