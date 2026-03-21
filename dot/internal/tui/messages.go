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
