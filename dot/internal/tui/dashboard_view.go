package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (d *Dashboard) View() string {
	if d.width < 60 || d.height < 12 {
		return "终端过小，请调整窗口大小（至少 60×12）"
	}

	lay := ComputeLayout(d.width, d.height, len(d.modules))
	borderSty := lipgloss.NewStyle().Foreground(lipgloss.Color(d.theme.PanelBorder()))

	// --- Build rows ---
	var rows []Row

	// Row 1: Channel Strip chips
	d.channel.SetVisibleChips(len(lay.ChipCols))
	chipContents := d.channel.ChipContents(lay.ChipCols)
	// Pad/truncate to match layout columns
	for len(chipContents) < len(lay.ChipCols) {
		chipContents = append(chipContents, "")
	}
	if len(chipContents) > len(lay.ChipCols) {
		chipContents = chipContents[:len(lay.ChipCols)]
	}
	rows = append(rows, Row{Cols: lay.ChipCols, Contents: chipContents, Height: lay.ChipH})

	// Row 2: Middle panels
	if lay.ShowOverview {
		ov := d.overview.View(lay.PanelCols[0], lay.PanelH)
		sc := d.scope.View(lay.PanelCols[1], lay.PanelH)
		tm := d.terminal.View(lay.PanelCols[2], lay.PanelH)
		rows = append(rows, Row{Cols: lay.PanelCols, Contents: []string{ov, sc, tm}, Height: lay.PanelH})
	} else {
		sc := d.scope.View(lay.PanelCols[0], lay.PanelH)
		tm := d.terminal.View(lay.PanelCols[1], lay.PanelH)
		rows = append(rows, Row{Cols: lay.PanelCols, Contents: []string{sc, tm}, Height: lay.PanelH})
	}

	// Row 3: Controls (optional)
	if lay.ShowControls {
		ctrlContents := d.controls.ViewCells(lay.CtrlCols)
		rows = append(rows, Row{Cols: lay.CtrlCols, Contents: ctrlContents, Height: 1})
	}

	// Row 4: Footer (optional)
	if lay.ShowFooter {
		footer := d.renderFooter(lay.FooterW)
		rows = append(rows, Row{Cols: []int{lay.FooterW}, Contents: []string{footer}, Height: 1})
	}

	// Title
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Dimmed())).
		Render("\uf054 CHANNEL STRIP")

	return RenderUnifiedFrame(borderSty, title, d.width, rows)
}

// renderFooter renders the key-hint content (without borders — frame adds them).
func (d *Dashboard) renderFooter(width int) string {
	hint := "←→ module · : terminal · esc back · p install · P push · d doctor · x uninstall · a add · q quit"
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.FooterFg())).
		Width(width)
	return style.Render(hint)
}
