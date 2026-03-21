package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// Scope displays the selected module's details: health meters, signal path, and deps.
type Scope struct {
	module     *module.Module
	manifest   *manifest.Manifest
	gitChanges []string
	focused    bool
	styles     Styles
	theme      Theme
}

// NewScope constructs a Scope panel.
func NewScope(
	mod *module.Module,
	mf *manifest.Manifest,
	gitChanges []string,
	styles Styles,
	theme Theme,
) *Scope {
	return &Scope{
		module:     mod,
		manifest:   mf,
		gitChanges: gitChanges,
		styles:     styles,
		theme:      theme,
	}
}

// Focused implements Panel.
func (s *Scope) Focused() bool { return s.focused }

// SetFocus implements Panel.
func (s *Scope) SetFocus(f bool) { s.focused = f }

// Update implements Panel.
func (s *Scope) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch m := msg.(type) {
	case ModuleSelectedMsg:
		s.module = m.Module
	case DataReloadMsg:
		s.manifest = m.Manifest
		s.gitChanges = m.GitChanges
		// Update module reference if it matches the current one.
		if s.module != nil {
			for _, mod := range m.Modules {
				if mod.Name == s.module.Name {
					s.module = mod
					break
				}
			}
		}
	}
	return s, nil
}

// View implements Panel.
func (s *Scope) View(width, height int) string {
	inner := width - 2 // subtract border columns
	if inner < 1 {
		inner = 1
	}

	var lines []string

	if s.module == nil {
		lines = append(lines, s.styles.Dimmed.Render("no module selected"))
	} else {
		lines = append(lines, s.renderHeader())
		lines = append(lines, "")
		lines = append(lines, s.renderMeters(inner))
		lines = append(lines, "")
		lines = append(lines, s.renderSignalPath())
		deps := s.renderDeps()
		if deps != "" {
			lines = append(lines, "")
			lines = append(lines, deps)
		}
	}

	content := strings.Join(lines, "\n")

	var borderStyle lipgloss.Style
	if s.focused {
		borderStyle = s.styles.PanelFocused
	} else {
		borderStyle = s.styles.Panel
	}

	return borderStyle.Width(inner).Render(content)
}

// renderHeader renders "▸ SCOPE: MODNAME" plus description.
func (s *Scope) renderHeader() string {
	title := s.styles.Header.Render("▸ SCOPE: " + strings.ToUpper(s.module.Name))
	if s.module.Description != "" {
		desc := lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.theme.Dimmed())).
			Render(s.module.Description)
		return title + "\n" + desc
	}
	return title
}

// renderMeters renders the LNK / SEC / GIT horizontal progress bars inside a box.
func (s *Scope) renderMeters(width int) string {
	const barWidth = 20

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Dimmed()))

	lnkBar := s.renderLinkMeter(barWidth)
	lnkLine := dimmed.Render("LNK ") + lnkBar

	var meterLines []string
	meterLines = append(meterLines, lnkLine)

	if len(s.module.Secrets) > 0 {
		secBar := s.renderSecretMeter(barWidth)
		secLine := dimmed.Render("SEC ") + secBar
		meterLines = append(meterLines, secLine)
	}

	gitBar := s.renderGitMeter(barWidth)
	gitLine := dimmed.Render("GIT ") + gitBar
	meterLines = append(meterLines, gitLine)

	content := strings.Join(meterLines, "\n")

	boxWidth := width - 2
	if boxWidth < 4 {
		boxWidth = 4
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(s.theme.PanelBorder())).
		Padding(0, 1).
		Width(boxWidth).
		Render(content)

	return box
}

// renderLinkMeter renders the LNK progress bar.
func (s *Scope) renderLinkMeter(barWidth int) string {
	total := len(s.module.Links)
	healthy := 0
	for _, lnk := range s.module.Links {
		if s.isLinkHealthy(lnk) {
			healthy++
		}
	}

	var pct int
	if total > 0 {
		pct = healthy * 100 / total
	} else {
		pct = 100
	}

	var barColor string
	if healthy == total {
		barColor = s.theme.Green()
	} else {
		barColor = s.theme.Red()
	}

	bar := s.progressBar(barWidth, pct, barColor)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Dimmed())).
		Render(fmt.Sprintf(" %d/%d", healthy, total))
	return bar + label
}

// renderSecretMeter renders the SEC progress bar.
func (s *Scope) renderSecretMeter(barWidth int) string {
	allOK := true
	for _, sec := range s.module.Secrets {
		plainPath := filepath.Join(s.module.Dir, sec.Source)
		info, err := os.Stat(plainPath)
		if err != nil || info.Mode().Perm() != 0o600 {
			allOK = false
			break
		}
	}

	var barColor string
	var pct int
	if allOK {
		barColor = s.theme.Green()
		pct = 100
	} else {
		barColor = s.theme.Yellow()
		pct = 50
	}

	bar := s.progressBar(barWidth, pct, barColor)
	var statusLabel string
	if allOK {
		statusLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Green())).Render(" ok")
	} else {
		statusLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Yellow())).Render(" ?")
	}
	return bar + statusLabel
}

// renderGitMeter renders the GIT progress bar.
func (s *Scope) renderGitMeter(barWidth int) string {
	changes := s.moduleGitChanges()

	var barColor string
	var pct int
	var label string

	if len(changes) == 0 {
		barColor = s.theme.Green()
		pct = 100
		label = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Green())).Render(" clean")
	} else {
		barColor = s.theme.Yellow()
		pct = 50
		label = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Yellow())).
			Render(fmt.Sprintf(" %d change(s)", len(changes)))
	}

	bar := s.progressBar(barWidth, pct, barColor)
	return bar + label
}

// progressBar returns a filled progress bar string.
func (s *Scope) progressBar(width, pct int, color string) string {
	if width < 1 {
		width = 1
	}
	filled := width * pct / 100
	empty := width - filled

	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).
		Render(strings.Repeat("█", filled))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Dimmed())).
		Render(strings.Repeat("░", empty))
	return filledStr + emptyStr
}

// renderSignalPath renders the list of links with health indicators.
func (s *Scope) renderSignalPath() string {
	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Dimmed()))
	header := dimmed.Render("SIGNAL PATH")

	var rows []string
	rows = append(rows, header)

	secretTargets := s.secretTargetSet()

	for _, lnk := range s.module.Links {
		healthy := s.isLinkHealthy(lnk)
		var dot string
		if healthy {
			dot = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Green())).Render("●")
		} else {
			dot = lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Red())).Render("●")
		}

		target := expandHome(lnk.Target)
		label := target

		if secretTargets[target] {
			label = label + " 🔐"
		}

		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Subtle()))
		rows = append(rows, dot+" "+pathStyle.Render(label))
	}

	return strings.Join(rows, "\n")
}

// renderDeps renders brew/apt dependencies if any.
func (s *Scope) renderDeps() string {
	var deps []string

	switch runtime.GOOS {
	case "darwin":
		deps = s.module.Deps.Darwin.Brew
	case "linux":
		deps = s.module.Deps.Linux.Apt
	}

	if len(deps) == 0 {
		return ""
	}

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Dimmed()))
	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color(s.theme.Subtle()))

	var lines []string
	lines = append(lines, dimmed.Render("DEPS"))
	for _, d := range deps {
		lines = append(lines, subtle.Render("  "+d))
	}
	return strings.Join(lines, "\n")
}

// isLinkHealthy checks whether a module link's symlink is healthy.
func (s *Scope) isLinkHealthy(lnk module.Link) bool {
	target := expandHome(lnk.Target)

	info, err := os.Lstat(target)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}

	actual, err := os.Readlink(target)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(actual) {
		actual = filepath.Join(filepath.Dir(target), actual)
	}

	// Resolve expected source path.
	source := lnk.Source
	if !filepath.IsAbs(source) && s.module.Dir != "" {
		source = filepath.Join(s.module.Dir, source)
	}

	return filepath.Clean(actual) == filepath.Clean(source)
}

// secretTargetSet builds a set of expanded target paths that belong to secrets.
func (s *Scope) secretTargetSet() map[string]bool {
	set := make(map[string]bool)
	for _, sec := range s.module.Secrets {
		set[expandHome(sec.Target)] = true
	}
	return set
}

// moduleGitChanges returns git changes that belong to this module.
func (s *Scope) moduleGitChanges() []string {
	if s.module == nil {
		return nil
	}
	var changes []string
	for _, path := range s.gitChanges {
		if strings.HasPrefix(path, s.module.Dir) || strings.Contains(path, s.module.Name) {
			changes = append(changes, path)
		}
	}
	return changes
}

// expandHome replaces a leading "~" with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
