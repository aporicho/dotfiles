package tui

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// focusable identifies which panel currently has focus.
type focusable int

const (
	focusChannel focusable = iota
	focusOverview
	focusScope
	focusTerminal
	focusControls
	focusCount // sentinel for wrapping
)

// Dashboard is the top-level bubbletea Model that composes all five panels.
type Dashboard struct {
	channel  *ChannelStrip
	overview *Overview
	scope    *Scope
	terminal *Terminal
	controls *Controls

	focus      focusable
	dfPath     string
	modules    []*module.Module
	manifest   *manifest.Manifest
	gitChanges []string
	executing  bool
	width      int
	height     int
	styles     Styles
	theme      Theme
}

// RunDashboard is the public entry point that loads data and runs the TUI.
func RunDashboard(dfPath string) error {
	modules, err := module.LoadAll(filepath.Join(dfPath, "modules"))
	if err != nil {
		modules = nil // degrade gracefully with no modules
	}

	mfPath, err := manifest.DefaultPath()
	if err != nil {
		return err
	}
	mf, err := manifest.Load(mfPath)
	if err != nil {
		return err
	}

	gitChanges, _ := gitpkg.Status(dfPath)

	theme := DetectTheme()
	styles := NewStyles(theme)

	var firstMod *module.Module
	if len(modules) > 0 {
		firstMod = modules[0]
	}

	ch := NewChannelStrip(modules, mf, gitChanges, styles, theme)
	ov := NewOverview(modules, mf, gitChanges, styles, theme, dfPath)
	sc := NewScope(firstMod, mf, gitChanges, styles, theme)
	tm := NewTerminal(styles, theme)
	ct := NewControls(styles, theme)

	// Channel strip starts focused.
	ch.SetFocus(true)

	d := &Dashboard{
		channel:    ch,
		overview:   ov,
		scope:      sc,
		terminal:   tm,
		controls:   ct,
		focus:      focusChannel,
		dfPath:     dfPath,
		modules:    modules,
		manifest:   mf,
		gitChanges: gitChanges,
		styles:     styles,
		theme:      theme,
	}

	p := tea.NewProgram(d, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Init implements tea.Model. Fires the initial module selection.
func (d *Dashboard) Init() tea.Cmd {
	if len(d.modules) == 0 {
		return nil
	}
	mod := d.channel.Selected()
	idx := d.channel.SelectedIndex()
	return func() tea.Msg {
		return ModuleSelectedMsg{Index: idx, Module: mod}
	}
}

// Update implements tea.Model.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = m.Width
		d.height = m.Height
		return d, nil

	case tea.KeyMsg:
		return d.handleKey(m)

	case ModuleSelectedMsg:
		var cmd tea.Cmd
		_, cmd = d.scope.Update(m)
		return d, cmd

	case CmdStartMsg:
		d.executing = true
		_, cmd := d.controls.Update(m)
		return d, cmd

	case CmdOutputMsg:
		d.executing = false
		var cmds []tea.Cmd
		_, cmd1 := d.terminal.Update(m)
		cmds = append(cmds, cmd1)
		_, cmd2 := d.controls.Update(m)
		cmds = append(cmds, cmd2)
		cmds = append(cmds, reloadData(d.dfPath))
		return d, tea.Batch(cmds...)

	case DataReloadMsg:
		d.modules = m.Modules
		d.manifest = m.Manifest
		d.gitChanges = m.GitChanges
		// Refresh the channel strip with new data.
		d.channel = NewChannelStrip(m.Modules, m.Manifest, m.GitChanges, d.styles, d.theme)
		d.channel.SetFocus(d.focus == focusChannel)
		d.overview.Update(m)
		d.scope.Update(m)
		return d, nil

	case TerminalExecMsg:
		return d.handleTerminalExec(m)
	}

	return d, nil
}

// handleKey routes key messages based on focus and state.
func (d *Dashboard) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := m.String()

	// When terminal is focused and in input mode, forward everything to terminal.
	if d.focus == focusTerminal && d.terminal.InputMode() {
		_, cmd := d.terminal.Update(m)
		return d, cmd
	}

	// When executing, only allow quit.
	if d.executing {
		if key == "q" || key == "ctrl+c" {
			return d, tea.Quit
		}
		return d, nil
	}

	switch key {
	case "q", "ctrl+c":
		return d, tea.Quit

	case "left", "h", "right", "l":
		_, cmd := d.channel.Update(m)
		return d, cmd

	case "tab":
		d.cycleFocus()
		return d, nil

	case "p":
		mod := d.channel.Selected()
		if mod == nil {
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, tea.Batch(
			func() tea.Msg { return CmdStartMsg{} },
			execInstall(d.dfPath, mod.Name),
		)

	case "P":
		d.executing = true
		d.controls.SetExecuting(true)
		return d, tea.Batch(
			func() tea.Msg { return CmdStartMsg{} },
			execPush(d.dfPath, "tui push"),
		)

	case "d":
		mod := d.channel.Selected()
		if mod == nil {
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, tea.Batch(
			func() tea.Msg { return CmdStartMsg{} },
			execDoctor(d.dfPath, mod.Name),
		)

	case "x":
		mod := d.channel.Selected()
		if mod == nil {
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, tea.Batch(
			func() tea.Msg { return CmdStartMsg{} },
			execUninstall(d.dfPath, mod.Name),
		)

	case "a":
		d.terminal.AppendOutput("提示：请使用 dot add <module> 添加模块")
		return d, nil

	case ":":
		d.setFocus(focusTerminal)
		_, cmd := d.terminal.Update(m)
		return d, cmd

	case "esc":
		// Return focus to channel strip.
		d.setFocus(focusChannel)
		return d, nil
	}

	return d, nil
}

// handleTerminalExec parses a terminal command and dispatches the appropriate action.
func (d *Dashboard) handleTerminalExec(m TerminalExecMsg) (tea.Model, tea.Cmd) {
	parts := strings.Fields(m.Input)
	if len(parts) == 0 {
		return d, nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "install":
		modName := ""
		if len(args) > 0 {
			modName = args[0]
		} else if mod := d.channel.Selected(); mod != nil {
			modName = mod.Name
		}
		if modName == "" {
			d.terminal.AppendOutput("错误：未指定模块")
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, execInstall(d.dfPath, modName)

	case "pull":
		d.executing = true
		d.controls.SetExecuting(true)
		return d, execPull(d.dfPath)

	case "push":
		msg := "tui push"
		if len(args) > 0 {
			msg = strings.Join(args, " ")
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, execPush(d.dfPath, msg)

	case "doctor":
		modName := ""
		if len(args) > 0 {
			modName = args[0]
		} else if mod := d.channel.Selected(); mod != nil {
			modName = mod.Name
		}
		if modName == "" {
			d.terminal.AppendOutput("错误：未指定模块")
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, execDoctor(d.dfPath, modName)

	case "uninstall":
		modName := ""
		if len(args) > 0 {
			modName = args[0]
		} else if mod := d.channel.Selected(); mod != nil {
			modName = mod.Name
		}
		if modName == "" {
			d.terminal.AppendOutput("错误：未指定模块")
			return d, nil
		}
		d.executing = true
		d.controls.SetExecuting(true)
		return d, execUninstall(d.dfPath, modName)

	case "remove":
		d.terminal.AppendOutput("请使用 dot remove <module> 命令行操作（不可逆操作，TUI 不支持）")
		return d, nil

	default:
		d.terminal.AppendOutput("未知命令: " + m.Input)
		return d, nil
	}
}

// cycleFocus moves focus to the next panel in order.
func (d *Dashboard) cycleFocus() {
	next := (d.focus + 1) % focusCount
	d.setFocus(next)
}

// setFocus switches focus to the given panel, unfocusing the previous one.
func (d *Dashboard) setFocus(f focusable) {
	d.panelAt(d.focus).SetFocus(false)
	d.focus = f
	d.panelAt(d.focus).SetFocus(true)
}

// panelAt returns the Panel for the given focusable index.
func (d *Dashboard) panelAt(f focusable) Panel {
	switch f {
	case focusChannel:
		return d.channel
	case focusOverview:
		return d.overview
	case focusScope:
		return d.scope
	case focusTerminal:
		return d.terminal
	case focusControls:
		return d.controls
	default:
		return d.channel
	}
}
