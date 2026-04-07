package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
)

const browserTimeout = 5 * time.Second

func (m *RootModel) openInBrowser() {
	t := m.currentTicket()
	if t == nil {
		return
	}
	url := fmt.Sprintf("%s/browse/%s", m.jiraURL, t.IssueKey)
	ctx, cancel := context.WithTimeout(context.Background(), browserTimeout)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url) //nolint:gosec // URL is from internal config, not user input
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url) //nolint:gosec // URL is from internal config, not user input
	default:
		cmd = exec.CommandContext(ctx, "open", url) //nolint:gosec // URL is from internal config, not user input
	}
	_ = cmd.Start()
}

type clipboardReadMsg string

func readSystemClipboard() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), browserTimeout)
		defer cancel()

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.CommandContext(ctx, "pbpaste")
		case "linux":
			cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-o")
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
