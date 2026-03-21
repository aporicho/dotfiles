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
	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

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
		return nil
	}

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
		return nil
	}

	fmt.Printf("\n共 %d 个备份文件，总计 %s\n", len(validFiles), formatSize(totalSize))
	fmt.Print("确认删除？(Y/n) ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "n" || answer == "no" {
		fmt.Println("取消操作")
		return nil
	}

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
	return nil
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
