package tui

// Layout holds the vertical budget and degradation flags.
// It does NOT contain any panel's column structure — each panel owns that.
type Layout struct {
	TotalW, TotalH int
	PanelH         int  // content height for the middle panels (self-adaptive)
	ShowOverview   bool // false when terminal width < 100
	ShowControls   bool // false when terminal height < 20
	ShowFooter     bool // false when terminal height < 24
}

const (
	chipW = 8 // exported within package as fixed chip content width
	chipH = 3 // exported within package as fixed chip content height
)

// ComputeLayout calculates the vertical budget and degradation flags.
func ComputeLayout(totalW, totalH int) Layout {
	lay := Layout{TotalW: totalW, TotalH: totalH}

	// Degradation flags
	lay.ShowFooter = totalH >= 24
	lay.ShowOverview = totalW >= 100
	lay.ShowControls = totalH >= 20

	// Vertical budget:
	//   title(1) + top(1) + chips(chipH) + divider(1)
	//   + panels(?) + bottom(1)
	//   + [divider(1) + controls(1)]  if ShowControls
	//   + [divider(1) + footer(1)]    if ShowFooter
	fixed := 1 + 1 + chipH + 1 + 1 // title + top + chips + divider + bottom
	if lay.ShowControls {
		fixed += 2 // divider + controls
	}
	if lay.ShowFooter {
		fixed += 2 // divider + footer
	}
	lay.PanelH = totalH - fixed
	if lay.PanelH < 1 {
		lay.PanelH = 1
	}

	return lay
}
