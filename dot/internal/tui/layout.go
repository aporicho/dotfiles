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
}

// ComputeLayout calculates all region sizes from terminal dimensions.
func ComputeLayout(totalW, totalH, moduleCount int) Layout {
	lay := Layout{Width: totalW, Height: totalH}

	// Channel strip: N chips (modules + ADD button)
	n := moduleCount + 1
	lay.ChipCount = n

	// Chip width: distribute total width among N chips + N+1 border chars
	lay.ChipW = (totalW - n - 1) / n
	if lay.ChipW < 6 {
		lay.ChipW = 6
	}

	// Chip height: W/2 for visual square (terminal chars are ~1:2 aspect)
	lay.ChipH = lay.ChipW / 2
	if lay.ChipH < 3 {
		lay.ChipH = 3
	}

	lay.ChannelFrameH = lay.ChipH + 2 // content + top/bottom border

	// Vertical budget: channelFrame + panelFrame + controls(1) + sep(1) + footer(1)
	panelBudget := totalH - lay.ChannelFrameH - 1 - 1 - 1
	lay.PanelInnerH = panelBudget - 2 // subtract panel frame top/bottom borders
	if lay.PanelInnerH < 1 {
		lay.PanelInnerH = 1
	}
	lay.PanelFrameH = lay.PanelInnerH + 2

	// Horizontal: panel content width = total - 4 border chars (│ + │ + │ + │)
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

	lay.Controls = Region{W: totalW, H: 1}
	lay.Footer = Region{W: totalW, H: 1}

	return lay
}
