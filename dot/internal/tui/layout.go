package tui

// Layout holds computed dimensions for all 6 dashboard sections.
// Every section's content widths and the vertical budget are calculated here.
// The rendering layer reads Layout and never computes sizes itself.
type Layout struct {
	TotalW, TotalH int

	// --- Channel Strip ---
	ChipW, ChipH int   // per-chip content size (fixed 8×3)
	ChipCols     []int // content widths for each visible column (last absorbs remainder)

	// --- Middle Panels (1/4 : 2/4 : 1/4) ---
	PanelCols []int // content widths: [overview, scope, terminal] or [scope, terminal]
	PanelH    int   // content height (self-adaptive)

	// --- Controls (4 equal buttons) ---
	CtrlCols []int // content widths for each button

	// --- Footer (single column) ---
	FooterW int // content width (totalW - 2 borders)

	// --- Degradation ---
	ShowFooter   bool
	ShowOverview bool
	ShowControls bool
}

// ComputeLayout calculates all region sizes from terminal dimensions.
func ComputeLayout(totalW, totalH, moduleCount int) Layout {
	lay := Layout{TotalW: totalW, TotalH: totalH}

	const chipW = 8
	const chipH = 3
	lay.ChipW = chipW
	lay.ChipH = chipH

	// --- Degradation flags ---
	lay.ShowFooter = totalH >= 24
	lay.ShowOverview = totalW >= 100
	lay.ShowControls = totalH >= 20

	// --- Vertical budget ---
	// title(1) + top(1) + chips(chipH) + div(1) + panels(?) + div(1) + controls(1) + div(1) + footer(1) + bottom(1)
	fixed := 1 + 1 + chipH + 1 + 1 // title + top border + chips + chip-panel divider + bottom border
	if lay.ShowControls {
		fixed += 1 + 1 // divider + controls row
	}
	if lay.ShowFooter {
		fixed += 1 + 1 // divider + footer row
	}
	// If neither controls nor footer, the panels row ends with bottom border (already counted)
	// If controls but no footer, bottom border is after controls
	// If footer, bottom border is after footer
	lay.PanelH = totalH - fixed
	if lay.PanelH < 1 {
		lay.PanelH = 1
	}

	// --- Channel Strip columns ---
	// Inner width = totalW - 2 (left + right border)
	innerW := totalW - 2
	if innerW < 1 {
		innerW = 1
	}
	// How many chips fit: each chip needs chipW, plus 1 separator between chips
	// N chips need: N*chipW + (N-1) separators = N*(chipW+1) - 1
	maxChips := (innerW + 1) / (chipW + 1)
	if maxChips < 1 {
		maxChips = 1
	}
	totalChips := moduleCount + 1 // modules + ADD
	visCols := totalChips
	if visCols > maxChips {
		visCols = maxChips
	}

	lay.ChipCols = make([]int, visCols)
	used := 0
	for i := 0; i < visCols; i++ {
		lay.ChipCols[i] = chipW
		used += chipW
		if i > 0 {
			used++ // separator
		}
	}
	// Last column absorbs remainder to fill totalW
	remainder := innerW - used
	if remainder > 0 && visCols > 0 {
		lay.ChipCols[visCols-1] += remainder
	}

	// --- Middle Panel columns ---
	if lay.ShowOverview {
		// 3 columns: inner = totalW - 4 borders (│col│col│col│)
		pw := totalW - 4
		if pw < 3 {
			pw = 3
		}
		ow := pw / 4
		tw := pw / 4
		sw := pw - ow - tw
		lay.PanelCols = []int{ow, sw, tw}
	} else {
		// 2 columns: inner = totalW - 3 borders (│col│col│)
		pw := totalW - 3
		if pw < 2 {
			pw = 2
		}
		sw := pw / 2
		tw := pw - sw
		lay.PanelCols = []int{sw, tw}
	}

	// --- Controls columns (4 buttons) ---
	// inner = totalW - 2 borders - 3 internal separators = totalW - 5
	cw := totalW - 5
	if cw < 4 {
		cw = 4
	}
	bw := cw / 4
	cr := cw - bw*4
	lay.CtrlCols = make([]int, 4)
	for i := 0; i < 4; i++ {
		lay.CtrlCols[i] = bw
		if i < cr {
			lay.CtrlCols[i]++
		}
	}

	// --- Footer ---
	lay.FooterW = totalW - 2
	if lay.FooterW < 1 {
		lay.FooterW = 1
	}

	return lay
}
