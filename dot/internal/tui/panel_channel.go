package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

// ChannelStrip is the top bar showing all modules as chips with LED status indicators.
type ChannelStrip struct {
	modules      []*module.Module
	manifest     *manifest.Manifest
	gitChanges   []string
	selected     int
	focused      bool
	styles       Styles
	theme        Theme
	scrollOffset int
	visibleChips int
}

// NewChannelStrip constructs a ChannelStrip with the given modules and state.
func NewChannelStrip(
	modules []*module.Module,
	mf *manifest.Manifest,
	gitChanges []string,
	styles Styles,
	theme Theme,
) *ChannelStrip {
	cs := &ChannelStrip{
		modules:    modules,
		manifest:   mf,
		gitChanges: gitChanges,
		styles:     styles,
		theme:      theme,
	}
	// Ensure selected starts on a navigable module.
	cs.selected = cs.firstNavigable()
	return cs
}

// firstNavigable returns the index of the first platform-matching module, or 0.
func (cs *ChannelStrip) firstNavigable() int {
	for i, m := range cs.modules {
		if platform.MatchesPlatform(m.Platforms) {
			return i
		}
	}
	return 0
}

// Focused implements Panel.
func (cs *ChannelStrip) Focused() bool { return cs.focused }

// SetFocus implements Panel.
func (cs *ChannelStrip) SetFocus(f bool) { cs.focused = f }

// Selected returns the currently selected module (may be nil if none exist).
func (cs *ChannelStrip) Selected() *module.Module {
	if len(cs.modules) == 0 {
		return nil
	}
	return cs.modules[cs.selected]
}

// SelectedIndex returns the index of the currently selected module.
func (cs *ChannelStrip) SelectedIndex() int { return cs.selected }

// Update implements Panel. Handles ←→ navigation when focused.
func (cs *ChannelStrip) Update(msg tea.Msg) (Panel, tea.Cmd) {
	if !cs.focused {
		return cs, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return cs, nil
	}

	switch keyMsg.String() {
	case "left", "h":
		cs.movePrev()
		return cs, cs.emitSelected()
	case "right", "l":
		cs.moveNext()
		return cs, cs.emitSelected()
	}

	return cs, nil
}

// movePrev moves the selection to the previous navigable module.
func (cs *ChannelStrip) movePrev() {
	for i := cs.selected - 1; i >= 0; i-- {
		if platform.MatchesPlatform(cs.modules[i].Platforms) {
			cs.selected = i
			cs.ensureVisible()
			return
		}
	}
}

// moveNext moves the selection to the next navigable module.
func (cs *ChannelStrip) moveNext() {
	for i := cs.selected + 1; i < len(cs.modules); i++ {
		if platform.MatchesPlatform(cs.modules[i].Platforms) {
			cs.selected = i
			cs.ensureVisible()
			return
		}
	}
}

// SetVisibleChips sets how many chips are visible at once.
func (cs *ChannelStrip) SetVisibleChips(n int) {
	cs.visibleChips = n
}

// ensureVisible adjusts scrollOffset so the selected chip is within the visible window.
func (cs *ChannelStrip) ensureVisible() {
	if cs.visibleChips <= 0 {
		return
	}
	if cs.selected < cs.scrollOffset {
		cs.scrollOffset = cs.selected
	}
	if cs.selected >= cs.scrollOffset+cs.visibleChips {
		cs.scrollOffset = cs.selected - cs.visibleChips + 1
	}
}

// ScrollOffset returns the current horizontal scroll offset.
func (cs *ChannelStrip) ScrollOffset() int { return cs.scrollOffset }

// CanScrollLeft reports whether there are chips scrolled off to the left.
func (cs *ChannelStrip) CanScrollLeft() bool { return cs.scrollOffset > 0 }

// CanScrollRight reports whether there are chips scrolled off to the right.
func (cs *ChannelStrip) CanScrollRight() bool {
	return cs.scrollOffset+cs.visibleChips < len(cs.modules)
}

// emitSelected returns a Cmd that sends a ModuleSelectedMsg for the current selection.
func (cs *ChannelStrip) emitSelected() tea.Cmd {
	if len(cs.modules) == 0 {
		return nil
	}
	idx := cs.selected
	mod := cs.modules[idx]
	return func() tea.Msg {
		return ModuleSelectedMsg{Index: idx, Module: mod}
	}
}

// hasGitChanges reports whether the module directory has tracked git changes.
func (cs *ChannelStrip) hasGitChanges(mod *module.Module) bool {
	for _, path := range cs.gitChanges {
		if strings.HasPrefix(path, mod.Dir) || strings.Contains(path, mod.Name) {
			return true
		}
	}
	return false
}

// View implements Panel. Not used directly — dashboard calls ChipContents + RenderFrame.
func (cs *ChannelStrip) View(width, height int) string {
	return "" // Rendering handled by dashboard_view via ChipContents + RenderFrame
}

// ChipContents returns pure content blocks for the visible chips (modules + ADD).
// colWidths provides the content width for each visible column (last may be wider
// to absorb the remainder and fill totalW). chipH is fixed at 3.
func (cs *ChannelStrip) ChipContents(colWidths []int) []string {
	const chipH = 3
	visibleCount := len(colWidths)

	// Build all chips at standard width first
	const stdW = 8
	stdCenter := lipgloss.NewStyle().Width(stdW).Align(lipgloss.Center)
	all := make([]string, 0, len(cs.modules))

	for i, mod := range cs.modules {
		all = append(all, cs.renderChip(mod, i, stdW, chipH, stdCenter))
	}

	// Slice to the visible window
	start := cs.scrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(all) {
		start = len(all)
	}
	end := start + visibleCount
	if end > len(all) {
		end = len(all)
	}
	visible := all[start:end]

	// Re-render chips that need a different width (e.g. last column absorbs remainder)
	for i, cw := range colWidths {
		if i >= len(visible) {
			break
		}
		if cw != stdW {
			idx := start + i
			if idx < len(cs.modules) {
				center := lipgloss.NewStyle().Width(cw).Align(lipgloss.Center)
				visible[i] = cs.renderChip(cs.modules[idx], idx, cw, chipH, center)
			}
		}
	}

	// Pad if fewer chips than columns (empty cells)
	for len(visible) < len(colWidths) {
		w := colWidths[len(visible)]
		visible = append(visible, lipgloss.NewStyle().Width(w).Height(chipH).Render(""))
	}

	return visible
}

// renderChip renders the pure content for a single module chip.
func (cs *ChannelStrip) renderChip(mod *module.Module, idx, chipW, chipH int, centerStyle lipgloss.Style) string {
	matchesPlatform := platform.MatchesPlatform(mod.Platforms)
	isSelected := idx == cs.selected
	isInstalled := cs.manifest.IsInstalled(mod.Name)
	hasSecrets := len(mod.Secrets) > 0
	hasChanges := cs.hasGitChanges(mod)

	// LED color
	var ledColor string
	switch {
	case !matchesPlatform:
		ledColor = cs.theme.Dimmed()
	case !isInstalled:
		ledColor = cs.theme.LedError()
	case hasChanges:
		ledColor = cs.theme.LedWarning()
	default:
		ledColor = cs.theme.LedHealthy()
	}
	if isSelected && matchesPlatform {
		switch {
		case !isInstalled:
			ledColor = cs.theme.Red()
		case hasChanges:
			ledColor = cs.theme.Yellow()
		default:
			ledColor = cs.theme.Green()
		}
	}

	led := lipgloss.NewStyle().Foreground(lipgloss.Color(ledColor)).Render("\uf111")

	// Name line
	name := strings.ToUpper(mod.Name)
	if len(name) > chipW {
		name = name[:chipW]
	}
	var nameLine string
	switch {
	case !matchesPlatform:
		nameLine = lipgloss.NewStyle().Foreground(lipgloss.Color(cs.theme.Dimmed())).Render(name)
	case isSelected:
		nameLine = lipgloss.NewStyle().Foreground(lipgloss.Color(cs.theme.Blue())).Bold(true).Render(name)
	default:
		nameLine = name
	}

	// Indicator line
	var indicators strings.Builder
	indicators.WriteString(led)
	if hasSecrets {
		indicators.WriteString(" \uf023")
	}
	if hasChanges {
		indicators.WriteString(" \ue725")
	}

	inner := centerStyle.Render(nameLine) + "\n" + centerStyle.Render(indicators.String())

	return lipgloss.NewStyle().
		Width(chipW).Height(chipH).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(inner)
}

// ChipCount returns the total number of chips.
func (cs *ChannelStrip) ChipCount() int {
	return len(cs.modules)
}
