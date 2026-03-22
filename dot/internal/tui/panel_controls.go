package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Controls is the action button bar.
type Controls struct {
	executing   bool
	confirming  bool
	confirmName string
	focused     bool
	styles      Styles
	theme       Theme
}

func NewControls(styles Styles, theme Theme) *Controls {
	return &Controls{styles: styles, theme: theme}
}

func (c *Controls) Focused() bool        { return c.focused }
func (c *Controls) SetFocus(f bool)       { c.focused = f }
func (c *Controls) SetExecuting(v bool)   { c.executing = v }
func (c *Controls) SetConfirming(v bool)  { c.confirming = v }
func (c *Controls) SetConfirmName(n string) { c.confirmName = n }

func (c *Controls) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg.(type) {
	case CmdStartMsg:
		c.executing = true
	case CmdOutputMsg:
		c.executing = false
	}
	return c, nil
}

func (c *Controls) View(width, _ int) string { return "" }

// BuildRow produces a Row for the controls bar.
// panelRow is the middle-panel Row whose separators we align to.
func (c *Controls) BuildRow(totalW int, panelRow Row) Row {
	cols := c.alignedCols(panelRow.Cols)
	contents := c.renderCells(cols)
	return Row{Cols: cols, Contents: contents, Height: 1}
}

// alignedCols produces 4 button column widths aligned to panel separators.
// 3 panels [ow, sw, tw]: buttons = [ow, (sw-1)/2, sw-1-(sw-1)/2, tw]
// 2 panels [sw, tw]:      buttons = [(sw-1)/2, sw-1-(sw-1)/2, (tw-1)/2, tw-1-(tw-1)/2]
func (c *Controls) alignedCols(panelCols []int) []int {
	split := func(w int) (int, int) {
		left := (w - 1) / 2
		return left, w - 1 - left
	}
	if len(panelCols) == 3 {
		l, r := split(panelCols[1])
		return []int{panelCols[0], l, r, panelCols[2]}
	}
	if len(panelCols) == 2 {
		l1, r1 := split(panelCols[0])
		l2, r2 := split(panelCols[1])
		return []int{l1, r1, l2, r2}
	}
	// Fallback: single column
	return []int{panelCols[0]}
}

// renderCells returns one content string per column.
func (c *Controls) renderCells(cols []int) []string {
	if c.confirming {
		return c.spanCells(cols,
			lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Red())),
			fmt.Sprintf("确认卸载 %s？Y/N", c.confirmName))
	}
	if c.executing {
		return c.spanCells(cols,
			lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Yellow())),
			"\uf110 执行中...")
	}

	type btn struct {
		icon, label, key, color string
	}
	buttons := []btn{
		{"\uf0ed", "INSTALL", "p", c.theme.BtnPull()},
		{"\uf0ee", "PUSH", "P", c.theme.BtnPush()},
		{"\uf21e", "DOCTOR", "d", c.theme.BtnDoctor()},
		{"\uf1f8", "UNINSTALL", "x", c.theme.BtnRemove()},
	}

	cells := make([]string, len(cols))
	for i := range cols {
		if i < len(buttons) {
			b := buttons[i]
			color := lipgloss.NewStyle().Foreground(lipgloss.Color(b.color))
			dim := lipgloss.NewStyle().Foreground(lipgloss.Color(c.theme.Dimmed()))
			text := fmt.Sprintf("%s %s", color.Render(b.icon+" "+b.label), dim.Render(b.key))
			cells[i] = lipgloss.NewStyle().Width(cols[i]).Align(lipgloss.Center).Render(text)
		} else {
			cells[i] = lipgloss.NewStyle().Width(cols[i]).Render("")
		}
	}
	return cells
}

func (c *Controls) spanCells(cols []int, sty lipgloss.Style, msg string) []string {
	cells := make([]string, len(cols))
	cells[0] = lipgloss.NewStyle().Width(cols[0]).Align(lipgloss.Center).Render(sty.Render(msg))
	for i := 1; i < len(cols); i++ {
		cells[i] = lipgloss.NewStyle().Width(cols[i]).Render("")
	}
	return cells
}
