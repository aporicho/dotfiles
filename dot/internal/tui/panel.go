package tui

import tea "github.com/charmbracelet/bubbletea"

// Panel is the interface all dashboard panels implement.
type Panel interface {
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View(width, height int) string
	Focused() bool
	SetFocus(bool)
	Weight(total int) int // column width weight; total = number of panels in this row
}
