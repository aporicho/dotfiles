package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/dotfiles/dot/internal/module"
)

// dispatchCommand parses terminal input and returns commands to execute and
// messages to broadcast. If errorText is non-empty, it should be shown as a
// TerminalHintMsg.
func dispatchCommand(input string, dfPath string, selectedMod *module.Module) (cmds []tea.Cmd, msgs []tea.Msg, errorText string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil, ""
	}
	cmd, args := parts[0], parts[1:]

	switch cmd {
	case "install":
		name := resolveModName(args, selectedMod)
		if name == "" {
			return nil, nil, "错误：未指定模块"
		}
		return []tea.Cmd{execInstall(dfPath, name)}, []tea.Msg{CmdStartMsg{}}, ""

	case "pull":
		return []tea.Cmd{execPull(dfPath)}, []tea.Msg{CmdStartMsg{}}, ""

	case "push":
		msg := "tui push"
		if len(args) > 0 {
			msg = strings.Join(args, " ")
		}
		return []tea.Cmd{execPush(dfPath, msg)}, []tea.Msg{CmdStartMsg{}}, ""

	case "doctor":
		name := resolveModName(args, selectedMod)
		if name == "" {
			return nil, nil, "错误：未指定模块"
		}
		return []tea.Cmd{execDoctor(dfPath, name)}, []tea.Msg{CmdStartMsg{}}, ""

	case "uninstall":
		name := resolveModName(args, selectedMod)
		if name == "" {
			return nil, nil, "错误：未指定模块"
		}
		return []tea.Cmd{execUninstall(dfPath, name)}, []tea.Msg{CmdStartMsg{}}, ""

	case "remove":
		return nil, nil, "请使用 dot remove <module> 命令行操作（不可逆操作，TUI 不支持）"

	default:
		return nil, nil, "未知命令: " + input
	}
}

// resolveModName picks a module name from explicit args or the current selection.
func resolveModName(args []string, selected *module.Module) string {
	if len(args) > 0 {
		return args[0]
	}
	if selected != nil {
		return selected.Name
	}
	return ""
}
