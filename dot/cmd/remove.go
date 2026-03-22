package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/ops"
)

var removeCmd = &cobra.Command{
	Use:   "remove <modules...>",
	Short: "Permanently delete module files from modules/",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func init() { rootCmd.AddCommand(removeCmd) }

func runRemove(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	// Load manifest to check installation status
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
		// Check module is not currently installed
		if mf.IsInstalled(name) {
			fmt.Fprintf(os.Stderr, "  ⚠ %s：当前已安装，请先运行 dot uninstall %s\n", name, name)
			continue
		}

		// Confirm deletion with user
		fmt.Printf("  警告：即将永久删除模块 %q 的所有文件，此操作不可撤销。\n", name)
		fmt.Print("  确认删除？(y/N) ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" {
			fmt.Printf("  跳过 %s\n", name)
			continue
		}

		// Delete module files
		if err := ops.RemoveModuleFiles(dfPath, name, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "  错误：%v\n", err)
		}
	}

	return nil
}
