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
			Render("⏳ 执行中...")
		return executing
	}

	type btnDef struct {
		icon  string
		label string
		key   string
		color string
	}

	buttons := []btnDef{
		{"⬇", "PULL", "p", c.theme.BtnPull()},
		{"⬆", "PUSH", "P", c.theme.BtnPush()},
		{"⚕", "DOCTOR", "d", c.theme.BtnDoctor()},
		{"✕", "REMOVE", "x", c.theme.BtnRemove()},
	}

	// Each button gets equal width. Account for 4 buttons total.
	// Each button has a rounded border (1 char each side), so inner width = btnWidth - 2.
	btnWidth := width / len(buttons)
	if btnWidth < 4 {
		btnWidth = 4
	}

	rendered := make([]string, len(buttons))
	for i, b := range buttons {
		colorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(b.color))
		dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Dimmed()))

		label := fmt.Sprintf("%s %s", colorStyle.Render(b.icon+" "+b.label), dimmedStyle.Render(b.key))

		btnStyle := c.styles.Button.
			Width(btnWidth - 2). // subtract border width
			BorderForeground(lipgloss.Color(b.color))

		rendered[i] = btnStyle.Render(label)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
