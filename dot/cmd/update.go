package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dot CLI and config from GitHub",
	Long: `Download the latest pre-built dot CLI binary from GitHub Releases
and pull the latest dotfiles config.

This replaces the running dot binary with the latest version
and runs git pull to sync configuration files.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Step 1: determine platform and download URL
	arch := runtime.GOARCH
	osName := runtime.GOOS
	if arch == "aarch64" {
		arch = "arm64"
	}
	if arch == "x86_64" || arch == "amd64" {
		arch = "amd64"
	}

	label := fmt.Sprintf("%s-%s", osName, arch)
	downloadURL := fmt.Sprintf(
		"https://github.com/aporicho/dotfiles/releases/download/latest/dot-%s", label,
	)

	fmt.Printf("→ 检查更新: %s\n", label)

	// Step 2: find current binary path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		exe, _ = os.Executable()
	}

	// Step 3: download to temp
	tmpFile := filepath.Join(os.TempDir(), "dot-new")
	fmt.Printf("→ 下载最新版本...\n")

	resp, err := http.DefaultClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed (HTTP %d): no release binary for %s", resp.StatusCode, label)
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("download incomplete: %w", err)
	}
	out.Close()

	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("setting permissions: %w", err)
	}

	fmt.Printf("  已下载: %d bytes\n", written)

	// Step 4: verify it runs (sanity check)
	// Just check it's an executable by running --help
	// We skip this to keep it simple

	// Step 5: check if we can write to the target directory
	binDir := filepath.Dir(exe)
	dirWritable := false
	testFile := filepath.Join(binDir, ".dot-write-test")
	if err := os.WriteFile(testFile, []byte{}, 0644); err == nil {
		os.Remove(testFile)
		dirWritable = true
	}

	if !dirWritable {
		// Binary is in a protected directory (e.g., /usr/local/bin)
		// Save new binary to temp and instruct user
		fmt.Printf("  已下载到: %s\n", tmpFile)
		fmt.Printf("\n需要管理员权限来更新 dot（位于 %s）:\n", binDir)
		fmt.Printf("  sudo mv %s %s\n", tmpFile, exe)
		fmt.Printf("  sudo chmod +x %s\n", exe)
		fmt.Println("")
	} else {
		// Step 6: replace binary directly
		backupPath := exe + ".bak"
		os.Remove(backupPath)

		if err := os.Rename(exe, backupPath); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("backup failed: %w", err)
		}
		if err := os.Rename(tmpFile, exe); err != nil {
			os.Rename(backupPath, exe) // restore backup
			os.Remove(tmpFile)
			return fmt.Errorf("replace failed: %w", err)
		}
		os.Remove(backupPath)
		fmt.Printf("✓ dot 已更新: %s\n", exe)
	}

	// Step 7: git pull latest config
	dfPath, err := DotfilesPath()
	if err == nil {
		fmt.Printf("→ 同步配置...\n")
		if err := gitpkg.Pull(dfPath); err != nil {
			fmt.Printf("  配置同步跳过: %v\n", err)
		} else {
			fmt.Printf("✓ 配置已同步\n")
		}
	}

	fmt.Println("✓ 更新完成")
	return nil
}
