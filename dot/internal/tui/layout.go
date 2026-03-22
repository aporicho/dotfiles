package tui

// Region describes a rectangular content area on screen.
type Region struct {
	X, Y, W, H int
}

// Layout holds the computed regions for all dashboard sections.
type Layout struct {
	Width, Height int

	// Channel strip: the framed chip area at top
	ChipW, ChipH int // per-chip content size
	ChipCount    int  // number of columns in channel frame
	ChipsPerRow  int  // number of chips visible at once

	// Middle panel content regions (inside the frame)
	Overview Region
	Scope    Region
	Terminal Region

	// Controls and footer (borderless)
	Controls Region
	Footer   Region

	// Heights for frame rendering
	ChannelFrameH int // total channel frame height (chipH + 2 border lines)
	PanelFrameH   int // total panel frame height (innerH + 2 border lines)
	PanelInnerH   int // panel content height

	// Degradation flags based on terminal size
	ShowFooter   bool
	ShowOverview bool
	ShowControls bool
}

// ComputeLayout calculates all region sizes from terminal dimensions.
func ComputeLayout(totalW, totalH, moduleCount int) Layout {
	lay := Layout{Width: totalW, Height: totalH}

	// Channel strip: N chips (modules + ADD button)
	n := moduleCount + 1

	// Fixed chip dimensions
	lay.ChipW = 8
	lay.ChipH = 4
	lay.ChannelFrameH = lay.ChipH + 2 // content + top/bottom border

	// Calculate visible chips per row
	chipsPerRow := (totalW - 1) / (lay.ChipW + 1)
	if chipsPerRow < 1 {
		chipsPerRow = 1
	}
	lay.ChipsPerRow = chipsPerRow
	lay.ChipCount = n // keep total count

	// Degradation flags based on terminal size
	lay.ShowFooter = totalH >= 24
	lay.ShowOverview = totalW >= 100
	lay.ShowControls = totalH >= 20

	// Vertical budget: channelFrame + panelFrame + controls(1) + sep(1) + footer(1)
	// Adjust budget for hidden panels
	fixedLines := 1 // sep is always shown
	if lay.ShowFooter {
		fixedLines++
	}
	if lay.ShowControls {
		fixedLines++
	}
	panelBudget := totalH - lay.ChannelFrameH - fixedLines
	lay.PanelInnerH = panelBudget - 2 // subtract panel frame top/bottom borders
	if lay.PanelInnerH < 1 {
		lay.PanelInnerH = 1
	}
	lay.PanelFrameH = lay.PanelInnerH + 2

	// Horizontal layout
	if lay.ShowOverview {
		// 3 columns: panel content width = total - 4 border chars (│ + │ + │ + │)
		contentW := totalW - 4
		if contentW < 3 {
			contentW = 3
		}
		ow := contentW / 4
		tw := contentW / 4
		sw := contentW - ow - tw

		lay.Overview = Region{X: 1, W: ow, H: lay.PanelInnerH}
		lay.Scope = Region{X: 2 + ow, W: sw, H: lay.PanelInnerH}
		lay.Terminal = Region{X: 3 + ow + sw, W: tw, H: lay.PanelInnerH}
	} else {
		// 2 columns (scope + terminal): panel content width = total - 3 border chars (│ + │ + │)
		contentW := totalW - 3
		if contentW < 2 {
			contentW = 2
		}
		half := contentW / 2
		sw := half
		tw := contentW - half

		lay.Scope = Region{X: 1, W: sw, H: lay.PanelInnerH}
		lay.Terminal = Region{X: 2 + sw, W: tw, H: lay.PanelInnerH}
	}

	lay.Controls = Region{W: totalW, H: 1}
	lay.Footer = Region{W: totalW, H: 1}

	return lay
}
