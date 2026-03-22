package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Controls is the full-width horizontal bar with 4 action buttons.
type Controls struct {
	executing bool
	focused   bool
	styles    Styles
	theme     Theme
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

// View implements Panel.
func (c *Controls) View(width, _ int) string {
	if c.executing {
		executing := lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.theme.Yellow())).
			Width(width).
			Align(lipgloss.Center).
			Render("\uf110 执行中...")
		return executing
	}

	type btnDef struct {
		icon  string
		label string
		key   string
		color string
	}

	buttons := []btnDef{
		{"\uf0ed", "PULL", "p", c.theme.BtnPull()},
		{"\uf0ee", "PUSH", "P", c.theme.BtnPush()},
		{"\uf21e", "DOCTOR", "d", c.theme.BtnDoctor()},
		{"\uf1f8", "REMOVE", "x", c.theme.BtnRemove()},
	}

	n := len(buttons)
	btnWidth := width / n
	// Distribute remainder evenly: first R buttons each get +1
	remainder := width - btnWidth*n

	rendered := make([]string, n)
	for i, b := range buttons {
		w := btnWidth
		if i < remainder {
			w++
		}

		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.color))
		dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Dimmed()))

		label := fmt.Sprintf("%s %s", colorStyle.Render(b.icon+" "+b.label), dimmedStyle.Render(b.key))

		btnStyle := lipgloss.NewStyle().
			Width(w).
			Align(lipgloss.Center)

		rendered[i] = btnStyle.Render(label)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
