# Implementation Alignment Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the dot CLI and TUI implementation with design documents (`docs/design/cli.md`, `docs/design/tui.md`).

**Architecture:** Two-batch approach. Batch 1 restructures CLI commands (rename remove→uninstall, split pull→pull+install, add new remove, complete clean). Batch 2 aligns TUI dashboard (fixed chips, scroll indicators, overview fields, command history, confirmation flow, Ctrl shortcuts, responsive degradation). Each batch ends with `go build` verification.

**Tech Stack:** Go 1.26, cobra, bubbletea + lipgloss, filippo.io/age

---

## File Structure

### New Files
- `dot/cmd/install.go` — install command (extracted from pull.go)
- `dot/cmd/uninstall.go` — uninstall command (renamed from remove.go)
- `dot/internal/ops/install.go` — InstallModule function (extracted from ops/pull.go)
- `dot/internal/ops/uninstall.go` — UninstallModule function (renamed from ops/remove.go)

### Modified Files (Batch 1: CLI)
- `dot/cmd/pull.go` — simplify to git-pull-only
- `dot/cmd/remove.go` — new: delete module files from modules/
- `dot/cmd/clean.go` — add orphan link + invalid manifest cleanup
- `dot/internal/ops/pull.go` — simplify to PullRepo
- `dot/internal/ops/remove.go` — new: RemoveModuleFiles
- `dot/internal/git/git.go` — add AheadBehind function
- `dot/internal/tui/commands.go` — update command dispatch
- `dot/internal/tui/dashboard.go` — update key bindings and terminal dispatch
- `dot/internal/tui/dashboard_view.go` — update footer hints, add degradation
- `dot/internal/tui/panel_controls.go` — rename buttons, add confirmation mode

### Modified Files (Batch 2: TUI)
- `dot/internal/tui/layout.go` — fixed chip size, viewport scroll offset
- `dot/internal/tui/panel_channel.go` — fixed chips, scroll indicators ◂▸
- `dot/internal/tui/panel_overview.go` — health formula with secrets, add fields, sync display
- `dot/internal/tui/panel_terminal.go` — command history ↑↓
- `dot/internal/tui/dashboard.go` — Ctrl shortcuts in input mode, remove confirmation
- `dot/internal/tui/dashboard_view.go` — responsive panel degradation

---

## Batch 1: CLI Command Restructuring

### Task 1: Rename remove → uninstall (ops layer)

**Files:**
- Rename: `dot/internal/ops/remove.go` → `dot/internal/ops/uninstall.go`

- [ ] **Step 1: Create ops/uninstall.go from ops/remove.go**

Copy `dot/internal/ops/remove.go` to `dot/internal/ops/uninstall.go`, rename the function from `RemoveModule` to `UninstallModule`:

```go
// File: dot/internal/ops/uninstall.go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

// UninstallModule removes symlinks for a single named module without interactive confirmation.
// Module files in modules/ are preserved. It writes progress output to w.
func UninstallModule(dfPath, modName string, w io.Writer) error {
	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)

	for _, mod := range allModules {
		if !platform.MatchesPlatform(mod.Platforms) {
			continue
		}
		modMap[mod.Name] = mod
		depsMap[mod.Name] = mod.Requires
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
		return fmt.Errorf("模块 %q 未安装", modName)
	}

	revDeps := deps.ReverseDeps(modName, depsMap)
	var installedRevDeps []string
	for _, rd := range revDeps {
		if mf.IsInstalled(rd) {
			installedRevDeps = append(installedRevDeps, rd)
		}
	}
	if len(installedRevDeps) > 0 {
		return fmt.Errorf("模块 %q ���以下已安装模块依赖：%s", modName, strings.Join(installedRevDeps, ", "))
	}

	if mod, ok := modMap[modName]; ok {
		if err := hook.Run(mod.Hooks.PreRemove, mod.Dir); err != nil {
			fmt.Fprintf(w, "  警告：%s pre_remove hook 失败：%v\n", modName, err)
		}
	}

	if rec, ok := mf.Modules[modName]; ok {
		for _, link := range rec.Links {
			if err := linker.RemoveLink(link.Target); err != nil {
				fmt.Fprintf(w, "  警告：移除链接 %s 失败：%v\n", link.Target, err)
			}
			if link.Backup != "" {
				if _, err := os.Stat(link.Backup); err == nil {
					if err := os.Rename(link.Backup, link.Target); err != nil {
						fmt.Fprintf(w, "  警告：恢复备份 %s 失败：%v\n", link.Backup, err)
					} else {
						fmt.Fprintf(w, "  ✓ 恢复备份 %s\n", link.Target)
					}
				}
			}
		}
	}

	if mod, ok := modMap[modName]; ok {
		if err := hook.Run(mod.Hooks.PostRemove, mod.Dir); err != nil {
			fmt.Fprintf(w, "  警告：%s post_remove hook 失败：%v\n", modName, err)
		}
	}

	mf.RemoveModule(modName)

	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Fprintf(w, "  ✓ %s：已卸载\n", modName)
	return nil
}
```

- [ ] **Step 2: Delete old ops/remove.go**

Run: `rm -f dot/internal/ops/remove.go`

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: compilation errors in files referencing `ops.RemoveModule` — these will be fixed in subsequent tasks.

- [ ] **Step 4: Commit**

```bash
git add dot/internal/ops/uninstall.go
git add dot/internal/ops/remove.go  # deleted
git commit -m "refactor: rename ops.RemoveModule to ops.UninstallModule"
```

---

### Task 2: Rename cmd/remove.go → cmd/uninstall.go

**Files:**
- Rename: `dot/cmd/remove.go` → `dot/cmd/uninstall.go`

- [ ] **Step 1: Create cmd/uninstall.go from cmd/remove.go**

Change command name from `remove` to `uninstall`, update Short description:

```go
// File: dot/cmd/uninstall.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <modules...>",
	Short: "Uninstall modules (remove symlinks, keep module files)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runUninstall,
}

func init() { rootCmd.AddCommand(uninstallCmd) }

func runUninstall(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)

	for _, mod := range allModules {
		if !platform.MatchesPlatform(mod.Platforms) {
			continue
		}
		modMap[mod.Name] = mod
		depsMap[mod.Name] = mod.Requires
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for _, name := range args {
		if !mf.IsInstalled(name) {
			fmt.Printf("  ⊘ %s：未安装，跳过\n", name)
			continue
		}

		revDeps := deps.ReverseDeps(name, depsMap)
		var installedRevDeps []string
		for _, rd := range revDeps {
			if mf.IsInstalled(rd) {
				installedRevDeps = append(installedRevDeps, rd)
			}
		}
		if len(installedRevDeps) > 0 {
			fmt.Printf("  ⚠ %s 被以下已安装模块依赖：%s\n", name, strings.Join(installedRevDeps, ", "))
			fmt.Print("  确认卸载？(y/N) ")
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" {
				fmt.Printf("  跳过 %s\n", name)
				continue
			}
		}

		if mod, ok := modMap[name]; ok {
			if err := hook.Run(mod.Hooks.PreRemove, mod.Dir); err != nil {
				fmt.Fprintf(os.Stderr, "  警告：%s pre_remove hook 失败：%v\n", name, err)
			}
		}

		if rec, ok := mf.Modules[name]; ok {
			for _, link := range rec.Links {
				if err := linker.RemoveLink(link.Target); err != nil {
					fmt.Fprintf(os.Stderr, "  警告：移除链接 %s 失败：%v\n", link.Target, err)
				}
				if link.Backup != "" {
					if _, err := os.Stat(link.Backup); err == nil {
						if err := os.Rename(link.Backup, link.Target); err != nil {
							fmt.Fprintf(os.Stderr, "  警告：恢复备份 %s 失败：%v\n", link.Backup, err)
						} else {
							fmt.Printf("  ✓ 恢复备份 %s\n", link.Target)
						}
					}
				}
			}
		}

		if mod, ok := modMap[name]; ok {
			if err := hook.Run(mod.Hooks.PostRemove, mod.Dir); err != nil {
				fmt.Fprintf(os.Stderr, "  警告：%s post_remove hook 失败：%v\n", name, err)
			}
		}

		mf.RemoveModule(name)
		fmt.Printf("  ✓ %s：已卸载\n", name)
	}

	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: Delete old cmd/remove.go**

Run: `rm -f dot/cmd/remove.go`

- [ ] **Step 3: Create new cmd/remove.go (delete module files)**

```go
// File: dot/cmd/remove.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

var removeCmd = &cobra.Command{
	Use:   "remove <modules...>",
	Short: "Permanently delete module files from modules/ directory",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func init() { rootCmd.AddCommand(removeCmd) }

func runRemove(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for _, name := range args {
		modDir := filepath.Join(dfPath, "modules", name)

		// Check module directory exists
		if _, err := os.Stat(modDir); os.IsNotExist(err) {
			fmt.Printf("  ⊘ %s：模块目录不存在，跳过\n", name)
			continue
		}

		// Check module is not installed (must uninstall first)
		if mf.IsInstalled(name) {
			fmt.Printf("  ⚠ %s：仍然已安装，请先运行 dot uninstall %s\n", name, name)
			continue
		}

		// Parse module for display
		mod, _ := module.Parse(modDir)
		desc := name
		if mod != nil && mod.Description != "" {
			desc = name + " (" + mod.Description + ")"
		}

		// Confirm deletion
		fmt.Printf("  将永久删除 modules/%s/ 目录，确认？(y/N) ", desc)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" {
			fmt.Printf("  跳过 %s\n", name)
			continue
		}

		// Delete module directory
		if err := os.RemoveAll(modDir); err != nil {
			return fmt.Errorf("删除 %s: %w", modDir, err)
		}

		fmt.Printf("  ✓ %s：已从 modules/ 删除\n", name)
	}

	return nil
}
```

- [ ] **Step 4: Create new ops/remove.go (for TUI use)**

```go
// File: dot/internal/ops/remove.go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
)

// RemoveModuleFiles permanently deletes a module's directory from modules/.
// The module must be uninstalled first (not in manifest).
func RemoveModuleFiles(dfPath, modName string, w io.Writer) error {
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if mf.IsInstalled(modName) {
		return fmt.Errorf("模块 %q 仍然已安装，请先 uninstall", modName)
	}

	modDir := filepath.Join(dfPath, "modules", modName)
	if _, err := os.Stat(modDir); os.IsNotExist(err) {
		return fmt.Errorf("模块目录 %s 不存在", modDir)
	}

	if err := os.RemoveAll(modDir); err != nil {
		return fmt.Errorf("删除 %s: %w", modDir, err)
	}

	fmt.Fprintf(w, "  ✓ %s：已从 modules/ 删除\n", modName)
	return nil
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: May still fail due to tui/commands.go referencing old `ops.RemoveModule`. Will fix in Task 5.

- [ ] **Step 6: Commit**

```bash
git add dot/cmd/uninstall.go dot/cmd/remove.go dot/internal/ops/remove.go
git commit -m "refactor: split remove into uninstall + remove per design"
```

---

### Task 3: Split pull → pull + install (ops layer)

**Files:**
- Create: `dot/internal/ops/install.go`
- Modify: `dot/internal/ops/pull.go`

- [ ] **Step 1: Create ops/install.go (extracted from ops/pull.go)**

Move `PullModule` and `installModuleOps` to `install.go`, rename `PullModule` to `InstallModule`:

```go
// File: dot/internal/ops/install.go
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
	"github.com/aporicho/dotfiles/dot/internal/secrets"
	"github.com/aporicho/dotfiles/dot/internal/sysdep"
)

// InstallModule installs a single named module and its dependencies.
// It writes progress output to w. Passphrases are retrieved from the keychain
// only; if none is available, encrypted secrets are skipped with a warning.
func InstallModule(dfPath, modName string, w io.Writer) error {
	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)

	for _, mod := range allModules {
		if !platform.MatchesPlatform(mod.Platforms) {
			continue
		}
		modMap[mod.Name] = mod
		depsMap[mod.Name] = mod.Requires
	}

	if _, ok := modMap[modName]; !ok {
		return fmt.Errorf("模块 %q 不存在或不适用于当前平台", modName)
	}

	order, err := deps.Resolve([]string{modName}, depsMap)
	if err != nil {
		return fmt.Errorf("dependency resolution: %w", err)
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
		fmt.Fprintf(w, "→ 安装 %s ...\n", name)
		if err := installModuleOps(dfPath, mod, mf, w); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
		fmt.Fprintf(w, "  ✓ %s 安装完成\n", name)
	}

	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	return nil
}

// installModuleOps runs hooks, installs system deps, creates symlinks, handles
// secrets (keychain only, no interactive prompt), and records in manifest.
func installModuleOps(dfPath string, mod *module.Module, mf *manifest.Manifest, w io.Writer) error {
	modDir := mod.Dir
	var createdLinks []*linker.LinkResult

	rollback := func() {
		for _, lr := range createdLinks {
			_ = linker.RemoveLink(lr.Target)
			if lr.BackupPath != "" {
				_ = os.Rename(lr.BackupPath, lr.Target)
			}
		}
	}

	if err := hook.Run(mod.Hooks.PreInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("pre_install hook: %w", err)
	}

	if err := sysdep.Install(mod.Deps.Darwin, mod.Deps.Linux); err != nil {
		fmt.Fprintf(w, "  警告：系统依赖安装失败：%v\n", err)
	}

	for _, link := range mod.Links {
		if !platform.MatchesPlatform(link.Platforms) {
			continue
		}

		target := expandHomeOps(link.Target)

		if link.Source == "." {
			exclude := mod.Exclude
			if exclude == nil {
				exclude = []string{"module.toml", "README*", "LICENSE*", ".gitignore"}
			}
			entries, err := linker.ExpandDirEntries(modDir, exclude)
			if err != nil {
				rollback()
				return fmt.Errorf("expanding dir entries: %w", err)
			}
			for _, entry := range entries {
				entryTarget := filepath.Join(target, filepath.Base(entry))
				lr, err := linker.CreateLink(entry, entryTarget)
				if err != nil {
					rollback()
					return fmt.Errorf("linking %s: %w", entry, err)
				}
				createdLinks = append(createdLinks, lr)
			}
		} else {
			source := filepath.Join(modDir, link.Source)
			lr, err := linker.CreateLink(source, target)
			if err != nil {
				rollback()
				return fmt.Errorf("linking %s -> %s: %w", source, target, err)
			}
			createdLinks = append(createdLinks, lr)
		}
	}

	for _, sec := range mod.Secrets {
		encPath := filepath.Join(modDir, sec.Encrypted)
		plainPath := filepath.Join(modDir, sec.Source)
		target := expandHomeOps(sec.Target)

		if _, err := os.Stat(encPath); os.IsNotExist(err) {
			continue
		}

		if _, err := os.Stat(plainPath); os.IsNotExist(err) {
			if !secrets.KeychainAvailable() {
				fmt.Fprintf(w, "  警告：跳过 %s（无 Keychain 密码短语）\n", sec.Encrypted)
				continue
			}
			passphrase, err := secrets.LoadPassphrase()
			if err != nil || passphrase == "" {
				fmt.Fprintf(w, "  警告：跳过 %s（无法从 Keychain 获取密码短语）\n", sec.Encrypted)
				continue
			}
			fmt.Fprintf(w, "  解密 %s\n", sec.Encrypted)
			if err := secrets.DecryptFile(encPath, plainPath, passphrase); err != nil {
				return fmt.Errorf("decrypting %s: %w", sec.Encrypted, err)
			}
		}

		lr, err := linker.CreateLink(plainPath, target)
		if err != nil {
			return fmt.Errorf("linking secret %s: %w", sec.Source, err)
		}
		createdLinks = append(createdLinks, lr)
	}

	if err := hook.Run(mod.Hooks.PostInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("post_install hook: %w", err)
	}

	var records []manifest.LinkRecord
	for _, lr := range createdLinks {
		records = append(records, manifest.LinkRecord{
			Source: lr.Source,
			Target: lr.Target,
			Backup: lr.BackupPath,
		})
	}
	mf.AddModule(mod.Name, records)

	return nil
}
```

- [ ] **Step 2: Simplify ops/pull.go to PullRepo only**

Replace entire file:

```go
// File: dot/internal/ops/pull.go
package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

// PullRepo executes git pull in the dotfiles repository.
func PullRepo(dfPath string, w io.Writer) error {
	fmt.Fprintln(w, "$ git pull")
	if err := gitpkg.Pull(dfPath); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	fmt.Fprintln(w, "✓ 拉取完成")
	return nil
}

// expandHomeOps replaces a leading "~/" with the user's home directory.
func expandHomeOps(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
```

- [ ] **Step 3: Verify compilation of ops package**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./internal/ops/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add dot/internal/ops/install.go dot/internal/ops/pull.go
git commit -m "refactor: split ops.PullModule into ops.InstallModule + ops.PullRepo"
```

---

### Task 4: Split cmd/pull.go → cmd/pull.go + cmd/install.go

**Files:**
- Create: `dot/cmd/install.go`
- Modify: `dot/cmd/pull.go`

- [ ] **Step 1: Create cmd/install.go (extracted install logic)**

```go
// File: dot/cmd/install.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
	"github.com/aporicho/dotfiles/dot/internal/sysdep"
	"github.com/aporicho/dotfiles/dot/internal/tui"
)

var installAll bool

var cachedPassphrase string

func getPassphrase() (string, error) {
	if cachedPassphrase != "" {
		return cachedPassphrase, nil
	}
	if secrets.KeychainAvailable() {
		if p, _ := secrets.LoadPassphrase(); p != "" {
			cachedPassphrase = p
			return p, nil
		}
	}
	for i := 0; i < 3; i++ {
		p, err := tui.RunPassphraseInput("Passphrase:", false)
		if err != nil {
			return "", err
		}
		cachedPassphrase = p
		return p, nil
	}
	return "", fmt.Errorf("超过最大重试次数")
}

var installCmd = &cobra.Command{
	Use:   "install [modules...]",
	Short: "Install modules (create symlinks, deps, secrets)",
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&installAll, "all", false, "Install all available modules")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	allModules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	modMap := make(map[string]*module.Module)
	depsMap := make(map[string][]string)
	var available []string

	for _, mod := range allModules {
		if !platform.MatchesPlatform(mod.Platforms) {
			continue
		}
		modMap[mod.Name] = mod
		depsMap[mod.Name] = mod.Requires
		available = append(available, mod.Name)
	}

	var requested []string
	if installAll {
		requested = available
	} else if len(args) > 0 {
		for _, name := range args {
			if _, ok := modMap[name]; !ok {
				return fmt.Errorf("模块 %q 不存在或不适用于当前平台", name)
			}
		}
		requested = args
	} else {
		// TUI mode: interactive module picker
		mfPath, err := manifest.DefaultPath()
		if err != nil {
			return err
		}
		mf, err := manifest.Load(mfPath)
		if err != nil {
			return fmt.Errorf("loading manifest: %w", err)
		}

		var items []tui.ModuleItem
		for _, name := range available {
			items = append(items, tui.ModuleItem{
				Name:      name,
				Installed: mf.IsInstalled(name),
				HasUpdate: false,
			})
		}

		selected, err := tui.RunPicker(items)
		if err != nil {
			return err
		}
		if selected == nil {
			fmt.Println("取消操作")
			return nil
		}
		if len(selected) == 0 {
			fmt.Println("未选择任何模块")
			return nil
		}
		requested = selected
	}

	order, err := deps.Resolve(requested, depsMap)
	if err != nil {
		return fmt.Errorf("dependency resolution: %w", err)
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
		fmt.Printf("→ 安装 %s ...\n", name)
		if err := installModule(dfPath, mod, mf); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
		fmt.Printf("  ✓ %s 安装完成\n", name)
	}

	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Println("全部完成。")
	return nil
}

// installModule runs hooks, installs deps, creates symlinks, and records in manifest.
func installModule(dfPath string, mod *module.Module, mf *manifest.Manifest) error {
	modDir := mod.Dir
	var createdLinks []*linker.LinkResult

	rollback := func() {
		for _, lr := range createdLinks {
			_ = linker.RemoveLink(lr.Target)
			if lr.BackupPath != "" {
				_ = os.Rename(lr.BackupPath, lr.Target)
			}
		}
	}

	if err := hook.Run(mod.Hooks.PreInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("pre_install hook: %w", err)
	}

	if err := sysdep.Install(mod.Deps.Darwin, mod.Deps.Linux); err != nil {
		fmt.Fprintf(os.Stderr, "  警告：系统依赖安装失败：%v\n", err)
	}

	for _, link := range mod.Links {
		if !platform.MatchesPlatform(link.Platforms) {
			continue
		}

		target := expandHome(link.Target)

		if link.Source == "." {
			exclude := mod.Exclude
			if exclude == nil {
				exclude = []string{"module.toml", "README*", "LICENSE*", ".gitignore"}
			}
			entries, err := linker.ExpandDirEntries(modDir, exclude)
			if err != nil {
				rollback()
				return fmt.Errorf("expanding dir entries: %w", err)
			}
			for _, entry := range entries {
				entryTarget := filepath.Join(target, filepath.Base(entry))
				lr, err := linker.CreateLink(entry, entryTarget)
				if err != nil {
					rollback()
					return fmt.Errorf("linking %s: %w", entry, err)
				}
				createdLinks = append(createdLinks, lr)
			}
		} else {
			source := filepath.Join(modDir, link.Source)
			lr, err := linker.CreateLink(source, target)
			if err != nil {
				rollback()
				return fmt.Errorf("linking %s -> %s: %w", source, target, err)
			}
			createdLinks = append(createdLinks, lr)
		}
	}

	for _, sec := range mod.Secrets {
		encPath := filepath.Join(modDir, sec.Encrypted)
		plainPath := filepath.Join(modDir, sec.Source)
		target := expandHome(sec.Target)

		if _, err := os.Stat(encPath); os.IsNotExist(err) {
			continue
		}

		if _, err := os.Stat(plainPath); os.IsNotExist(err) {
			passphrase, err := getPassphrase()
			if err != nil {
				return fmt.Errorf("getting passphrase: %w", err)
			}
			fmt.Printf("  解密 %s\n", sec.Encrypted)
			if err := secrets.DecryptFile(encPath, plainPath, passphrase); err != nil {
				return fmt.Errorf("decrypting %s: %w", sec.Encrypted, err)
			}
			if secrets.KeychainAvailable() {
				secrets.OfferSaveToKeychain(passphrase)
				fmt.Println("  ✓ 已保存到系统 Keychain")
			}
		}

		lr, err := linker.CreateLink(plainPath, target)
		if err != nil {
			return fmt.Errorf("linking secret %s: %w", sec.Source, err)
		}
		createdLinks = append(createdLinks, lr)
	}

	if err := hook.Run(mod.Hooks.PostInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("post_install hook: %w", err)
	}

	var records []manifest.LinkRecord
	for _, lr := range createdLinks {
		records = append(records, manifest.LinkRecord{
			Source: lr.Source,
			Target: lr.Target,
			Backup: lr.BackupPath,
		})
	}
	mf.AddModule(mod.Name, records)

	return nil
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
```

- [ ] **Step 2: Simplify cmd/pull.go to git-pull-only**

Replace entire file:

```go
// File: dot/cmd/pull.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest changes from remote (git pull only)",
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	fmt.Println("$ git pull")
	if err := gitpkg.Pull(dfPath); err != nil {
		fmt.Fprintf(os.Stderr, "git pull 失败：%v\n", err)
		return err
	}
	fmt.Println("✓ 拉取完成")
	return nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./cmd/...`
Expected: May fail due to `expandHome` being defined in both install.go and pull.go. The `expandHome` function was in old pull.go. It's now in install.go. Make sure there's no duplicate. If add.go also has it, one of them needs to be removed. Check: `expandHome` is used in `cmd/install.go` and `cmd/add.go` references it in `panel_scope.go`. In cmd package, it was originally in pull.go. Now it moves to install.go. The add.go file doesn't define `expandHome` — it was calling the one from pull.go. Since they're in the same package, install.go's definition covers it.

- [ ] **Step 4: Commit**

```bash
git add dot/cmd/install.go dot/cmd/pull.go
git commit -m "refactor: split pull into pull (git only) + install (module setup)"
```

---

### Task 5: Update TUI command dispatcher + button labels

**Files:**
- Modify: `dot/internal/tui/commands.go`
- Modify: `dot/internal/tui/dashboard.go`
- Modify: `dot/internal/tui/panel_controls.go`
- Modify: `dot/internal/tui/dashboard_view.go`

- [ ] **Step 1: Update tui/commands.go**

Replace `execPull` with `execInstall`, rename `execRemove` to `execUninstall`, add `execPull` for git pull:

```go
// File: dot/internal/tui/commands.go
package tui

import (
	"bytes"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/ops"
)

func execInstall(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot install %s\n", modName)
		err := ops.InstallModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execPull(dfPath string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		err := ops.PullRepo(dfPath, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execPush(dfPath, msg string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot push\n")
		err := ops.PushChanges(dfPath, msg, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execDoctor(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot doctor %s\n", modName)
		err := ops.DoctorCheck(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execUninstall(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot uninstall %s\n", modName)
		err := ops.UninstallModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func reloadData(dfPath string) tea.Cmd {
	return func() tea.Msg {
		modules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
		mfPath, _ := manifest.DefaultPath()
		mf, _ := manifest.Load(mfPath)
		gitChanges, _ := git.Status(dfPath)
		return DataReloadMsg{Modules: modules, Manifest: mf, GitChanges: gitChanges}
	}
}
```

- [ ] **Step 2: Update dashboard.go key bindings**

In `dashboard.go`, update the `handleKey` method. Change `"p"` to call `execInstall`, `"x"` to call `execUninstall`. Update `handleTerminalExec` to handle `install`, `uninstall`, `pull`, `remove` commands correctly.

In `handleKey`:
- `"p"` → `execInstall(d.dfPath, mod.Name)` (was `execPull`)
- `"x"` → `execUninstall(d.dfPath, mod.Name)` (was `execRemove`)

In `handleTerminalExec`:
- `"install"` → `execInstall` (new)
- `"pull"` → `execPull(d.dfPath)` (git pull only, no module arg)
- `"uninstall"` → `execUninstall` (new)
- `"remove"` → display message "请使用 dot remove <module> 命令行操作"（安全考虑：TUI 内不执行不可逆的文件删除操作，偏离 spec C5 的 remove 映射）
- `"push"` → unchanged
- `"doctor"` → unchanged

- [ ] **Step 3: Update panel_controls.go button labels**

Change button labels from PULL/REMOVE to INSTALL/UNINSTALL:

In `View` method, change `buttons` slice:
```go
buttons := []btnDef{
    {"\uf0ed", "INSTALL", "p", c.theme.BtnPull()},
    {"\uf0ee", "PUSH", "P", c.theme.BtnPush()},
    {"\uf21e", "DOCTOR", "d", c.theme.BtnDoctor()},
    {"\uf1f8", "UNINSTALL", "x", c.theme.BtnRemove()},
}
```

- [ ] **Step 4: Update footer hints in dashboard_view.go**

In `renderFooter`, update hint text:
```go
hint := "←→ module · : terminal · esc back · p install · P push · d doctor · x uninstall · a add · q quit"
```

- [ ] **Step 5: Verify full compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS — all references to old function names should be resolved.

- [ ] **Step 6: Commit**

```bash
git add dot/internal/tui/commands.go dot/internal/tui/dashboard.go dot/internal/tui/panel_controls.go dot/internal/tui/dashboard_view.go
git commit -m "refactor: update TUI to use install/uninstall commands"
```

---

### Task 6: Complete clean command

**Files:**
- Modify: `dot/cmd/clean.go`

- [ ] **Step 1: Add orphan link and invalid manifest cleanup**

After the existing backup cleanup code (line 123), add two new sections. Restructure the function to present all three categories separately. The full replacement for `runClean` adds:

1. Orphan links: iterate manifest link records, check if the symlink target exists and source exists. If not, offer to remove the link record.
2. Invalid manifest: iterate manifest module records, check if `modules/<name>/` still exists. If not, offer to remove the manifest record.

Add these helper functions and integrate into `runClean`. The key additions to the existing file:

After `"✓ 已清理 %d 个备份文件\n"` section, add orphan link section:

```go
// --- Orphan links ---
var orphanLinks []struct{ Module, Target string }
for modName, rec := range mf.Modules {
    for _, link := range rec.Links {
        target := link.Target
        info, err := os.Lstat(target)
        if err != nil || info.Mode()&os.ModeSymlink == 0 {
            orphanLinks = append(orphanLinks, struct{ Module, Target string }{modName, target})
            continue
        }
        // Check if source still exists
        actual, err := os.Readlink(target)
        if err != nil {
            orphanLinks = append(orphanLinks, struct{ Module, Target string }{modName, target})
            continue
        }
        if _, err := os.Stat(actual); os.IsNotExist(err) {
            orphanLinks = append(orphanLinks, struct{ Module, Target string }{modName, target})
        }
    }
}

if len(orphanLinks) > 0 {
    fmt.Printf("\n找到 %d 个孤立链接：\n", len(orphanLinks))
    for _, ol := range orphanLinks {
        fmt.Printf("  [%s] %s\n", ol.Module, ol.Target)
    }
    fmt.Print("清理孤立链接？(Y/n) ")
    answer, _ = reader.ReadString('\n')
    answer = strings.TrimSpace(strings.ToLower(answer))
    if answer != "n" && answer != "no" {
        cleaned := 0
        for _, ol := range orphanLinks {
            _ = os.Remove(ol.Target) // best-effort
            cleaned++
        }
        fmt.Printf("✓ 已清理 %d 个孤立链接\n", cleaned)
    }
}
```

Then add invalid manifest section:

```go
// --- Invalid manifest records ---
modulesDir := filepath.Join(dfPath, "modules")
var invalidMods []string
for modName := range mf.Modules {
    modDir := filepath.Join(modulesDir, modName)
    if _, err := os.Stat(modDir); os.IsNotExist(err) {
        invalidMods = append(invalidMods, modName)
    }
}

if len(invalidMods) > 0 {
    fmt.Printf("\n找到 %d 个无效 manifest 记录（模块目录已不存在）：\n", len(invalidMods))
    for _, name := range invalidMods {
        fmt.Printf("  %s\n", name)
    }
    fmt.Print("清理无效记录？(Y/n) ")
    answer, _ = reader.ReadString('\n')
    answer = strings.TrimSpace(strings.ToLower(answer))
    if answer != "n" && answer != "no" {
        for _, name := range invalidMods {
            mf.RemoveModule(name)
        }
        if err := manifest.Save(mf, mfPath); err != nil {
            return fmt.Errorf("saving manifest: %w", err)
        }
        fmt.Printf("✓ 已清理 %d 个无效记录\n", len(invalidMods))
    }
}
```

Note: `runClean` will also need access to `dfPath` — add `DotfilesPath()` call at the top, and add the necessary import for `filepath`.

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add dot/cmd/clean.go
git commit -m "feat: complete clean command with orphan links and invalid manifest cleanup"
```

---

## Batch 2: TUI Dashboard Alignment

### Task 7: Fixed chip size + scroll indicators (T1+T2)

**Files:**
- Modify: `dot/internal/tui/layout.go`
- Modify: `dot/internal/tui/panel_channel.go`
- Modify: `dot/internal/tui/dashboard_view.go`

- [ ] **Step 1: Fix chip size in layout.go**

In `ComputeLayout`, replace the dynamic chip calculation with fixed values and add scroll viewport tracking:

```go
// Replace lines 40-48 in layout.go with:
lay.ChipW = 8
lay.ChipH = 4
lay.ChannelFrameH = lay.ChipH + 2

// Calculate how many chips fit in the available width
// Each chip takes ChipW + 1 (for border), plus 1 for left border
chipsPerRow := (totalW - 1) / (lay.ChipW + 1)
if chipsPerRow < 1 {
    chipsPerRow = 1
}
lay.ChipsPerRow = chipsPerRow
lay.ChipCount = n // keep total count (modules + ADD)
```

Add `ChipsPerRow int` field to the `Layout` struct (visible chips per row for scroll calculations).

- [ ] **Step 2: Add scroll state to ChannelStrip**

Add `scrollOffset int` field to `ChannelStrip`. When the selected chip is out of the visible range, adjust `scrollOffset`. Add methods `canScrollLeft()` and `canScrollRight()`.

In `ChannelStrip`:
```go
type ChannelStrip struct {
    // ... existing fields ...
    scrollOffset int
}
```

In `movePrev`/`moveNext`, after changing `cs.selected`, call `cs.ensureVisible(chipsPerRow)` to adjust scroll.

```go
func (cs *ChannelStrip) ensureVisible(visibleCount int) {
    if cs.selected < cs.scrollOffset {
        cs.scrollOffset = cs.selected
    }
    if cs.selected >= cs.scrollOffset+visibleCount {
        cs.scrollOffset = cs.selected - visibleCount + 1
    }
}
```

- [ ] **Step 3: Add scroll indicators in ChipContents**

Modify `ChipContents` to accept `visibleCount` and return only the visible slice. Add ◂/▸ indicators:

In `dashboard_view.go`, when calling `ChipContents`, pass `lay.ChipsPerRow` to determine visible range. Render ◂ before the frame if `scrollOffset > 0`, and ▸ after if more chips exist beyond visible range.

- [ ] **Step 4: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/tui/layout.go dot/internal/tui/panel_channel.go dot/internal/tui/dashboard_view.go
git commit -m "feat: fixed 8x4 chip size with horizontal scroll indicators"
```

---

### Task 8: Overview improvements (T3+T4+T5)

**Files:**
- Modify: `dot/internal/tui/panel_overview.go`
- Modify: `dot/internal/git/git.go`

- [ ] **Step 1: Add AheadBehind to git package**

Add `"fmt"` to the import block of `dot/internal/git/git.go`, then add:

```go
// AheadBehind returns the number of commits ahead and behind the upstream branch.
// Returns (0, 0, nil) if there is no upstream or on error.
func AheadBehind(repoDir string) (ahead, behind int, err error) {
	out, err := run(repoDir, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return 0, 0, nil // no upstream or error — degrade gracefully
	}
	out = strings.TrimSpace(out)
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, nil
	}
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)
	return ahead, behind, nil
}
```

- [ ] **Step 2: Update health formula in panel_overview.go**

First, add required imports to `panel_overview.go`: `"os"`, `"path/filepath"`, and `gitpkg "github.com/aporicho/dotfiles/dot/internal/git"`.

Replace `linkCounts()` method with `healthCounts()` that includes secrets:

```go
func (o *Overview) healthCounts() (total, healthy int) {
	for _, mod := range o.modules {
		// Count links
		for range mod.Links {
			total++
			if o.manifest.IsInstalled(mod.Name) {
				healthy++
			}
		}
		// Count secrets
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
```

Update `renderHealthBar` to call `o.healthCounts()` instead of `o.linkCounts()`.

- [ ] **Step 3: Add shell/pkg/font fields to renderSystemInfo**

Add after the existing `keychainLine`:

```go
// Shell
shellPath := os.Getenv("SHELL")
shellName := filepath.Base(shellPath)
if shellName == "" || shellName == "." {
    shellName = "unknown"
}
shellLine := dimmed.Render("shell  ") + cyan.Render(shellName)

// Package manager
pkgMgr := platform.PackageManager()
if pkgMgr == "" {
    pkgMgr = "none"
}
pkgLine := dimmed.Render("pkg    ") + cyan.Render(pkgMgr)

// Font (fixed value)
fontLine := dimmed.Render("font   ") + cyan.Render("JetBrainsMono NF")
```

Update the return to include all fields in order: branch, sync, os, shell, pkg, key, font.

- [ ] **Step 4: Update sync display to ahead/behind format**

Replace the sync status section:

```go
// Sync status
ahead, behind, _ := gitpkg.AheadBehind(dfPath)  // Need dfPath in Overview
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
```

Note: Overview needs access to `dfPath`. Add `dfPath string` field to `Overview` struct, pass it in `NewOverview`.

- [ ] **Step 5: Update NewOverview and callers**

Add `dfPath` parameter to `NewOverview` and update `dashboard.go` where `NewOverview` is called.

- [ ] **Step 6: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add dot/internal/tui/panel_overview.go dot/internal/git/git.go dot/internal/tui/dashboard.go
git commit -m "feat: overview improvements - health with secrets, shell/pkg/font fields, ahead/behind sync"
```

---

### Task 9: Terminal command history (T6)

**Files:**
- Modify: `dot/internal/tui/panel_terminal.go`

- [ ] **Step 1: Add history buffer to Terminal**

Add fields to `Terminal` struct:

```go
type Terminal struct {
    // ... existing fields ...
    history      []string
    historyIndex int // -1 means not browsing history
}
```

Initialize `historyIndex` to `-1` in `NewTerminal`.

- [ ] **Step 2: Update handleInputMode for ↑↓ history**

Add cases in `handleInputMode`:

```go
case tea.KeyUp:
    if len(t.history) > 0 {
        if t.historyIndex == -1 {
            t.historyIndex = len(t.history) - 1
        } else if t.historyIndex > 0 {
            t.historyIndex--
        }
        t.input = t.history[t.historyIndex]
    }
case tea.KeyDown:
    if t.historyIndex >= 0 {
        t.historyIndex++
        if t.historyIndex >= len(t.history) {
            t.historyIndex = -1
            t.input = ""
        } else {
            t.input = t.history[t.historyIndex]
        }
    }
```

- [ ] **Step 3: Record commands in history on Enter**

In the `tea.KeyEnter` case, before clearing input, append to history:

```go
case tea.KeyEnter:
    if t.input != "" {
        cmd := t.input
        t.history = append(t.history, cmd)
        t.historyIndex = -1
        t.input = ""
        t.inputMode = false
        return t, func() tea.Msg {
            return TerminalExecMsg{Input: cmd}
        }
    }
```

- [ ] **Step 4: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/tui/panel_terminal.go
git commit -m "feat: add terminal command history with up/down navigation"
```

---

### Task 10: Remove confirmation + Ctrl shortcuts (T7+T8)

**Files:**
- Modify: `dot/internal/tui/dashboard.go`
- Modify: `dot/internal/tui/panel_controls.go`

- [ ] **Step 1: Add confirmation state to Dashboard**

Add field to `Dashboard`:

```go
type Dashboard struct {
    // ... existing fields ...
    confirmRemove bool // true when waiting for Y/N confirmation
}
```

- [ ] **Step 2: Update handleKey for confirmation flow**

In `handleKey`, add confirmation state handling at the top (after terminal input check):

```go
// Handle remove confirmation mode
if d.confirmRemove {
    switch key {
    case "y", "Y":
        d.confirmRemove = false
        mod := d.channel.Selected()
        if mod != nil {
            d.executing = true
            d.controls.SetExecuting(true)
            d.controls.SetConfirming(false)
            return d, tea.Batch(
                func() tea.Msg { return CmdStartMsg{} },
                execUninstall(d.dfPath, mod.Name),
            )
        }
    case "n", "N", "esc":
        d.confirmRemove = false
        d.controls.SetConfirming(false)
    }
    return d, nil
}
```

Change the `"x"` case to enter confirmation mode instead of executing directly:

```go
case "x":
    mod := d.channel.Selected()
    if mod == nil {
        return d, nil
    }
    d.confirmRemove = true
    d.controls.SetConfirming(true)
    d.controls.SetConfirmName(mod.Name)
    return d, nil
```

- [ ] **Step 3: Add Ctrl shortcuts for terminal input mode**

In `handleKey`, after the terminal input mode check, add a new section that checks for Ctrl keys when terminal is focused and in input mode. Actually, we need to intercept Ctrl keys BEFORE forwarding to terminal.

Restructure the terminal input mode handling:

```go
// When terminal is focused and in input mode, check for Ctrl shortcuts first
if d.focus == focusTerminal && d.terminal.InputMode() {
    switch key {
    case "ctrl+left", "ctrl+h":
        _, cmd := d.channel.Update(tea.KeyMsg{Type: tea.KeyLeft})
        return d, cmd
    case "ctrl+right", "ctrl+l":
        _, cmd := d.channel.Update(tea.KeyMsg{Type: tea.KeyRight})
        return d, cmd
    case "ctrl+p":
        mod := d.channel.Selected()
        if mod == nil { return d, nil }
        d.executing = true
        d.controls.SetExecuting(true)
        return d, tea.Batch(
            func() tea.Msg { return CmdStartMsg{} },
            execInstall(d.dfPath, mod.Name),
        )
    case "ctrl+u":  // Ctrl+U for push (upload) — Ctrl+Shift+P is indistinguishable from Ctrl+P in terminals
        d.executing = true
        d.controls.SetExecuting(true)
        return d, tea.Batch(
            func() tea.Msg { return CmdStartMsg{} },
            execPush(d.dfPath, "tui push"),
        )
    case "ctrl+d":
        mod := d.channel.Selected()
        if mod == nil { return d, nil }
        d.executing = true
        d.controls.SetExecuting(true)
        return d, tea.Batch(
            func() tea.Msg { return CmdStartMsg{} },
            execDoctor(d.dfPath, mod.Name),
        )
    case "ctrl+x":
        mod := d.channel.Selected()
        if mod == nil { return d, nil }
        d.confirmRemove = true
        d.controls.SetConfirming(true)
        d.controls.SetConfirmName(mod.Name)
        return d, nil
    case "ctrl+a":
        d.terminal.AppendOutput("提示：请使用 dot add <module> 添加模块")
        return d, nil
    case "ctrl+q":
        return d, tea.Quit
    }
    // All other keys go to terminal
    _, cmd := d.terminal.Update(m)
    return d, cmd
}
```

- [ ] **Step 4: Add confirmation rendering to Controls**

Add fields and methods to `Controls`:

```go
type Controls struct {
    executing  bool
    confirming bool
    confirmName string
    focused    bool
    styles     Styles
    theme      Theme
}

func (c *Controls) SetConfirming(v bool) { c.confirming = v }
func (c *Controls) SetConfirmName(n string) { c.confirmName = n }
```

In `View`, add confirmation rendering before the buttons:

```go
if c.confirming {
    prompt := lipgloss.NewStyle().
        Foreground(lipgloss.Color(c.theme.Red())).
        Width(width).
        Align(lipgloss.Center).
        Render(fmt.Sprintf("确认卸载 %s？Y/N", c.confirmName))
    return prompt
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add dot/internal/tui/dashboard.go dot/internal/tui/panel_controls.go
git commit -m "feat: add uninstall confirmation flow and Ctrl shortcuts in terminal input mode"
```

---

### Task 11: Responsive degradation (T9)

**Files:**
- Modify: `dot/internal/tui/dashboard_view.go`
- Modify: `dot/internal/tui/layout.go`

- [ ] **Step 1: Add degradation flags to Layout**

Add fields to `Layout` struct:

```go
type Layout struct {
    // ... existing fields ...
    ShowFooter   bool
    ShowOverview bool
    ShowControls bool
}
```

In `ComputeLayout`, set these flags based on terminal size:

```go
lay.ShowFooter = totalH >= 24
lay.ShowOverview = totalW >= 100
lay.ShowControls = totalH >= 20
```

Adjust vertical budget calculations to account for hidden panels.

- [ ] **Step 2: Update View in dashboard_view.go**

Conditionally render panels based on layout flags:

```go
func (d *Dashboard) View() string {
    if d.width < 60 || d.height < 12 {
        return "终端过小，请调整窗口大小"
    }

    lay := ComputeLayout(d.width, d.height, len(d.modules))
    // ... border setup ...

    // Channel strip (always shown)
    // ...

    // Middle panels: conditionally include Overview
    if lay.ShowOverview {
        // 3-column: overview + scope + terminal
    } else {
        // 2-column: scope + terminal (overview hidden)
    }

    // Controls (conditionally)
    // Footer (conditionally)

    parts := []string{channelFrame, panelFrame}
    if lay.ShowControls {
        parts = append(parts, controlsView)
    }
    parts = append(parts, sep)
    if lay.ShowFooter {
        parts = append(parts, footer)
    }

    return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
```

- [ ] **Step 3: Adjust layout calculation for hidden panels**

When Overview is hidden, redistribute its width to Scope and Terminal. When Controls or Footer are hidden, add that height to panel area.

- [ ] **Step 4: Verify compilation**

Run: `cd /Users/aporicho/dotfiles/dot && go build ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dot/internal/tui/dashboard_view.go dot/internal/tui/layout.go
git commit -m "feat: add responsive panel degradation for small terminals"
```

---

### Task 12: Update design documents + final verification

**Files:**
- Modify: `docs/design/secrets.md`
- Modify: `docs/design/tui.md`

- [ ] **Step 1: Update secrets.md CLI references**

Replace references to `dot pull` with `dot install` for secrets handling.

- [ ] **Step 2: Update tui.md button labels**

Update operation bar section to show INSTALL/UNINSTALL instead of PULL/REMOVE.

- [ ] **Step 3: Final compilation and run**

Run: `cd /Users/aporicho/dotfiles/dot && go build -o dot && ./dot --help`
Expected: Shows updated commands: `install`, `uninstall`, `pull`, `push`, `add`, `remove`, `doctor`, `clean`, `status`.

- [ ] **Step 4: Commit**

```bash
git add docs/design/secrets.md docs/design/tui.md
git commit -m "docs: sync design documents with implementation alignment"
```
