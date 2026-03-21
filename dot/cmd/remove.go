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

var removeCmd = &cobra.Command{
	Use:   "remove <modules...>",
	Short: "Uninstall modules (remove symlinks)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func init() { rootCmd.AddCommand(removeCmd) }

func runRemove(cmd *cobra.Command, args []string) error {
	// Step 1: load dotfiles path, all modules, build maps
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

	// Step 2: load manifest
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Step 3: process each requested module
	for _, name := range args {
		// 3a: check if installed
		if !mf.IsInstalled(name) {
			fmt.Printf("  ⊘ %s：未安装，跳过\n", name)
			continue
		}

		// 3b: check reverse dependencies (only installed ones)
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

		// 3c: run pre_remove hook
		if mod, ok := modMap[name]; ok {
			if err := hook.Run(mod.Hooks.PreRemove, mod.Dir); err != nil {
				fmt.Fprintf(os.Stderr, "  警告：%s pre_remove hook 失败：%v\n", name, err)
			}
		}

		// 3d: remove symlinks from manifest records
		if rec, ok := mf.Modules[name]; ok {
			for _, link := range rec.Links {
				if err := linker.RemoveLink(link.Target); err != nil {
					fmt.Fprintf(os.Stderr, "  警告：移除链接 %s 失败：%v\n", link.Target, err)
				}
				// Restore backup if present
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

		// 3e: run post_remove hook
		if mod, ok := modMap[name]; ok {
			if err := hook.Run(mod.Hooks.PostRemove, mod.Dir); err != nil {
				fmt.Fprintf(os.Stderr, "  警告：%s post_remove hook 失败：%v\n", name, err)
			}
		}

		// 3f: remove module from manifest
		mf.RemoveModule(name)

		// 3g: print success
		fmt.Printf("  ✓ %s：已卸载\n", name)
	}

	// Step 4: save manifest
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	return nil
}
