# TUI Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a synthesizer-style multi-panel TUI dashboard that launches when running bare `dot` (no subcommands).

**Architecture:** Extract command core logic into `dot/internal/ops/` package as `io.Writer` functions (avoiding circular import with `cmd` ↔ `tui`). Create 5 panel sub-models (Channel Strip, Overview, Scope, Terminal, Controls), compose them in a Dashboard model, and wire to `rootCmd`. Tokyo Night themed with lipgloss.

**Tech Stack:** Go, bubbletea v1.3.10, lipgloss v1.1.0, cobra

**Spec:** `docs/superpowers/specs/2026-03-22-tui-dashboard-design.md`

---

### Task 1: Create Tokyo Night styles

**Files:**
- Create: `dot/internal/tui/styles.go`

- [ ] **Step 1: Create theme.go using ANSI 0-15 palette (terminal-driven, not hardcoded)**

Follow the pattern from markcli (`/Users/aporicho/Documents/GitHub/markcli/internal/theme/theme.go`): use ANSI color numbers ("0"-"15") so the terminal's color scheme drives all colors. This means the TUI automatically adapts to any terminal theme (Tokyo Night, Nord, Dracula, etc.) without hardcoded hex values.

Create `dot/internal/tui/theme.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

// Theme uses ANSI 0-15 colors so the terminal's palette drives all styling.
// Users control the look by configuring their terminal theme.
type Theme struct {
	Dark bool
}

// DetectTheme returns a Theme with Dark set based on the terminal background.
func DetectTheme() Theme {
	return Theme{Dark: lipgloss.HasDarkBackground()}
}

func (t Theme) contrastFg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

// --- Semantic colors (ANSI numbers) ---

func (t Theme) Green() string   { return "2" }
func (t Theme) Yellow() string  { return "3" }
func (t Theme) Blue() string    { return "4" }
func (t Theme) Purple() string  { return "5" }
func (t Theme) Cyan() string    { return "6" }
func (t Theme) Red() string     { return "1" }
func (t Theme) Dimmed() string  { return "8" }
func (t Theme) Subtle() string  { return "7" }

// --- Panel ---

func (t Theme) PanelBorder() string       { return "8" }
func (t Theme) PanelFocusBorder() string   { return "4" }
func (t Theme) PanelBg() string {
	if t.Dark { return "0" }
	return "15"
}

// --- Channel Strip ---

func (t Theme) ChipBorder() string        { return "8" }
func (t Theme) ChipSelectedBorder() string { return "4" }

// --- Status LEDs ---

func (t Theme) LedHealthy() string { return "2" }
func (t Theme) LedWarning() string { return "3" }
func (t Theme) LedError() string   { return "1" }

// --- Controls ---

func (t Theme) BtnPull() string   { return "4" }
func (t Theme) BtnPush() string   { return "5" }
func (t Theme) BtnDoctor() string { return "2" }
func (t Theme) BtnRemove() string { return "1" }

// --- Footer ---

func (t Theme) FooterFg() string { return "8" }
func (t Theme) FooterBg() string { return t.PanelBg() }
```

Then create `dot/internal/tui/styles.go` with lipgloss helpers that consume the Theme:

```go
package tui

import "github.com/charmbracelet/lipgloss"

// Styles derived from a Theme. Created once at dashboard startup.
type Styles struct {
	Panel        lipgloss.Style
	PanelFocused lipgloss.Style
	Header       lipgloss.Style
	Value        lipgloss.Style
	Green        lipgloss.Style
	Yellow       lipgloss.Style
	Red          lipgloss.Style
	Blue         lipgloss.Style
	Purple       lipgloss.Style
	Dimmed       lipgloss.Style
	Chip         lipgloss.Style
	ChipSelected lipgloss.Style
	Button       lipgloss.Style
}

// NewStyles creates a Styles set from a Theme.
func NewStyles(t Theme) Styles {
	return Styles{
		Panel: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(t.PanelBorder())).
			Border(lipgloss.RoundedBorder()),
		PanelFocused: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(t.PanelFocusBorder())).
			Border(lipgloss.RoundedBorder()),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Dimmed())).
			Uppercase(true),
		Value:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Subtle())),
		Green:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Green())),
		Yellow: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Yellow())),
		Red:    lipgloss.NewStyle().Foreground(lipgloss.Color(t.Red())),
		Blue:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Blue())),
		Purple: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Purple())),
		Dimmed: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dimmed())),
		Chip: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.ChipBorder())).
			Padding(0, 1),
		ChipSelected: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.ChipSelectedBorder())).
			Padding(0, 1),
		Button: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Align(lipgloss.Center),
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/theme.go dot/internal/tui/styles.go
git commit -m "feat(tui): add ANSI palette theme and styles"
```

---

### Task 2: Extract pull core logic into ops package

**Files:**
- Create: `dot/internal/ops/pull.go`

The core logic lives in `dot/internal/ops/` to avoid circular import (`cmd` imports `tui`, `tui` would need `cmd`). The existing `runPull` cobra handler in `cmd/pull.go` is NOT modified — the ops function is a parallel path for the TUI.

- [ ] **Step 1: Create PullModule function**

In `dot/internal/ops/pull.go`, create a new function that:
1. Loads all modules via `module.LoadAll`
2. Finds the named module
3. Resolves dependencies via `deps.Resolve`
4. Loads manifest
5. Calls `installModule` for each dependency in order
6. Saves manifest
7. All `fmt.Printf` calls replaced with `fmt.Fprintf(w, ...)`

```go
package ops

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// PullModule installs a single module by name, writing output to w.
func PullModule(dfPath, modName string, w io.Writer) error {
	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)
	for _, m := range allModules {
		modMap[m.Name] = m
		depsMap[m.Name] = m.Requires
	}

	if _, ok := modMap[modName]; !ok {
		return fmt.Errorf("module %q not found", modName)
	}

	order, err := deps.Resolve([]string{modName}, depsMap)
	if err != nil {
		return fmt.Errorf("resolving dependencies: %w", err)
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}
	mf.DotfilesPath = dfPath

	for _, name := range order {
		mod := modMap[name]
		fmt.Fprintf(w, "📦 安装模块: %s\n", name)
		if err := installModule(dfPath, mod, mf); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
	}

	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}
	fmt.Fprintf(w, "✓ %s 安装完成\n", modName)
	return nil
}
```

Note: This duplicates some logic from `cmd/pull.go`'s `runPull`/`installModule`. This is intentional — the ops version is a simplified path for the TUI that skips interactive elements (picker, passphrase TUI). The installModule call is reused from `cmd` by having a shared utility, or reimplemented simply here. For v1, call `installModule` from `cmd` by making it a package-level function exported via a helper, or reimplement the symlink creation inline.

Actually, the simplest approach: `PullModule` in ops/ calls the internal packages directly (module, linker, manifest, etc.) to create links — same as installModule does but without hooks and TUI interactions. This keeps it self-contained.

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/ops/pull.go
git commit -m "feat(ops): extract PullModule core logic for TUI"
```

---

### Task 3: Extract push core logic into ops package

**Files:**
- Create: `dot/internal/ops/push.go`

- [ ] **Step 1: Create PushChanges function**

Create `dot/internal/ops/push.go` with `PushChanges(dfPath, msg string, w io.Writer) error` that pushes without stdin confirmation:

```go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

// PushChanges commits and pushes dotfiles changes, writing output to w.
// Unlike runPush, it skips interactive confirmation.
func PushChanges(dfPath, msg string, w io.Writer) error {
	changes, err := gitpkg.Status(dfPath)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(changes) == 0 {
		fmt.Fprintln(w, "无变更")
		return nil
	}

	// Group changes by module for display
	grouped := make(map[string][]string)
	for _, line := range changes {
		if len(line) < 4 {
			continue
		}
		file := strings.TrimSpace(line[3:])
		if idx := strings.Index(file, " -> "); idx >= 0 {
			file = file[idx+4:]
		}
		parts := strings.SplitN(file, "/", 3)
		if len(parts) >= 2 && parts[0] == "modules" {
			grouped[parts[1]] = append(grouped[parts[1]], file)
		} else {
			grouped["other"] = append(grouped["other"], file)
		}
	}

	for mod, files := range grouped {
		fmt.Fprintf(w, "  %s: %d files\n", mod, len(files))
	}

	// Encrypt secrets if needed
	allModules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
	for _, mod := range allModules {
		for _, sec := range mod.Secrets {
			plainPath := filepath.Join(mod.Dir, sec.Source)
			encPath := filepath.Join(mod.Dir, sec.Encrypted)
			if _, err := os.Stat(plainPath); os.IsNotExist(err) {
				continue
			}
			// Try keychain passphrase silently
			pass, _ := secrets.LoadPassphrase()
			if pass == "" {
				fmt.Fprintf(w, "  ⚠ 跳过加密 %s（无 passphrase）\n", sec.Source)
				continue
			}
			changed, err := secrets.HasChanged(plainPath, encPath, pass)
			if err != nil || !changed {
				continue
			}
			fmt.Fprintf(w, "  🔐 加密 %s\n", sec.Source)
			if err := secrets.EncryptFile(plainPath, encPath, pass); err != nil {
				return fmt.Errorf("encrypting %s: %w", sec.Source, err)
			}
		}
	}

	// Auto-generate commit message if not provided
	if msg == "" {
		var names []string
		for name := range grouped {
			names = append(names, name)
		}
		sort.Strings(names)
		msg = "update " + strings.Join(names, ", ") + " config"
	}

	if err := gitpkg.AddAndCommit(dfPath, []string{"modules/", ".gitignore"}, msg); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	fmt.Fprintf(w, "✓ 已提交: %s\n", msg)

	if err := gitpkg.Push(dfPath); err != nil {
		return fmt.Errorf("push: %w", err)
	}
	fmt.Fprintln(w, "✓ 已推送")
	return nil
}
```

- [ ] **Step 2: Build and test**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/ops/push.go
git commit -m "feat(ops): extract PushChanges core logic for TUI"
```

---

### Task 4: Extract doctor core logic into ops package

**Files:**
- Create: `dot/internal/ops/doctor.go`

- [ ] **Step 1: Create DoctorCheck function for single module**

Create `dot/internal/ops/doctor.go`:

```go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

// DoctorCheck checks health of a single module, writing output to w.
func DoctorCheck(dfPath, modName string, w io.Writer) error {
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	rec, ok := mf.Modules[modName]
	if !ok {
		return fmt.Errorf("module %q not installed", modName)
	}

	issues := 0
	for _, link := range rec.Links {
		target := link.Target
		expectedSource := link.Source
		if !filepath.IsAbs(expectedSource) {
			expectedSource = filepath.Join(dfPath, expectedSource)
		}

		info, err := os.Lstat(target)
		if err != nil {
			fmt.Fprintf(w, "  ✗ %s：链接不存在\n", target)
			issues++
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			fmt.Fprintf(w, "  ✗ %s：不是符号链接\n", target)
			issues++
			continue
		}
		actual, err := os.Readlink(target)
		if err != nil {
			fmt.Fprintf(w, "  ✗ %s：无法读取链接\n", target)
			issues++
			continue
		}
		if !filepath.IsAbs(actual) {
			actual = filepath.Join(filepath.Dir(target), actual)
		}
		absActual, _ := filepath.EvalSymlinks(target)
		absExpected, _ := filepath.Abs(expectedSource)
		if absActual != absExpected {
			fmt.Fprintf(w, "  ✗ %s → %s（应为 %s）\n", target, actual, expectedSource)
			issues++
			continue
		}
		fmt.Fprintf(w, "  ✓ %s → %s\n", filepath.Base(target), target)
	}

	// Check secrets
	mod, parseErr := module.Parse(filepath.Join(dfPath, "modules", modName))
	if parseErr == nil {
		for _, sec := range mod.Secrets {
			plainPath := filepath.Join(mod.Dir, sec.Source)
			encPath := filepath.Join(mod.Dir, sec.Encrypted)

			if _, err := os.Stat(encPath); os.IsNotExist(err) {
				fmt.Fprintf(w, "  ✗ %s 加密文件不存在\n", sec.Encrypted)
				issues++
				continue
			}
			if _, err := os.Stat(plainPath); os.IsNotExist(err) {
				fmt.Fprintf(w, "  ✗ %s 未解密\n", sec.Source)
				issues++
				continue
			}
			info, _ := os.Stat(plainPath)
			if info.Mode().Perm() != 0o600 {
				fmt.Fprintf(w, "  ✗ %s 权限 %o，应为 0600\n", sec.Source, info.Mode().Perm())
				issues++
			} else {
				fmt.Fprintf(w, "  ✓ %s 已解密 · 0600\n", sec.Source)
			}
		}
	}

	if issues == 0 {
		fmt.Fprintln(w, "  All checks passed ✓")
	} else {
		fmt.Fprintf(w, "  %d issues found\n", issues)
	}
	return nil
}
```

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/ops/doctor.go
git commit -m "feat(ops): extract DoctorCheck core logic for TUI"
```

---

### Task 5: Extract remove core logic into ops package

**Files:**
- Create: `dot/internal/ops/remove.go`

- [ ] **Step 1: Create RemoveModule function (no stdin confirmation)**

Create `dot/internal/ops/remove.go`. Include backup restoration from existing `runRemove`:

```go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

// RemoveModule uninstalls a module without interactive confirmation, writing to w.
func RemoveModule(dfPath, modName string, w io.Writer) error {
	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}
	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)
	for _, m := range allModules {
		modMap[m.Name] = m
		depsMap[m.Name] = m.Requires
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if !mf.IsInstalled(modName) {
		return fmt.Errorf("module %q is not installed", modName)
	}

	// Check reverse deps
	revDeps := deps.ReverseDeps(modName, depsMap)
	for _, rd := range revDeps {
		if mf.IsInstalled(rd) {
			return fmt.Errorf("cannot remove %s: %s depends on it", modName, rd)
		}
	}

	mod := modMap[modName]
	rec := mf.Modules[modName]

	// Pre-remove hook
	if mod != nil {
		hook.Run(mod.Hooks.PreRemove, mod.Dir)
	}

	// Remove symlinks and restore backups
	for _, link := range rec.Links {
		if err := linker.RemoveLink(link.Target); err != nil {
			fmt.Fprintf(w, "  ⚠ %s: %v\n", link.Target, err)
		} else {
			fmt.Fprintf(w, "  ✓ 移除 %s\n", link.Target)
		}
		// Restore backup if exists
		if link.Backup != "" {
			if _, err := os.Stat(link.Backup); err == nil {
				if err := os.Rename(link.Backup, link.Target); err == nil {
					fmt.Fprintf(w, "  ✓ 恢复备份 %s\n", link.Target)
				}
			}
		}
	}

	// Post-remove hook
	if mod != nil {
		hook.Run(mod.Hooks.PostRemove, mod.Dir)
	}

	mf.RemoveModule(modName)
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}
	fmt.Fprintf(w, "✓ %s 已卸载\n", modName)
	return nil
}
```

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/ops/remove.go
git commit -m "feat(ops): extract RemoveModule core logic for TUI"
```

---

### Task 6: Create panel types and messages

**Files:**
- Create: `dot/internal/tui/panel.go`
- Create: `dot/internal/tui/messages.go`

- [ ] **Step 1: Create panel.go with Panel interface**

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

// Panel is the interface all dashboard panels implement.
type Panel interface {
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View(width, height int) string
	Focused() bool
	SetFocus(bool)
}
```

- [ ] **Step 2: Create messages.go with shared message types**

```go
package tui

import (
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// ModuleSelectedMsg is sent when the user selects a module in the Channel Strip.
type ModuleSelectedMsg struct {
	Index  int
	Module *module.Module
}

// CmdStartMsg signals a command has started executing.
type CmdStartMsg struct{}

// CmdOutputMsg carries command output back to the dashboard.
type CmdOutputMsg struct {
	Output string
	Err    error
}

// DataReloadMsg carries refreshed data after a command completes.
type DataReloadMsg struct {
	Modules    []*module.Module
	Manifest   *manifest.Manifest
	GitChanges []string
}
```

- [ ] **Step 3: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 4: Commit**

```bash
git add dot/internal/tui/panel.go dot/internal/tui/messages.go
git commit -m "feat(tui): add Panel interface and message types"
```

---

### Task 7: Create Channel Strip panel

**Files:**
- Create: `dot/internal/tui/panel_channel.go`

This is the top bar showing all modules as chips with LED indicators.

- [ ] **Step 1: Implement panel_channel.go**

```go
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

type ChannelStrip struct {
	modules    []*module.Module
	manifest   *manifest.Manifest
	gitChanges []string
	selected   int
	offset     int // horizontal scroll offset
}

func NewChannelStrip(modules []*module.Module, mf *manifest.Manifest, gitChanges []string) ChannelStrip {
	return ChannelStrip{modules: modules, manifest: mf, gitChanges: gitChanges}
}

func (c ChannelStrip) Update(msg tea.Msg) (ChannelStrip, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if c.selected > 0 {
				c.selected--
			}
			return c, func() tea.Msg {
				return ModuleSelectedMsg{Index: c.selected, Module: c.modules[c.selected]}
			}
		case "right":
			if c.selected < len(c.modules)-1 {
				c.selected++
			}
			return c, func() tea.Msg {
				return ModuleSelectedMsg{Index: c.selected, Module: c.modules[c.selected]}
			}
		}
	}
	return c, nil
}

func (c ChannelStrip) View(width, _ int) string {
	var chips []string

	logo := s.Blue.Bold(true).Render("DOT")
	chips = append(chips, logo)

	for i, mod := range c.modules {
		name := strings.ToUpper(mod.Name)
		led := ledFor(mod, c.manifest, c.gitChanges)
		label := name + " " + led

		if !platform.MatchesPlatform(mod.Platforms) {
			chips = append(chips, s.Dimmed.Render("["+label+"]"))
			continue
		}

		if i == c.selected {
			chips = append(chips, s.ChipSelected.Render(label))
		} else {
			chips = append(chips, s.Chip.Render(label))
		}
	}

	// ADD slot
	addSlot := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Foreground(ColorBorder).
		Padding(0, 1).
		Render("+ ADD a")
	chips = append(chips, addSlot)

	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(chips, " "))
}

func (c ChannelStrip) Selected() *module.Module {
	if len(c.modules) == 0 {
		return nil
	}
	return c.modules[c.selected]
}

func (c ChannelStrip) SelectedIndex() int { return c.selected }

func ledFor(mod *module.Module, mf *manifest.Manifest, gitChanges []string) string {
	if !mf.IsInstalled(mod.Name) {
		return s.Red.Render("●")
	}

	// Check git changes
	prefix := "modules/" + mod.Name + "/"
	for _, line := range gitChanges {
		if len(line) >= 4 && strings.Contains(line[3:], prefix) {
			label := s.Yellow.Render("●")
			if len(mod.Secrets) > 0 {
				label += " 🔐"
			}
			return label
		}
	}

	label := s.Green.Render("●")
	if len(mod.Secrets) > 0 {
		label += " 🔐"
	}
	return label
}
```

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/panel_channel.go
git commit -m "feat(tui): add Channel Strip panel"
```

---

### Task 8: Create Overview panel

**Files:**
- Create: `dot/internal/tui/panel_overview.go`

- [ ] **Step 1: Implement Overview panel**

Shows: health %, 4-stat grid, system info. Doesn't change with module selection.

Key rendering: health progress bar using block chars `█░`, stat boxes with lipgloss, git/platform/keychain info.

- [ ] **Step 2: Build and commit**

```bash
git add dot/internal/tui/panel_overview.go
git commit -m "feat(tui): add Overview panel"
```

---

### Task 9: Create Scope panel

**Files:**
- Create: `dot/internal/tui/panel_scope.go`

- [ ] **Step 1: Implement Scope panel**

Shows for selected module: meter bars (LNK/SEC/GIT progress), signal path (links with status), deps.

Meter bars: use `█` (filled) and `░` (empty) chars with lipgloss colors. Progress = healthy / total.

Signal path: list each link with ● prefix, green if healthy, red if broken, yellow for secrets.

- [ ] **Step 2: Build and commit**

```bash
git add dot/internal/tui/panel_scope.go
git commit -m "feat(tui): add Scope panel with meters"
```

---

### Task 10: Create Terminal panel

**Files:**
- Create: `dot/internal/tui/panel_terminal.go`

- [ ] **Step 1: Implement Terminal panel**

Features:
- `lines []string` buffer (max 1000 lines)
- `input string` for command input
- `inputMode bool` toggled by `:` / `esc`
- Render: scrollable output area + `dot ▸` input prompt at bottom
- Handle `CmdOutputMsg` to append output lines
- Handle key input when in input mode (runes, backspace, enter)

When enter is pressed with input, emit a `tea.Cmd` that returns a custom `TerminalExecMsg{Args: parsed_args}` for the dashboard to handle.

- [ ] **Step 2: Build and commit**

```bash
git add dot/internal/tui/panel_terminal.go
git commit -m "feat(tui): add Terminal panel"
```

---

### Task 11: Create Controls panel

**Files:**
- Create: `dot/internal/tui/panel_controls.go`

- [ ] **Step 1: Implement Controls panel**

Four horizontal buttons: PULL(p) / PUSH(P) / DOCTOR(d) / REMOVE(x). Each rendered as a lipgloss box with icon + label + key. When `executing` is true, show spinner text instead.

- [ ] **Step 2: Build and commit**

```bash
git add dot/internal/tui/panel_controls.go
git commit -m "feat(tui): add Controls panel"
```

---

### Task 12: Create commands bridge

**Files:**
- Create: `dot/internal/tui/commands.go`

- [ ] **Step 1: Implement command execution bridge**

```go
package tui

import (
	"bytes"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/ops"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

// execPull returns a tea.Cmd that runs PullModule.
func execPull(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		err := ops.PullModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

// execPush returns a tea.Cmd that runs PushChanges.
func execPush(dfPath, msg string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		err := ops.PushChanges(dfPath, msg, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

// execDoctor returns a tea.Cmd that runs DoctorCheck.
func execDoctor(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot doctor %s\n", modName)
		err := ops.DoctorCheck(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

// execRemove returns a tea.Cmd that runs RemoveModule.
func execRemove(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		err := ops.RemoveModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

// reloadData returns a tea.Cmd that refreshes all dashboard data.
func reloadData(dfPath string) tea.Cmd {
	return func() tea.Msg {
		modules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
		mfPath, _ := manifest.DefaultPath()
		mf, _ := manifest.Load(mfPath)
		gitChanges, _ := gitpkg.Status(dfPath)
		return DataReloadMsg{Modules: modules, Manifest: mf, GitChanges: gitChanges}
	}
}
```

Note: Imports from `ops` package (not `cmd`), avoiding the circular import.

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/commands.go
git commit -m "feat(tui): add command execution bridge"
```

---

### Task 13: Create Dashboard model

**Files:**
- Create: `dot/internal/tui/dashboard.go`

- [ ] **Step 1: Implement Dashboard model**

The main bubbletea model that:
- Holds all 5 panels + shared state (modules, manifest, gitChanges, selected, executing, width, height)
- `Init()`: returns `tea.WindowSize` command
- `Update()`:
  - `tea.WindowSizeMsg` → store width/height
  - `tea.KeyMsg` → route based on focus: `←→` to channel strip, `tab` to cycle focus, `:` to focus terminal, `q` to quit, `p/P/d/x` to execute commands (if not executing), `a` to quit TUI with message "run: dot add <path>"
  - `ModuleSelectedMsg` → update selected, refresh scope
  - `CmdStartMsg` → set executing=true
  - `CmdOutputMsg` → append to terminal, set executing=false, fire reloadData
  - `DataReloadMsg` → update all panels with new data
- `RunDashboard()`: public entry point that creates a `tea.Program` and runs it

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/dashboard.go
git commit -m "feat(tui): add Dashboard model"
```

---

### Task 14: Create Dashboard view

**Files:**
- Create: `dot/internal/tui/dashboard_view.go`

- [ ] **Step 1: Implement Dashboard View**

Layout calculation:
- Total width/height from `tea.WindowSizeMsg`
- If width < 80 || height < 24 → render "终端过小" message
- Channel Strip: full width, 3 rows
- Controls: full width, 3 rows
- Footer: full width, 1 row
- Middle area: remaining height split among Overview (fixed 22 cols) | Scope (50%) | Terminal (50%)
- Use `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` to compose

- [ ] **Step 2: Build**

Run: `cd dot && go build -o /dev/null .`

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/dashboard_view.go
git commit -m "feat(tui): add Dashboard view and layout"
```

---

### Task 15: Wire up root.go entry point

**Files:**
- Modify: `dot/cmd/root.go`

- [ ] **Step 1: Add RunE to rootCmd**

In `root.go`, set `rootCmd.RunE` to launch the TUI dashboard when no subcommand is given:

```go
var rootCmd = &cobra.Command{
	Use:   "dot",
	Short: "Dotfiles module manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunDashboard()
	},
}
```

Add import for `tui` package.

- [ ] **Step 2: Build and test manually**

```bash
cd dot && go build -o dot .
./dot          # Should launch TUI dashboard
./dot status   # Should still work as CLI
./dot --help   # Should show help with subcommands
```

- [ ] **Step 3: Commit**

```bash
git add dot/cmd/root.go
git commit -m "feat: launch TUI dashboard on bare dot command"
```

---

### Task 16: Integration test and polish

**Files:**
- Various TUI files (bug fixes found during testing)

- [ ] **Step 1: Build the full binary**

```bash
cd dot && go build -o dot .
```

- [ ] **Step 2: Test TUI launch**

```bash
./dot
```

Verify:
- Channel Strip shows all modules with correct LED status
- Overview shows correct stats
- ←→ switches modules, Scope updates
- `d` runs doctor on selected module, output in Terminal
- `q` exits cleanly

- [ ] **Step 3: Test CLI compatibility**

```bash
./dot status
./dot doctor
./dot pull --help
```

All existing CLI commands should work exactly as before.

- [ ] **Step 4: Fix any issues found**

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: TUI dashboard polish and fixes"
```
