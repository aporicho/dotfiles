package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (d *Dashboard) View() string {
	if d.width < 60 || d.height < 12 {
		return "终端过小，请调整窗口大小（至少 60×12）"
	}

	lay := ComputeLayout(d.width, d.height, len(d.modules))
	border := lipgloss.NormalBorder()
	borderSty := lipgloss.NewStyle().Foreground(lipgloss.Color(d.theme.PanelBorder()))

	// 1. Channel strip: framed chip grid with horizontal scroll indicators
	d.channel.SetVisibleChips(lay.ChipsPerRow)
	chipContents := d.channel.ChipContents(lay.ChipsPerRow)
	chipWidths := make([]int, len(chipContents))
	for i := range chipWidths {
		chipWidths[i] = lay.ChipW
	}

	// Prepend/append scroll indicator columns when chips overflow
	if d.channel.CanScrollLeft() {
		indicator := lipgloss.NewStyle().
			Width(1).Height(lay.ChipH).
			Align(lipgloss.Center).AlignVertical(lipgloss.Center).
			Foreground(lipgloss.Color(d.theme.Blue())).
			Render("◂")
		chipContents = append([]string{indicator}, chipContents...)
		chipWidths = append([]int{1}, chipWidths...)
	}
	if d.channel.CanScrollRight() {
		indicator := lipgloss.NewStyle().
			Width(1).Height(lay.ChipH).
			Align(lipgloss.Center).AlignVertical(lipgloss.Center).
			Foreground(lipgloss.Color(d.theme.Blue())).
			Render("▸")
		chipContents = append(chipContents, indicator)
		chipWidths = append(chipWidths, 1)
	}

	channelFrame := RenderFrame(border, borderSty, chipWidths, chipContents, lay.ChipH)

	// 2. Middle panels
	var panelWidths []int
	var panelContents []string

	if lay.ShowOverview {
		overviewContent := d.overview.View(lay.Overview.W, lay.PanelInnerH)
		scopeContent := d.scope.View(lay.Scope.W, lay.PanelInnerH)
		terminalContent := d.terminal.View(lay.Terminal.W, lay.PanelInnerH)
		panelWidths = []int{lay.Overview.W, lay.Scope.W, lay.Terminal.W}
		panelContents = []string{overviewContent, scopeContent, terminalContent}
	} else {
		scopeContent := d.scope.View(lay.Scope.W, lay.PanelInnerH)
		terminalContent := d.terminal.View(lay.Terminal.W, lay.PanelInnerH)
		panelWidths = []int{lay.Scope.W, lay.Terminal.W}
		panelContents = []string{scopeContent, terminalContent}
	}
	panelFrame := RenderFrame(border, borderSty, panelWidths, panelContents, lay.PanelInnerH)

	// 3. Controls (borderless)
	controlsView := d.controls.View(lay.Controls.W, lay.Controls.H)

	// 4. Separator
	sep := borderSty.Render(strings.Repeat("─", d.width))

	// 5. Footer
	footer := d.renderFooter()

	// Assemble
	parts := []string{channelFrame, panelFrame}
	if lay.ShowControls {
		parts = append(parts, controlsView)
	}
	parts = append(parts, sep)
	if lay.ShowFooter {
		parts = append(parts, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderFooter renders the key-hint bar at the very bottom.
func (d *Dashboard) renderFooter() string {
	hint := "←→ module · : terminal · esc back · p install · P push · d doctor · x uninstall · a add · q quit"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.FooterFg())).
		Width(d.width)
	return style.Render(hint)
}
