package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (d *Dashboard) View() string {
	if d.width < 60 || d.height < 12 {
		return "终端过小，请调整窗口大小（至少 60×12）"
	}

	lay := ComputeLayout(d.width, d.height)
	borderSty := lipgloss.NewStyle().Foreground(lipgloss.Color(d.theme.PanelBorder()))

	// Each panel builds its own Row.
	var rows []Row

	// Row 1: Channel Strip (owns its column structure)
	chipRow := d.channel.BuildRow(lay.TotalW)
	rows = append(rows, chipRow)

	// Row 2: Middle panels
	panelRow := buildPanelRow(lay.TotalW, lay.PanelH, lay.ShowOverview, d.overview, d.scope, d.terminal)
	rows = append(rows, panelRow)

	// Row 3: Controls (aligns to panel separators)
	if lay.ShowControls {
		rows = append(rows, d.controls.BuildRow(lay.TotalW, panelRow))
	}

	// Row 4: Footer
	if lay.ShowFooter {
		rows = append(rows, buildFooterRow(lay.TotalW, d.theme))
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Dimmed())).
		Render("\uf054 CHANNEL STRIP")

	return RenderUnifiedFrame(borderSty, title, d.width, rows)
}

// buildPanelRow produces the middle-panel Row. It owns the column ratio logic.
func buildPanelRow(totalW, panelH int, showOverview bool, ov *Overview, sc *Scope, tm *Terminal) Row {
	if showOverview {
		pw := totalW - 4 // 4 border chars for 3 columns
		if pw < 3 {
			pw = 3
		}
		ow := pw / 4
		tw := pw / 4
		sw := pw - ow - tw
		cols := []int{ow, sw, tw}
		contents := []string{
			ov.View(ow, panelH),
			sc.View(sw, panelH),
			tm.View(tw, panelH),
		}
		return Row{Cols: cols, Contents: contents, Height: panelH}
	}

	pw := totalW - 3 // 3 border chars for 2 columns
	if pw < 2 {
		pw = 2
	}
	sw := pw / 2
	tw := pw - sw
	cols := []int{sw, tw}
	contents := []string{
		sc.View(sw, panelH),
		tm.View(tw, panelH),
	}
	return Row{Cols: cols, Contents: contents, Height: panelH}
}

// buildFooterRow produces a single-column footer Row.
func buildFooterRow(totalW int, theme Theme) Row {
	w := totalW - 2
	if w < 1 {
		w = 1
	}
	hint := "←→ module · : terminal · esc back · p install · P push · d doctor · x uninstall · q quit"
	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.FooterFg())).
		Width(w).
		Render(hint)
	return Row{Cols: []int{w}, Contents: []string{content}, Height: 1}
}
