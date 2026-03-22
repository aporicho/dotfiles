package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up backup files created during module installation",
	RunE:  runClean,
}

func init() { rootCmd.AddCommand(cleanCmd) }

func runClean(cmd *cobra.Command, args []string) error {
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

	// ── Backup files ──────────────────────────────────────────────────────────
	// Collect backup files from two sources:
	// 1) Manifest link records with non-empty Backup fields
	// 2) Filesystem scan of common config directories

	backupFiles := make(map[string]struct{})

	// Source 1: manifest records
	for _, rec := range mf.Modules {
		for _, link := range rec.Links {
			if link.Backup != "" {
				backupFiles[link.Backup] = struct{}{}
			}
		}
	}

	// Source 2: scan common directories for *.backup.* files
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	scanDirs := []string{
		filepath.Join(home, ".config"),
		home,
		filepath.Join(home, "bin"),
	}

	for _, dir := range scanDirs {
		scanForBackupFiles(dir, backupFiles)
	}

	if len(backupFiles) == 0 {
		fmt.Println("没有找到备份文件")
	} else {
		// List all found backup files with sizes
		var totalSize int64
		var validFiles []string

		fmt.Println("找到以下备份文件：")
		for path := range backupFiles {
			info, err := os.Stat(path)
			if err != nil {
				// File referenced in manifest but no longer on disk — skip
				continue
			}
			size := info.Size()
			totalSize += size
			validFiles = append(validFiles, path)
			fmt.Printf("  %s (%s)\n", path, formatSize(size))
		}

		if len(validFiles) == 0 {
			fmt.Println("没有找到备份文件")
		} else {
			fmt.Printf("\n共 %d 个备份文件，总计 %s\n", len(validFiles), formatSize(totalSize))
			fmt.Print("确认删除？(Y/n) ")

			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "n" || answer == "no" {
				fmt.Println("取消操作")
			} else {
				// Delete all backup files
				deleted := 0
				for _, path := range validFiles {
					if err := os.Remove(path); err != nil {
						fmt.Fprintf(os.Stderr, "  删除失败: %s: %v\n", path, err)
						continue
					}
					deleted++
				}

				// Clear Backup fields in manifest records
				for _, rec := range mf.Modules {
					for i := range rec.Links {
						rec.Links[i].Backup = ""
					}
				}
				if err := manifest.Save(mf, mfPath); err != nil {
					return fmt.Errorf("saving manifest: %w", err)
				}

				fmt.Printf("✓ 已清理 %d 个备份文件\n", deleted)
			}
		}
	}

	// ── Orphan links ──────────────────────────────────────────────────────────
	var orphans []string
	for _, rec := range mf.Modules {
		for _, link := range rec.Links {
			if isOrphanLink(link.Target) {
				orphans = append(orphans, link.Target)
			}
		}
	}

	if len(orphans) > 0 {
		fmt.Println("\n找到以下孤立链接：")
		for _, target := range orphans {
			fmt.Printf("  %s\n", target)
		}
		fmt.Print("\n确认删除孤立链接？(Y/n) ")
		answer2, _ := reader.ReadString('\n')
		answer2 = strings.TrimSpace(strings.ToLower(answer2))
		if answer2 != "n" && answer2 != "no" {
			removed := 0
			for _, target := range orphans {
				if err := os.Remove(target); err == nil {
					removed++
				}
			}
			fmt.Printf("✓ 已清理 %d 个孤立链接\n", removed)
		} else {
			fmt.Println("跳过孤立链接清理")
		}
	}

	// ── Invalid manifest records ──────────────────────────────────────────────
	var invalidModules []string
	for name := range mf.Modules {
		modDir := filepath.Join(dfPath, "modules", name)
		if _, err := os.Stat(modDir); os.IsNotExist(err) {
			invalidModules = append(invalidModules, name)
		}
	}

	if len(invalidModules) > 0 {
		fmt.Println("\n找到以下无效的清单记录（模块目录已不存在）：")
		for _, name := range invalidModules {
			fmt.Printf("  %s\n", name)
		}
		fmt.Print("\n确认从清单中删除这些记录？(Y/n) ")
		answer3, _ := reader.ReadString('\n')
		answer3 = strings.TrimSpace(strings.ToLower(answer3))
		if answer3 != "n" && answer3 != "no" {
			for _, name := range invalidModules {
				mf.RemoveModule(name)
			}
			if err := manifest.Save(mf, mfPath); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}
			fmt.Printf("✓ 已清理 %d 条无效清单记录\n", len(invalidModules))
		} else {
			fmt.Println("跳过无效记录清理")
		}
	}

	return nil
}

// isOrphanLink reports whether target is missing, not a symlink, or points to
// a non-existent source.
func isOrphanLink(target string) bool {
	info, err := os.Lstat(target)
	if err != nil {
		// target does not exist at all
		return true
	}
	if info.Mode()&os.ModeSymlink == 0 {
		// exists but is not a symlink
		return true
	}
	// symlink exists — check that its destination exists
	dest, err := os.Readlink(target)
	if err != nil {
		return true
	}
	_, err = os.Stat(dest)
	return err != nil
}

// scanForBackupFiles scans a directory (non-recursively for home, recursively
// for subdirectories like .config) and adds any file whose name contains
// ".backup." to the result map.
func scanForBackupFiles(dir string, result map[string]struct{}) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return
	}

	home, _ := os.UserHomeDir()
	isHome := filepath.Clean(dir) == filepath.Clean(home)

	if isHome {
		// For home directory, only scan top-level entries (no recursion)
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.Contains(entry.Name(), ".backup.") {
				result[filepath.Join(dir, entry.Name())] = struct{}{}
			}
		}
	} else {
		// For other directories, walk recursively
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			if strings.Contains(d.Name(), ".backup.") {
				result[path] = struct{}{}
			}
			return nil
		})
	}
}
