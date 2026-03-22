package ops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
)

// PullRepo runs git pull on the dotfiles repository.
// It writes progress output to w.
func PullRepo(dfPath string, w io.Writer) error {
	fmt.Fprintf(w, "$ git pull\n")
	if err := gitpkg.Pull(dfPath); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	fmt.Fprintf(w, "已同步最新更改。\n")
	return nil
}

// expandHomeOps replaces a leading "~/" with the user's home directory.
func expandHomeOps(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
