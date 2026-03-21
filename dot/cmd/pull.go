package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/deps"
	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/hook"
	"github.com/aporicho/dotfiles/dot/internal/linker"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
	"github.com/aporicho/dotfiles/dot/internal/sysdep"
	"github.com/aporicho/dotfiles/dot/internal/tui"
)

var pullAll bool

var pullCmd = &cobra.Command{
	Use:   "pull [modules...]",
	Short: "Install or update modules",
	RunE:  runPull,
}

func init() {
	pullCmd.Flags().BoolVar(&pullAll, "all", false, "Install all available modules")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	// Step 1: git pull (non-fatal for offline use)
	if err := gitpkg.Pull(dfPath); err != nil {
		fmt.Fprintf(os.Stderr, "警告：git pull 失败（离线？）：%v\n", err)
	}

	// Step 2: load modules, filter by platform
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

	// Step 3: determine requested modules
	var requested []string
	if pullAll {
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

	// Step 4: resolve install order
	order, err := deps.Resolve(requested, depsMap)
	if err != nil {
		return fmt.Errorf("dependency resolution: %w", err)
	}

	// Step 5: load manifest
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}
	mf.DotfilesPath = dfPath

	// Step 6: install each module
	for _, name := range order {
		mod := modMap[name]
		fmt.Printf("→ 安装 %s ...\n", name)
		if err := installModule(dfPath, mod, mf); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
		fmt.Printf("  ✓ %s 安装完成\n", name)
	}

	// Step 7: save manifest
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Println("全部完成。")
	return nil
}

// installModule runs hooks, installs deps, creates symlinks, and records in manifest.
// On failure it rolls back created links.
func installModule(dfPath string, mod *module.Module, mf *manifest.Manifest) error {
	modDir := mod.Dir
	var createdLinks []*linker.LinkResult

	// rollback helper
	rollback := func() {
		for _, lr := range createdLinks {
			_ = linker.RemoveLink(lr.Target)
			if lr.BackupPath != "" {
				_ = os.Rename(lr.BackupPath, lr.Target)
			}
		}
	}

	// Step 1: pre_install hook
	if err := hook.Run(mod.Hooks.PreInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("pre_install hook: %w", err)
	}

	// Step 2: system deps (warn on failure, don't abort)
	if err := sysdep.Install(mod.Deps.Darwin, mod.Deps.Linux); err != nil {
		fmt.Fprintf(os.Stderr, "  警告：系统依赖安装失败：%v\n", err)
	}

	// Step 3: create links
	for _, link := range mod.Links {
		// Skip links not matching current platform
		if !platform.MatchesPlatform(link.Platforms) {
			continue
		}

		target := expandHome(link.Target)

		if link.Source == "." {
			// Expand directory entries
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

	// Step 4: post_install hook
	if err := hook.Run(mod.Hooks.PostInstall, modDir); err != nil {
		rollback()
		return fmt.Errorf("post_install hook: %w", err)
	}

	// Step 5: record in manifest
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
