package tui

import (
	"fmt"
	"strings"
)

const (
	borderColorDefault = "38;5;240"
	borderColorFocused = "38;5;245"
)

// borderedBox renders content lines inside a box with a titled top border.
// The total output is exactly height lines tall and width visible chars wide.
func borderedBox(title, titleColor string, contentLines []string, width, height int, focused bool) string {
	if width < 3 || height < 3 {
		return ""
	}

	innerWidth := width - 2
	borderColor := borderColorDefault
	if focused {
		borderColor = borderColorFocused
	}

	titleRendered := "\033[1;" + titleColor + "m" + title + "\033[0m"
	leftBorder := "\033[" + borderColor + "m│\033[0m"
	rightBorder := leftBorder

	var out strings.Builder
	out.Grow(height * (innerWidth + 20))

	// Top border: ╭─ Title ─╮
	fmt.Fprintf(&out, "\033[%sm╭─\033[0m %s \033[%sm", borderColor, titleRendered, borderColor)
	topUsed := 1 + 1 + len(title) + 1 // "─ " + title + " "
	for topUsed < innerWidth {
		out.WriteString("─")
		topUsed++
	}
	out.WriteString("╮\033[0m\n")

	// Content lines with side borders
	contentH := height - 2
	for i := range contentH {
		out.WriteString(leftBorder)

		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}

		lineW := visibleWidth(line)
		if lineW > innerWidth {
			out.WriteString(truncateAnsi(line, innerWidth))
		} else {
			out.WriteString(line)
			for j := lineW; j < innerWidth; j++ {
				out.WriteByte(' ')
			}
		}

		out.WriteString(rightBorder)
		if i < contentH-1 {
			out.WriteByte('\n')
		}
	}

	// Bottom border: ╰──╯
	out.WriteByte('\n')
	fmt.Fprintf(&out, "\033[%sm╰", borderColor)
	for range innerWidth {
		out.WriteString("─")
	}
	out.WriteString("╯\033[0m")

	return out.String()
}
