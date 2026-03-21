package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

var addName string

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add local config to dotfiles management",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "Module name (required for single files)")
	rootCmd.AddCommand(addCmd)
}

// moduleToml is used for TOML encoding of the generated module.toml.
type moduleToml struct {
	Name        string      `toml:"name"`
	Description string      `toml:"description"`
	Exclude     []string    `toml:"exclude,omitempty"`
	Links       []linkEntry `toml:"links"`
}

type linkEntry struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

func runAdd(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	// Step 1: resolve absolute path, check existence
	srcPath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("解析路径失败: %w", err)
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("路径不存在: %s", srcPath)
	}

	isDir := info.IsDir()

	// Step 2: determine module name
	modName := addName
	if isDir {
		if modName == "" {
			modName = filepath.Base(srcPath)
		}
	} else {
		if modName == "" {
			return fmt.Errorf("添加单个文件时必须指定 --name 参数")
		}
	}

	// Step 3: check module doesn't already exist
	modDir := filepath.Join(dfPath, "modules", modName)
	if _, err := os.Stat(modDir); err == nil {
		return fmt.Errorf("模块 %q 已存在: %s", modName, modDir)
	}

	// Step 4: list files and ask confirmation
	fmt.Printf("将添加到模块 %q：\n", modName)
	if isDir {
		err := filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(srcPath, path)
			if rel == "." {
				return nil
			}
			if d.IsDir() {
				fmt.Printf("  %s/\n", rel)
			} else {
				fi, _ := d.Info()
				fmt.Printf("  %s (%s)\n", rel, formatSize(fi.Size()))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("遍历目录失败: %w", err)
		}
	} else {
		fmt.Printf("  %s (%s)\n", filepath.Base(srcPath), formatSize(info.Size()))
	}

	fmt.Print("\n确认添加？(Y/n) ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "n" || answer == "no" {
		fmt.Println("取消操作")
		return nil
	}

	// Step 5: create module directory and copy files
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		return fmt.Errorf("创建模块目录失败: %w", err)
	}

	if isDir {
		if err := copyDir(srcPath, modDir); err != nil {
			return fmt.Errorf("复制目录失败: %w", err)
		}
	} else {
		dstFile := filepath.Join(modDir, filepath.Base(srcPath))
		if err := copyFile(srcPath, dstFile); err != nil {
			return fmt.Errorf("复制文件失败: %w", err)
		}
	}

	// Step 6: generate module.toml
	home, _ := os.UserHomeDir()
	targetPath := srcPath
	if strings.HasPrefix(targetPath, home) {
		targetPath = "~" + targetPath[len(home):]
	}

	mt := moduleToml{
		Name:        modName,
		Description: modName + " configuration",
	}

	if isDir {
		mt.Exclude = []string{"module.toml", "README*"}
		mt.Links = []linkEntry{{Source: ".", Target: targetPath}}
	} else {
		mt.Links = []linkEntry{{Source: filepath.Base(srcPath), Target: targetPath}}
	}

	tomlPath := filepath.Join(modDir, "module.toml")
	f, err := os.Create(tomlPath)
	if err != nil {
		return fmt.Errorf("创建 module.toml 失败: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(mt); err != nil {
		return fmt.Errorf("写入 module.toml 失败: %w", err)
	}

	fmt.Printf("已创建模块：%s\n", modDir)

	// Step 7: parse the new module and install symlinks
	mod, err := module.Parse(modDir)
	if err != nil {
		return fmt.Errorf("解析新模块失败: %w", err)
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

	fmt.Printf("→ 安装 %s ...\n", modName)
	if err := installModule(dfPath, mod, mf); err != nil {
		return fmt.Errorf("installing %s: %w", modName, err)
	}
	fmt.Printf("  ✓ %s 安装完成\n", modName)

	// Step 8: save manifest
	if err := manifest.Save(mf, mfPath); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	return nil
}

// copyDir recursively copies the contents of src into dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// formatSize returns a human-readable file size string.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
