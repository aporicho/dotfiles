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

	var rows []Row

	// Row 1: Channel Strip
	chipRow := d.channel.BuildRow(lay.TotalW)
	rows = append(rows, chipRow)

	// Row 2: Middle panels (uses Weight for column ratios)
	var middlePanels []Panel
	if lay.ShowOverview {
		middlePanels = []Panel{d.overview, d.scope, d.terminal}
	} else {
		middlePanels = []Panel{d.scope, d.terminal}
	}
	panelRow := buildPanelRow(lay.TotalW, lay.PanelH, middlePanels)
	rows = append(rows, panelRow)

	// Row 3: Controls
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

// buildPanelRow produces the middle-panel Row using Panel.Weight(n) for column ratios.
func buildPanelRow(totalW, panelH int, panels []Panel) Row {
	n := len(panels)
	pw := totalW - (n + 1) // border chars
	if pw < n {
		pw = n
	}

	totalWeight := 0
	for _, p := range panels {
		totalWeight += p.Weight(n)
	}
	if totalWeight == 0 {
		totalWeight = n
	}

	cols := make([]int, n)
	used := 0
	for i, p := range panels {
		cols[i] = pw * p.Weight(n) / totalWeight
		used += cols[i]
	}
	cols[n/2] += pw - used

	contents := make([]string, n)
	for i, p := range panels {
		contents[i] = p.View(cols[i], panelH)
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
