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

	// --- Controls columns (4 buttons, aligned to panel separators) ---
	// Button separators must align with panel separators so ┼ connectors work.
	// With 3 panels [ow, sw, tw]: btn1=ow, btn2+btn3 split sw, btn4=tw.
	// With 2 panels [sw, tw]: btn1+btn2 split sw, btn3+btn4 split tw.
	if lay.ShowOverview && len(lay.PanelCols) == 3 {
		ow := lay.PanelCols[0]
		sw := lay.PanelCols[1]
		tw := lay.PanelCols[2]
		// Split Scope column into two halves for buttons 2 and 3
		// sw contains the content width; splitting it needs to account for
		// the extra separator: btn2 + 1(sep) + btn3 = sw
		left := (sw - 1) / 2
		right := sw - 1 - left
		lay.CtrlCols = []int{ow, left, right, tw}
	} else if len(lay.PanelCols) == 2 {
		sw := lay.PanelCols[0]
		tw := lay.PanelCols[1]
		left1 := (sw - 1) / 2
		right1 := sw - 1 - left1
		left2 := (tw - 1) / 2
		right2 := tw - 1 - left2
		lay.CtrlCols = []int{left1, right1, left2, right2}
	} else {
		// Fallback: equal split
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
	}

	// --- Footer ---
	lay.FooterW = totalW - 2
	if lay.FooterW < 1 {
		lay.FooterW = 1
	}

	return lay
}
