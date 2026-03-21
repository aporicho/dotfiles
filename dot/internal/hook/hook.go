package hook

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a shell command in the given directory.
// Empty or whitespace-only commands are no-ops (return nil).
// Hook output (stdout/stderr) is forwarded to the calling process
// so the user can see it in real time.
func Run(command, dir string) error {
	if strings.TrimSpace(command) == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q: %w", command, err)
	}
	return nil
}
