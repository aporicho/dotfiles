package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// borderSet holds the box-drawing characters used by the frame renderer.
type borderSet struct {
	H, V                         string // ─ │
	TL, TR, BL, BR               string // ┌ ┐ └ ┘
	LT, RT                       string // ├ ┤
	TDown, TUp, Cross            string // ┬ ┴ ┼
}

func defaultBorder() borderSet {
	b := lipgloss.NormalBorder()
	return borderSet{
		H: b.Top, V: b.Left,
		TL: b.TopLeft, TR: b.TopRight, BL: b.BottomLeft, BR: b.BottomRight,
		LT: b.MiddleLeft, RT: b.MiddleRight,
		TDown: b.MiddleTop, TUp: b.MiddleBottom, Cross: b.Middle,
	}
}

// sepPositions returns the character positions (0-based within innerW) where
// vertical separators sit, given column content widths.
// For cols [8, 8, 8], inner positions are: 8, 17 (after each col + its width).
func sepPositions(cols []int) []int {
	if len(cols) <= 1 {
		return nil
	}
	pos := make([]int, 0, len(cols)-1)
	x := 0
	for i := 0; i < len(cols)-1; i++ {
		x += cols[i]
		pos = append(pos, x)
		x++ // the separator itself occupies 1 char
	}
	return pos
}

// buildHLine builds a horizontal border/divider line of exactly totalW characters.
// left/right are the endpoint chars (e.g. ┌/┐ or ├/┤ or └/┘).
// aboveSeps and belowSeps are separator positions from the rows above and below.
// For a top edge, aboveSeps is nil. For a bottom edge, belowSeps is nil.
func buildHLine(b borderSet, sty lipgloss.Style, totalW int, left, right string, aboveSeps, belowSeps []int) string {
	// Build a set of positions for quick lookup
	aboveSet := make(map[int]bool, len(aboveSeps))
	for _, p := range aboveSeps {
		aboveSet[p] = true
	}
	belowSet := make(map[int]bool, len(belowSeps))
	for _, p := range belowSeps {
		belowSet[p] = true
	}

	var buf strings.Builder
	buf.WriteString(left)

	innerW := totalW - 2 // minus left + right endpoints
	for i := 0; i < innerW; i++ {
		above := aboveSet[i]
		below := belowSet[i]
		switch {
		case above && below:
			buf.WriteString(b.Cross)
		case above:
			buf.WriteString(b.TUp)
		case below:
			buf.WriteString(b.TDown)
		default:
			buf.WriteString(b.H)
		}
	}

	buf.WriteString(right)
	return sty.Render(buf.String())
}

// buildContentRow builds a single content row with │ separators at the given positions.
// contentLines[col][row] is the pre-split line for each column.
// cols are the content widths.
func buildContentRow(b borderSet, sty lipgloss.Style, cols []int, contentLines [][]string, row int) string {
	var buf strings.Builder
	buf.WriteString(sty.Render(b.V))

	for i, w := range cols {
		if i > 0 {
			buf.WriteString(sty.Render(b.V))
		}
		line := ""
		if row < len(contentLines[i]) {
			line = contentLines[i][row]
		}
		// Pad to column width (in case content is short)
		_ = w // width enforcement is done by the panel's View method
		buf.WriteString(line)
	}

	buf.WriteString(sty.Render(b.V))
	return buf.String()
}

// Row describes one horizontal band of the dashboard frame.
type Row struct {
	Cols     []int      // content widths for each column
	Contents []string   // pre-rendered content blocks (one per column)
	Height   int        // number of content lines
}

// RenderUnifiedFrame renders the complete dashboard as a single box-drawing frame.
// title is rendered above the top border.
// rows are rendered top-to-bottom with divider lines between them.
// The frame is exactly totalW characters wide.
func RenderUnifiedFrame(sty lipgloss.Style, title string, totalW int, rows []Row) string {
	b := defaultBorder()

	// Split all content into lines
	type splitRow struct {
		cols  []int
		lines [][]string
		seps  []int
	}
	split := make([]splitRow, len(rows))
	for i, r := range rows {
		lines := make([][]string, len(r.Contents))
		for j, c := range r.Contents {
			lines[j] = strings.Split(c, "\n")
		}
		split[i] = splitRow{
			cols:  r.Cols,
			lines: lines,
			seps:  sepPositions(r.Cols),
		}
	}

	var out []string

	// Title
	if title != "" {
		out = append(out, title)
	}

	for i, sr := range split {
		// Horizontal line above this row
		if i == 0 {
			// Top border: ┌───┬───┐
			out = append(out, buildHLine(b, sty, totalW, b.TL, b.TR, nil, sr.seps))
		} else {
			// Divider: ├───┼───┤ (connecting previous row's seps to this row's seps)
			out = append(out, buildHLine(b, sty, totalW, b.LT, b.RT, split[i-1].seps, sr.seps))
		}

		// Content rows
		for r := 0; r < rows[i].Height; r++ {
			out = append(out, buildContentRow(b, sty, sr.cols, sr.lines, r))
		}
	}

	// Bottom border: └───┴───┘
	lastSeps := split[len(split)-1].seps
	out = append(out, buildHLine(b, sty, totalW, b.BL, b.BR, lastSeps, nil))

	return strings.Join(out, "\n")
}

// RenderFrame draws a box-drawing border around N columns of content.
// Kept for backward compatibility where a standalone frame is needed.
func RenderFrame(
	border lipgloss.Border,
	borderSty lipgloss.Style,
	colWidths []int,
	contents []string,
	innerHeight int,
) string {
	row := Row{Cols: colWidths, Contents: contents, Height: innerHeight}
	return RenderUnifiedFrame(borderSty, "", sumCols(colWidths)+len(colWidths)+1, []Row{row})
}

func sumCols(cols []int) int {
	s := 0
	for _, c := range cols {
		s += c
	}
	return s
}
