package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

// Overview is the global status panel on the left side.
// It does NOT change with module selection.
type Overview struct {
	modules    []*module.Module
	manifest   *manifest.Manifest
	gitChanges []string
	focused    bool
	styles     Styles
	theme      Theme
}

// NewOverview constructs an Overview panel.
func NewOverview(
	modules []*module.Module,
	mf *manifest.Manifest,
	gitChanges []string,
	styles Styles,
	theme Theme,
) *Overview {
	return &Overview{
		modules:    modules,
		manifest:   mf,
		gitChanges: gitChanges,
		styles:     styles,
		theme:      theme,
	}
}

// Focused implements Panel.
func (o *Overview) Focused() bool { return o.focused }

// SetFocus implements Panel.
func (o *Overview) SetFocus(f bool) { o.focused = f }

// Update implements Panel. Only handles DataReloadMsg to refresh data.
func (o *Overview) Update(msg tea.Msg) (Panel, tea.Cmd) {
	if reload, ok := msg.(DataReloadMsg); ok {
		o.modules = reload.Modules
		o.manifest = reload.Manifest
		o.gitChanges = reload.GitChanges
	}
	return o, nil
}

// View implements Panel. Renders the overview content within the given dimensions.
func (o *Overview) View(width, height int) string {
	inner := width - 2 // subtract border columns
	if inner < 1 {
		inner = 1
	}

	lines := []string{
		o.renderHeader(inner),
		"",
		o.renderHealthBar(inner),
		"",
		o.renderStats(inner),
		"",
		o.renderSystemInfo(inner),
	}

	content := strings.Join(lines, "\n")

	var borderStyle lipgloss.Style
	if o.focused {
		borderStyle = o.styles.PanelFocused
	} else {
		borderStyle = o.styles.Panel
	}

	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}
	return borderStyle.Width(inner).Height(innerHeight).Render(content)
}

// renderHeader renders the "▸ OVERVIEW" title line.
func (o *Overview) renderHeader(width int) string {
	_ = width
	return o.styles.Header.Render("▸ OVERVIEW")
}

// renderHealthBar renders a progress bar showing healthy links ratio.
func (o *Overview) renderHealthBar(width int) string {
	total, healthy := o.linkCounts()

	var pct int
	if total > 0 {
		pct = healthy * 100 / total
	} else {
		pct = 100
	}

	// Bar width: leave room for percentage label " 100%"
	labelWidth := 5 // " 100%"
	barWidth := width - labelWidth
	if barWidth < 1 {
		barWidth = 1
	}

	filled := barWidth * pct / 100
	empty := barWidth - filled

	var barColor string
	switch {
	case pct >= 80:
		barColor = o.theme.Green()
	case pct >= 50:
		barColor = o.theme.Yellow()
	default:
		barColor = o.theme.Red()
	}

	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).
		Render(strings.Repeat("█", filled))
	rest := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Dimmed())).
		Render(strings.Repeat("░", empty))

	label := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Subtle())).
		Render(fmt.Sprintf(" %3d%%", pct))

	return bar + rest + label
}

// renderStats renders a 2x2 grid of key counts.
func (o *Overview) renderStats(width int) string {
	modCount := len(o.modules)
	installedCount := o.countInstalled()
	secretsCount := o.countSecrets()
	changedCount := len(o.gitChanges)

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Dimmed()))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Subtle()))

	// Each cell: label on one line, number on next, side by side.
	// Layout two columns, each roughly half the width.
	colWidth := width / 2
	if colWidth < 6 {
		colWidth = 6
	}

	cell := func(label string, n int, color string) string {
		numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		top := dimmed.Render(label)
		bottom := numStyle.Render(fmt.Sprintf("%d", n))
		_ = value
		return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
	}

	c1 := cell("modules", modCount, o.theme.Blue())
	c2 := cell("installed", installedCount, o.theme.Green())
	c3 := cell("secrets", secretsCount, o.theme.Purple())
	c4 := cell("changed", changedCount, o.theme.Yellow())

	colStyle := lipgloss.NewStyle().Width(colWidth)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(c1), colStyle.Render(c2))
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(c3), colStyle.Render(c4))

	return lipgloss.JoinVertical(lipgloss.Left, row1, row2)
}

// renderSystemInfo renders git branch, sync status, platform, and keychain status.
func (o *Overview) renderSystemInfo(width int) string {
	_ = width

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Dimmed()))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Cyan()))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Green()))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Yellow()))

	// Git branch (static "main" as per spec).
	branch := cyan.Render("main")
	branchLine := dimmed.Render("branch ") + branch

	// Sync status: clean if no git changes.
	var syncStatus string
	if len(o.gitChanges) == 0 {
		syncStatus = green.Render("clean")
	} else {
		syncStatus = yellow.Render(fmt.Sprintf("%d change(s)", len(o.gitChanges)))
	}
	syncLine := dimmed.Render("sync   ") + syncStatus

	// Platform.
	plat := platform.Current()
	platformLine := dimmed.Render("os     ") + cyan.Render(plat)

	// Keychain status.
	var keychainStatus string
	if secrets.KeychainAvailable() {
		keychainStatus = green.Render("available")
	} else {
		keychainStatus = yellow.Render("unavailable")
	}
	keychainLine := dimmed.Render("key    ") + keychainStatus

	return strings.Join([]string{branchLine, syncLine, platformLine, keychainLine}, "\n")
}

// linkCounts returns the total and healthy link counts across all modules.
func (o *Overview) linkCounts() (total, healthy int) {
	for _, mod := range o.modules {
		for _, link := range mod.Links {
			_ = link
			total++
			if o.manifest.IsInstalled(mod.Name) {
				healthy++
			}
		}
	}
	return
}

// countInstalled returns the number of installed modules.
func (o *Overview) countInstalled() int {
	count := 0
	for _, mod := range o.modules {
		if o.manifest.IsInstalled(mod.Name) {
			count++
		}
	}
	return count
}

// countSecrets returns the total number of secrets across all modules.
func (o *Overview) countSecrets() int {
	count := 0
	for _, mod := range o.modules {
		count += len(mod.Secrets)
	}
	return count
}
