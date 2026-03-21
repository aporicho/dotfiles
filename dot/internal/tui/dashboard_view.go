package tui

import "github.com/charmbracelet/lipgloss"

// View implements tea.Model. Lays out the dashboard panels.
func (d *Dashboard) View() string {
	if d.width < 80 || d.height < 24 {
		return "终端过小，请调整窗口大小（至少 80×24）"
	}

	const (
		channelHeight  = 3
		controlsHeight = 3
		footerHeight   = 1
	)

	middleHeight := d.height - channelHeight - controlsHeight - footerHeight
	if middleHeight < 1 {
		middleHeight = 1
	}

	overviewWidth := 22
	remaining := d.width - overviewWidth
	scopeWidth := remaining / 2
	terminalWidth := remaining - scopeWidth

	// Render each section.
	channelView := d.channel.View(d.width, channelHeight)

	overviewView := d.overview.View(overviewWidth, middleHeight)
	scopeView := d.scope.View(scopeWidth, middleHeight)
	terminalView := d.terminal.View(terminalWidth, middleHeight)

	middleRow := lipgloss.JoinHorizontal(lipgloss.Top,
		overviewView,
		scopeView,
		terminalView,
	)

	controlsView := d.controls.View(d.width, controlsHeight)

	footer := d.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		channelView,
		middleRow,
		controlsView,
		footer,
	)
}

// renderFooter renders the key-hint bar at the very bottom.
func (d *Dashboard) renderFooter() string {
	hint := "←→ module · tab panel · : terminal · esc back · p pull · P push · d doctor · x remove · a add · q quit"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.FooterFg())).
		Background(lipgloss.Color(d.theme.FooterBg())).
		Width(d.width)
	return style.Render(hint)
}
