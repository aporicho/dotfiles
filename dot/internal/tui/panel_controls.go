package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Controls is the full-width horizontal bar with 4 action buttons.
type Controls struct {
	executing   bool
	confirming  bool
	confirmName string
	focused     bool
	styles      Styles
	theme       Theme
}

// NewControls constructs a Controls panel.
func NewControls(styles Styles, theme Theme) *Controls {
	return &Controls{
		styles: styles,
		theme:  theme,
	}
}

// Focused implements Panel.
func (c *Controls) Focused() bool { return c.focused }

// SetFocus implements Panel.
func (c *Controls) SetFocus(f bool) { c.focused = f }

// SetExecuting sets the executing state directly.
func (c *Controls) SetExecuting(executing bool) { c.executing = executing }

func (c *Controls) SetConfirming(v bool)    { c.confirming = v }
func (c *Controls) SetConfirmName(n string) { c.confirmName = n }

// Update implements Panel.
func (c *Controls) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg.(type) {
	case CmdStartMsg:
		c.executing = true
	case CmdOutputMsg:
		c.executing = false
	}
	return c, nil
}

// View implements Panel. Returns a single-line string (used as fallback).
func (c *Controls) View(width, _ int) string {
	cells := c.ViewCells([]int{width})
	return cells[0]
}

// ViewCells returns one content string per column, for use with the unified frame.
// When confirming or executing, all columns are merged into a single centered message
// spread across the provided widths.
func (c *Controls) ViewCells(colWidths []int) []string {
	totalW := 0
	for _, w := range colWidths {
		totalW += w
	}

	if c.confirming {
		return c.spanMessage(colWidths,
			lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Red())),
			fmt.Sprintf("确认卸载 %s？Y/N", c.confirmName))
	}

	if c.executing {
		return c.spanMessage(colWidths,
			lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Yellow())),
			"\uf110 执行中...")
	}

	type btnDef struct {
		icon  string
		label string
		key   string
		color string
	}

	buttons := []btnDef{
		{"\uf0ed", "INSTALL", "p", c.theme.BtnPull()},
		{"\uf0ee", "PUSH", "P", c.theme.BtnPush()},
		{"\uf21e", "DOCTOR", "d", c.theme.BtnDoctor()},
		{"\uf1f8", "UNINSTALL", "x", c.theme.BtnRemove()},
	}

	cells := make([]string, len(colWidths))
	for i := 0; i < len(colWidths) && i < len(buttons); i++ {
		b := buttons[i]
		w := colWidths[i]

		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.color))
		dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Dimmed()))

		label := fmt.Sprintf("%s %s", colorStyle.Render(b.icon+" "+b.label), dimmedStyle.Render(b.key))

		cells[i] = lipgloss.NewStyle().
			Width(w).
			Align(lipgloss.Center).
			Render(label)
	}
	// Fill remaining columns (if any) with empty space
	for i := len(buttons); i < len(colWidths); i++ {
		cells[i] = lipgloss.NewStyle().Width(colWidths[i]).Render("")
	}

	return cells
}

// spanMessage renders a centered message spanning all columns.
func (c *Controls) spanMessage(colWidths []int, style lipgloss.Style, msg string) []string {
	cells := make([]string, len(colWidths))
	// Put the message centered in all space
	// First cell gets the message, rest are empty
	totalInner := 0
	for _, w := range colWidths {
		totalInner += w
	}
	// We need to distribute the message across cells to match the column widths
	rendered := style.Render(msg)
	// Put it all in the first cell, padded to its width
	cells[0] = lipgloss.NewStyle().Width(colWidths[0]).Align(lipgloss.Center).Render(rendered)
	for i := 1; i < len(colWidths); i++ {
		cells[i] = lipgloss.NewStyle().Width(colWidths[i]).Render("")
	}
	return cells
}
