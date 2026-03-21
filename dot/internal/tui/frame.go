package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderFrame draws a box-drawing border around N columns of content.
// colWidths[i] is the content width of column i (excluding border chars).
// contents[i] is the pre-rendered content block for column i,
// rendered to exactly colWidths[i] wide and innerHeight tall.
// All border characters are styled with borderSty.
func RenderFrame(
	border lipgloss.Border,
	borderSty lipgloss.Style,
	colWidths []int,
	contents []string,
	innerHeight int,
) string {
	// Split each content block into lines
	contentLines := make([][]string, len(contents))
	for i, c := range contents {
		contentLines[i] = strings.Split(c, "\n")
	}

	var lines []string

	// Top border: ┌───┬───┬───┐
	lines = append(lines, buildEdge(border, borderSty, colWidths, true))

	// Content rows
	for r := 0; r < innerHeight; r++ {
		lines = append(lines, buildRow(border, borderSty, colWidths, contentLines, r))
	}

	// Bottom border: └───┴───┴───┘
	lines = append(lines, buildEdge(border, borderSty, colWidths, false))

	return strings.Join(lines, "\n")
}

// buildEdge builds a horizontal border line (top or bottom).
func buildEdge(b lipgloss.Border, sty lipgloss.Style, colWidths []int, isTop bool) string {
	var buf strings.Builder

	for i, w := range colWidths {
		if i == 0 {
			if isTop {
				buf.WriteString(b.TopLeft)
			} else {
				buf.WriteString(b.BottomLeft)
			}
		} else {
			if isTop {
				buf.WriteString(b.MiddleTop)
			} else {
				buf.WriteString(b.MiddleBottom)
			}
		}
		buf.WriteString(strings.Repeat(b.Top, w))
	}

	if isTop {
		buf.WriteString(b.TopRight)
	} else {
		buf.WriteString(b.BottomRight)
	}

	return sty.Render(buf.String())
}

// buildRow builds a single content row with vertical borders.
func buildRow(b lipgloss.Border, sty lipgloss.Style, colWidths []int, contentLines [][]string, row int) string {
	var buf strings.Builder

	buf.WriteString(sty.Render(b.Left))

	for i := range colWidths {
		if i > 0 {
			buf.WriteString(sty.Render(b.Left)) // column separator (│)
		}

		line := ""
		if row < len(contentLines[i]) {
			line = contentLines[i][row]
		}
		buf.WriteString(line)
	}

	buf.WriteString(sty.Render(b.Right))

	return buf.String()
}
