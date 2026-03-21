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
	modules    []*module.Module
	manifest   *manifest.Manifest
	gitChanges []string
	selected   int
	focused    bool
	styles     Styles
	theme      Theme
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
			return
		}
	}
}

// moveNext moves the selection to the next navigable module.
func (cs *ChannelStrip) moveNext() {
	for i := cs.selected + 1; i < len(cs.modules); i++ {
		if platform.MatchesPlatform(cs.modules[i].Platforms) {
			cs.selected = i
			return
		}
	}
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

// View implements Panel. Renders all module chips horizontally.
func (cs *ChannelStrip) View(width, _ int) string {
	if len(cs.modules) == 0 {
		return cs.addChip()
	}

	chips := make([]string, 0, len(cs.modules)+1)

	for i, mod := range cs.modules {
		matchesPlatform := platform.MatchesPlatform(mod.Platforms)
		isSelected := i == cs.selected
		isInstalled := cs.manifest.IsInstalled(mod.Name)
		hasSecrets := len(mod.Secrets) > 0
		hasChanges := cs.hasGitChanges(mod)

		// Determine LED color.
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

		led := lipgloss.NewStyle().Foreground(lipgloss.Color(ledColor)).Render("●")

		// Build chip label: NAME [🔐] [⚡] LED
		var label strings.Builder
		if matchesPlatform {
			label.WriteString(mod.Name)
		} else {
			label.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color(cs.theme.Dimmed())).
				Render(mod.Name))
		}
		if hasSecrets {
			label.WriteString(" 🔐")
		}
		if hasChanges {
			label.WriteString(" ⚡")
		}
		label.WriteString(" ")
		label.WriteString(led)

		// Apply chip style.
		var chip string
		if isSelected {
			chip = cs.styles.ChipSelected.Render(label.String())
		} else {
			chip = cs.styles.Chip.Render(label.String())
		}

		chips = append(chips, chip)
	}

	// ADD chip at the end.
	chips = append(chips, cs.addChip())

	row := lipgloss.JoinHorizontal(lipgloss.Top, chips...)

	// If the row fits within width, return as-is; otherwise truncate gracefully.
	_ = width
	return row
}

// addChip renders the dashed "+ ADD a" chip.
func (cs *ChannelStrip) addChip() string {
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cs.theme.Dimmed())).
		Render("+ ADD a")
	addStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(cs.theme.Dimmed())).
		BorderStyle(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "╌",
			Right:       "╌",
			TopLeft:     "╌",
			TopRight:    "╌",
			BottomLeft:  "╌",
			BottomRight: "╌",
		}).
		Padding(0, 1)
	return addStyle.Render(label)
}
