package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // cyan
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	inactiveStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray
	footerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// View renders the picker UI.
func (m Model) View() string {
	if m.Done || m.Quitted {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("选择要安装的模块："))
	b.WriteString("\n\n")

	// Calculate max name length for alignment
	maxLen := 0
	for _, item := range m.Items {
		if len(item.Name) > maxLen {
			maxLen = len(item.Name)
		}
	}

	for i, item := range m.Items {
		// Cursor indicator
		cursor := "  "
		if i == m.Cursor {
			cursor = cursorStyle.Render("\u25b8 ")
		}

		// Checkbox
		check := "[ ]"
		if item.Selected {
			check = selectedStyle.Render("[x]")
		}

		// Module name with padding
		name := fmt.Sprintf("%-*s", maxLen, item.Name)
		if i == m.Cursor {
			name = cursorStyle.Render(name)
		}

		// Status text
		status := item.StatusText()
		if item.IsHighlight() {
			status = highlightStyle.Render(status)
		} else {
			status = inactiveStyle.Render(status)
		}

		b.WriteString(fmt.Sprintf("%s%s %s    %s\n", cursor, check, name, status))
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("\u2191\u2193/jk \u79fb\u52a8 \u00b7 \u7a7a\u683c \u52fe\u9009 \u00b7 a \u5168\u9009 \u00b7 \u56de\u8f66 \u786e\u8ba4 \u00b7 q \u53d6\u6d88"))
	b.WriteString("\n")

	return b.String()
}
