package tui

import (
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// ModuleSelectedMsg is sent when the user selects a module in the Channel Strip.
type ModuleSelectedMsg struct {
	Index  int
	Module *module.Module
}

// CmdStartMsg signals a command has started executing.
type CmdStartMsg struct{}

// CmdOutputMsg carries command output back to the dashboard.
type CmdOutputMsg struct {
	Output string
	Err    error
}

// DataReloadMsg carries refreshed data after a command completes.
type DataReloadMsg struct {
	Modules    []*module.Module
	Manifest   *manifest.Manifest
	GitChanges []string
}

// ConfirmStartMsg signals the dashboard entered remove-confirmation mode.
type ConfirmStartMsg struct{ ModuleName string }

// ConfirmCancelMsg signals the dashboard exited remove-confirmation mode.
type ConfirmCancelMsg struct{}

// TerminalHintMsg carries a hint/info line to display in the terminal panel.
type TerminalHintMsg struct{ Text string }

// TerminalExecMsg is sent when the user submits a command in the terminal panel.
type TerminalExecMsg struct{ Input string }
