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

func execPull(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot pull %s\n", modName)
		err := ops.PullModule(dfPath, modName, &buf)
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

func execRemove(dfPath, modName string) tea.Cmd {
	return func() tea.Msg {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$ dot remove %s\n", modName)
		err := ops.RemoveModule(dfPath, modName, &buf)
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
