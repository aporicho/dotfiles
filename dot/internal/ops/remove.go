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

// RemoveModule uninstalls a single named module without interactive confirmation.
// It writes progress output to w.
func RemoveModule(dfPath, modName string, w io.Writer) error {
	// Step 1: load all modules and build maps
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

	// Step 2: load manifest
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	// Step 3: check module is installed
	if !mf.IsInstalled(modName) {
		return fmt.Errorf("模块 %q 未安装", modName)
	}

	// Step 4: check reverse dependencies — error if another installed module depends on this one
	revDeps := deps.ReverseDeps(modName, depsMap)
	var installedRevDeps []string
	for _, rd := range revDeps {
		if mf.IsInstalled(rd) {
			installedRevDeps = append(installedRevDeps, rd)
		}
	}
	if len(installedRevDeps) > 0 {
		return fmt.Errorf("模块 %q 被以下已安装模块依赖：%s", modName, strings.Join(installedRevDeps, ", "))
	}

	// Step 5: run pre_remove hook
	if mod, ok := modMap[modName]; ok {
		if err := hook.Run(mod.Hooks.PreRemove, mod.Dir); err != nil {
			fmt.Fprintf(w, "  警告：%s pre_remove hook 失败：%v\n", modName, err)
		}
	}

	// Step 6: remove symlinks and restore backups
	if rec, ok := mf.Modules[modName]; ok {
		for _, link := range rec.Links {
			if err := linker.RemoveLink(link.Target); err != nil {
				fmt.Fprintf(w, "  警告：移除链接 %s 失败：%v\n", link.Target, err)
			}
			// Restore backup if present
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

	// Step 7: run post_remove hook
	if mod, ok := modMap[modName]; ok {
		if err := hook.Run(mod.Hooks.PostRemove, mod.Dir); err != nil {
			fmt.Fprintf(w, "  警告：%s post_remove hook 失败：%v\n", modName, err)
		}
	}

	// Step 8: remove from manifest
	mf.RemoveModule(modName)

	// Step 9: save manifest
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Fprintf(w, "  ✓ %s：已卸载\n", modName)

	return nil
}
