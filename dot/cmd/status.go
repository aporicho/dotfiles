package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/platform"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show module installation status",
	RunE:  runStatus,
}

func init() { rootCmd.AddCommand(statusCmd) }

func runStatus(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	modules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		return fmt.Errorf("loading modules: %w", err)
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	gitChanges, _ := gitpkg.Status(dfPath) // non-fatal

	fmt.Println("模块状态：")
	for _, mod := range modules {
		if !platform.MatchesPlatform(mod.Platforms) {
			fmt.Printf("  %-12s ⊘ 不适用\n", mod.Name)
			continue
		}

		if !mf.IsInstalled(mod.Name) {
			fmt.Printf("  %-12s ✗ 未安装\n", mod.Name)
			continue
		}

		changes := detectChanges(dfPath, mf, mod.Name, gitChanges)
		if changes == "" {
			fmt.Printf("  %-12s ✓ 已安装 · 无变更\n", mod.Name)
		} else {
			fmt.Printf("  %-12s ✓ 已安装 · %s\n", mod.Name, changes)
		}
	}

	return nil
}

// detectChanges checks symlink health and git changes for a module.
// It returns a description string or empty if everything is clean.
func detectChanges(dfPath string, mf *manifest.Manifest, modName string, gitChanges []string) string {
	var issues []string

	// Check symlink health
	rec, ok := mf.Modules[modName]
	if ok {
		for _, link := range rec.Links {
			target := link.Target
			info, err := os.Lstat(target)
			if err != nil {
				if os.IsNotExist(err) {
					issues = append(issues, filepath.Base(target)+" 链接丢失")
				}
				continue
			}

			if info.Mode()&os.ModeSymlink != 0 {
				// It's a symlink — check if the destination exists
				dest, err := os.Readlink(target)
				if err == nil {
					if _, err := os.Stat(dest); os.IsNotExist(err) {
						issues = append(issues, filepath.Base(target)+" 链接已断")
					}
				}
			} else {
				// Regular file replaced the symlink
				issues = append(issues, filepath.Base(target)+" 已被替换为普通文件")
			}
		}
	}

	// Check git changes for files under modules/<modName>/
	prefix := filepath.Join("modules", modName) + "/"
	for _, line := range gitChanges {
		// git status --porcelain output: "XY filename" or "XY filename -> newname"
		if len(line) < 4 {
			continue
		}
		filePath := strings.TrimSpace(line[3:])
		// Handle renames: "old -> new"
		if idx := strings.Index(filePath, " -> "); idx >= 0 {
			filePath = filePath[idx+4:]
		}
		if strings.HasPrefix(filePath, prefix) {
			relName := strings.TrimPrefix(filePath, prefix)
			issues = append(issues, relName+" 已修改")
		}
	}

	return strings.Join(issues, ", ")
}
