package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/secrets"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check symlink health for installed modules",
	RunE:  runDoctor,
}

func init() { rootCmd.AddCommand(doctorCmd) }

func runDoctor(cmd *cobra.Command, args []string) error {
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

	if len(mf.Modules) == 0 {
		fmt.Println("没有已安装的模块")
		return nil
	}

	fmt.Println("检查符号链接健康状态：")

	var brokenModules []string
	totalIssues := 0

	for modName, rec := range mf.Modules {
		var issues []string

		for _, link := range rec.Links {
			target := link.Target
			expectedSource := link.Source
			if !filepath.IsAbs(expectedSource) {
				expectedSource = filepath.Join(dfPath, expectedSource)
			}

			// Check if target exists
			info, err := os.Lstat(target)
			if err != nil {
				if os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("    ✗ %s：链接不存在", target))
				} else {
					issues = append(issues, fmt.Sprintf("    ✗ %s：无法访问: %v", target, err))
				}
				continue
			}

			// Check if it's a symlink
			if info.Mode()&os.ModeSymlink == 0 {
				issues = append(issues, fmt.Sprintf("    ✗ %s：不是符号链接（已被替换为普通文件）", target))
				continue
			}

			// Check if it points to the correct source
			actual, err := os.Readlink(target)
			if err != nil {
				issues = append(issues, fmt.Sprintf("    ✗ %s：无法读取链接目标: %v", target, err))
				continue
			}

			// Resolve to absolute path for comparison
			if !filepath.IsAbs(actual) {
				actual = filepath.Join(filepath.Dir(target), actual)
			}

			actualClean := filepath.Clean(actual)
			expectedClean := filepath.Clean(expectedSource)

			if actualClean != expectedClean {
				issues = append(issues, fmt.Sprintf("    ✗ %s：指向 %s（应为 %s）", target, actualClean, expectedClean))
				continue
			}

			// Check if the source file actually exists
			if _, err := os.Stat(expectedSource); err != nil {
				issues = append(issues, fmt.Sprintf("    ✗ %s：源文件不存在 %s", target, expectedSource))
				continue
			}
		}

		// Check secrets
		mod, parseErr := module.Parse(filepath.Join(dfPath, "modules", modName))
		if parseErr == nil && len(mod.Secrets) > 0 {
			for _, sec := range mod.Secrets {
				plainPath := filepath.Join(mod.Dir, sec.Source)
				encPath := filepath.Join(mod.Dir, sec.Encrypted)
				target := expandHomePath(sec.Target)

				if _, err := os.Stat(encPath); os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("    ✗ %s 加密文件不存在", sec.Encrypted))
					continue
				}

				if _, err := os.Stat(plainPath); os.IsNotExist(err) {
					issues = append(issues, fmt.Sprintf("    ✗ %s 未解密（运行 dot pull %s）", sec.Source, modName))
					continue
				}

				info, _ := os.Stat(plainPath)
				if info.Mode().Perm() != 0o600 {
					issues = append(issues, fmt.Sprintf("    ✗ %s 权限为 %o，应为 0600", sec.Source, info.Mode().Perm()))
				} else {
					fmt.Printf("    ✓ %s 已解密且权限正确\n", sec.Source)
				}

				linkInfo, err := os.Lstat(target)
				if err != nil {
					issues = append(issues, fmt.Sprintf("    ✗ %s 符号链接不存在", target))
				} else if linkInfo.Mode()&os.ModeSymlink == 0 {
					issues = append(issues, fmt.Sprintf("    ✗ %s 不是符号链接", target))
				} else {
					fmt.Printf("    ✓ %s → %s\n", sec.Source, target)
				}

				// Content consistency check
				if secrets.KeychainAvailable() {
					if p, _ := secrets.LoadPassphrase(); p != "" {
						changed, err := secrets.HasChanged(plainPath, encPath, p)
						if err == nil && changed {
							issues = append(issues, fmt.Sprintf("    ✗ %s 与 %s 内容不一致（运行 dot push 同步）", sec.Source, sec.Encrypted))
						} else if err == nil && !changed {
							fmt.Printf("    ✓ %s 与 %s 内容一致\n", sec.Source, sec.Encrypted)
						}
					}
				}
			}
		}

		if len(issues) == 0 {
			fmt.Printf("  %s：✓ 正常\n", modName)
		} else {
			fmt.Printf("  %s：发现 %d 个问题\n", modName, len(issues))
			for _, issue := range issues {
				fmt.Println(issue)
			}
			brokenModules = append(brokenModules, modName)
			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		fmt.Printf("\n发现 %d 个问题。运行以下命令修复：\n", totalIssues)
		fmt.Printf("  dot pull %s\n", strings.Join(brokenModules, " "))
	} else {
		fmt.Println("\n所有模块链接正常。")
	}

	return nil
}

func expandHomePath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
