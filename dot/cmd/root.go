package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dot",
	Short: "Dotfiles module manager",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// DotfilesPath returns the root path of the dotfiles repository.
// It walks up from the executable location looking for a modules/ directory.
// Falls back to ~/dotfiles if not found.
func DotfilesPath() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		for {
			if info, err := os.Stat(filepath.Join(dir, "modules")); err == nil && info.IsDir() {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dfPath := filepath.Join(home, "dotfiles")
	if _, err := os.Stat(filepath.Join(dfPath, "modules")); err != nil {
		return "", fmt.Errorf("dotfiles not found at %s (no modules/ directory)", dfPath)
	}
	return dfPath, nil
}
