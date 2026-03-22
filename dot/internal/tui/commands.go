package tui

import (
	"bytes"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
	"github.com/aporicho/dotfiles/dot/internal/ops"
)

func execInstall(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot install %s\n", modName)
		err := ops.InstallModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execPull(dfPath string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot pull\n")
		err := ops.PullRepo(dfPath, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execPush(dfPath, msg string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot push\n")
		err := ops.PushChanges(dfPath, msg, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execDoctor(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot doctor %s\n", modName)
		err := ops.DoctorCheck(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func execUninstall(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot uninstall %s\n", modName)
		err := ops.UninstallModule(dfPath, modName, &buf)
		return CmdOutputMsg{Output: buf.String(), Err: err}
	}
}

func reloadData(dfPath string) tea.Cmd {
	return func() tea.Msg {
		modules, _ := module.LoadAll(filepath.Join(dfPath, "modules"))
		mfPath, _ := manifest.DefaultPath()
		mf, _ := manifest.Load(mfPath)
		gitChanges, _ := gitpkg.Status(dfPath)
		return DataReloadMsg{Modules: modules, Manifest: mf, GitChanges: gitChanges}
	}
}
