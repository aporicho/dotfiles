package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest changes from remote (git pull only)",
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	dfPath, err := DotfilesPath()
	if err != nil {
		return err
	}

	if err := gitpkg.Pull(dfPath); err != nil {
		fmt.Fprintf(os.Stderr, "警告：git pull 失败（离线？）：%v\n", err)
		return err
	}

	fmt.Println("已同步最新更改。")
	return nil
}
