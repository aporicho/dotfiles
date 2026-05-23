package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aporicho/dotfiles/dot/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open the dotfiles TUI dashboard",
	Long:  "Launch the interactive terminal UI for managing dotfiles modules, viewing status, and running installs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dfPath, err := DotfilesPath()
		if err != nil {
			return err
		}
		return tui.RunDashboard(dfPath)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
