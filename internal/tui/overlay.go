package tui

import (
	"fmt"
	"strings"
)

// overlayCenter renders popupContent centered over mainContent.
// mainW/mainH are the full screen dimensions.
func overlayCenter(mainContent, popupContent string, mainW, mainH int) string {
	popupLines := strings.Split(popupContent, "\n")
	popupH := len(popupLines)
	popupW := 0
	for _, l := range popupLines {
		if w := visibleWidth(l); w > popupW {
			popupW = w
		}
	}

	startRow := (mainH - popupH) / 2
	startCol := (mainW - popupW) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	mainLines := strings.Split(mainContent, "\n")
	// Ensure mainLines has at least mainH entries
	for len(mainLines) < mainH {
		mainLines = append(mainLines, "")
	}

	for pi, pLine := range popupLines {
		row := startRow + pi
		if row >= len(mainLines) {
			break
		}
		mainLines[row] = overlayLine(mainLines[row], pLine, startCol, mainW)
	}

	return strings.Join(mainLines, "\n")
}

// overlayLine replaces the region starting at col in base with overlay.
// Strips ANSI from base for simplicity in the replaced region.
func overlayLine(base, overlay string, col, totalW int) string {
	// Left padding up to col
	baseRunes := []rune(stripAnsiStr(base))
	var left string
	if col > 0 && col <= len(baseRunes) {
		left = string(baseRunes[:col])
	} else if col > len(baseRunes) {
		left = string(baseRunes) + strings.Repeat(" ", col-len(baseRunes))
	}

	overlayW := visibleWidth(overlay)
	endCol := col + overlayW

	var right string
	if endCol < len(baseRunes) {
		right = string(baseRunes[endCol:])
	}

	result := left + overlay + right
	// Pad/truncate to totalW
	vis := visibleWidth(result)
	if vis < totalW {
		result += strings.Repeat(" ", totalW-vis)
	} else if vis > totalW {
		result = truncateAnsi(result, totalW)
	}
	return result
}

func stripAnsiStr(s string) string {
	var out strings.Builder
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

// popupBox renders a titled box with given content lines and width.
// Returns a multi-line string.
func popupBox(title string, lines []string, width int) string {
	if width < 10 {
		width = 10
	}
	inner := width - 2 // minus left/right borders

	titleFmt := fmt.Sprintf(" %s ", title)
	topBar := "╭" + centerTitle(titleFmt, inner) + "╮"

	var sb strings.Builder
	sb.WriteString(topBar + "\n")
	for _, l := range lines {
		vis := visibleWidth(l)
		if vis < inner {
			l += strings.Repeat(" ", inner-vis)
		} else if vis > inner {
			l = truncateAnsi(l, inner)
		}
		sb.WriteString("│" + l + "│\n")
	}
	sb.WriteString("╰" + strings.Repeat("─", inner) + "╯")
	return sb.String()
}

func centerTitle(title string, width int) string {
	vis := visibleWidth(title)
	if vis >= width {
		return truncateAnsi(title, width)
	}
	left := (width - vis) / 2
	right := width - vis - left
	return strings.Repeat("─", left) + title + strings.Repeat("─", right)
}
