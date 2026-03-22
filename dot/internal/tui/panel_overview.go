package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
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
	dfPath     string
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
	dfPath string,
) *Overview {
	return &Overview{
		modules:    modules,
		manifest:   mf,
		gitChanges: gitChanges,
		dfPath:     dfPath,
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
	lines := []string{
		o.renderHeader(width),
		"",
		o.renderHealthBar(width),
		"",
		o.renderStats(width),
		"",
		o.renderSystemInfo(width),
	}

	content := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1).
		Render(content)
}

// renderHeader renders the "\uf054 OVERVIEW" title line.
func (o *Overview) renderHeader(width int) string {
	_ = width
	return o.styles.Header.Render("\uf054 OVERVIEW")
}

// renderHealthBar renders a progress bar showing healthy links+secrets ratio.
func (o *Overview) renderHealthBar(width int) string {
	total, healthy := o.healthCounts()

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

// renderSystemInfo renders git branch, sync status, platform, shell, pkg manager, keychain, and font.
func (o *Overview) renderSystemInfo(width int) string {
	_ = width

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Dimmed()))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Cyan()))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Green()))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color(o.theme.Yellow()))

	// Git branch (static "main" as per spec).
	branch := cyan.Render("main")
	branchLine := dimmed.Render("branch ") + branch

	// Sync status: ahead/behind upstream, or local changes count.
	ahead, behind, _ := gitpkg.AheadBehind(o.dfPath)
	var syncStatus string
	switch {
	case ahead == 0 && behind == 0 && len(o.gitChanges) == 0:
		syncStatus = green.Render("clean")
	case behind > 0 && ahead > 0:
		syncStatus = yellow.Render(fmt.Sprintf("%d behind · %d ahead", behind, ahead))
	case behind > 0:
		syncStatus = yellow.Render(fmt.Sprintf("%d behind", behind))
	case ahead > 0:
		syncStatus = yellow.Render(fmt.Sprintf("%d ahead", ahead))
	default:
		syncStatus = yellow.Render(fmt.Sprintf("%d change(s)", len(o.gitChanges)))
	}
	syncLine := dimmed.Render("sync   ") + syncStatus

	// Platform.
	plat := platform.Current()
	platformLine := dimmed.Render("os     ") + cyan.Render(plat)

	// Shell.
	shellPath := os.Getenv("SHELL")
	shellName := filepath.Base(shellPath)
	if shellName == "" || shellName == "." {
		shellName = "unknown"
	}
	shellLine := dimmed.Render("shell  ") + cyan.Render(shellName)

	// Package manager.
	pkgMgr := platform.PackageManager()
	if pkgMgr == "" {
		pkgMgr = "none"
	}
	pkgLine := dimmed.Render("pkg    ") + cyan.Render(pkgMgr)

	// Keychain status.
	var keychainStatus string
	if secrets.KeychainAvailable() {
		keychainStatus = green.Render("available")
	} else {
		keychainStatus = yellow.Render("unavailable")
	}
	keychainLine := dimmed.Render("key    ") + keychainStatus

	// Font.
	fontLine := dimmed.Render("font   ") + cyan.Render("JetBrainsMono NF")

	return strings.Join([]string{branchLine, syncLine, platformLine, shellLine, pkgLine, keychainLine, fontLine}, "\n")
}

// healthCounts returns the total and healthy counts across all module links and secrets.
func (o *Overview) healthCounts() (total, healthy int) {
	for _, mod := range o.modules {
		for range mod.Links {
			total++
			if o.manifest.IsInstalled(mod.Name) {
				healthy++
			}
		}
		for _, sec := range mod.Secrets {
			total++
			if o.manifest.IsInstalled(mod.Name) {
				plainPath := filepath.Join(mod.Dir, sec.Source)
				info, err := os.Stat(plainPath)
				if err == nil && info.Mode().Perm() == 0o600 {
					healthy++
				}
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
