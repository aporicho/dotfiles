// Package sysdep installs system-level dependencies via the platform package manager.
package sysdep

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/aporicho/dotfiles/dot/internal/module"
)

// Install installs packages for the current platform.
// On darwin it runs "brew install <packages>"; on linux it runs
// "sudo apt-get install -y <packages>". An empty package list is a no-op.
func Install(darwin, linux module.PlatformDeps) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = buildCommand("brew", darwin.Brew)
	case "linux":
		cmd = buildCommand("sudo", linux.Apt)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if cmd == nil {
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildCommand creates an exec.Cmd for the given package manager and packages.
// It returns nil when there are no packages to install.
func buildCommand(pm string, packages []string) *exec.Cmd {
	if len(packages) == 0 {
		return nil
	}

	var args []string
	switch pm {
	case "brew":
		args = append([]string{"install"}, packages...)
	case "sudo":
		args = append([]string{"apt-get", "install", "-y"}, packages...)
	default:
		args = append([]string{"install"}, packages...)
	}

	return exec.Command(pm, args...)
}
