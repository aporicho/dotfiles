package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/dotfiles/dot/internal/manifest"
	"github.com/aporicho/dotfiles/dot/internal/module"
)

// --- Key-to-action mappings ---

// normalKeyMap maps keys in normal mode to action names.
var normalKeyMap = map[string]string{
	"p": "install",
	"P": "push",
	"d": "doctor",
	"x": "confirm-remove",
}

// ctrlKeyMap maps ctrl-keys in terminal input mode to action names.
var ctrlKeyMap = map[string]string{
	"ctrl+p":     "install",
	"ctrl+u":     "push",
	"ctrl+d":     "doctor",
	"ctrl+x":     "confirm-remove",
	"ctrl+left":  "channel-prev",
	"ctrl+right": "channel-next",
}

// --- Dashboard ---

// Dashboard is the top-level bubbletea Model. It is a pure message router:
// it broadcasts messages to all panels and maintains only the state needed
// for key routing (executing, confirmRemove).
type Dashboard struct {
	channel  *ChannelStrip // concrete ref for Selected()/SelectedIndex()
	overview *Overview     // concrete ref for View assembly
	scope    *Scope        // concrete ref for View assembly
	terminal *Terminal     // concrete ref for InputMode() check + View assembly
	controls *Controls     // concrete ref for BuildRow + View assembly
	panels   []Panel       // ordered list for broadcast and focus cycling

	focus         int
	dfPath        string
	modules       []*module.Module
	manifest      *manifest.Manifest
	gitChanges    []string
	executing     bool
	confirmRemove bool
	width, height int
	theme         Theme
}

// RunDashboard is the public entry point that loads data and runs the TUI.
func RunDashboard(dfPath string) error {
	theme := DetectTheme()

	ch := NewChannelStrip(theme)
	ov := NewOverview(theme, dfPath)
	sc := NewScope(theme)
	tm := NewTerminal(theme)
	ct := NewControls(theme)

	// Channel strip starts focused.
	ch.SetFocus(true)

	d := &Dashboard{
		channel:  ch,
		overview: ov,
		scope:    sc,
		terminal: tm,
		controls: ct,
		panels:   []Panel{ch, ov, sc, tm, ct},
		focus:    0, // channel
		dfPath:   dfPath,
		theme:    theme,
	}

	p := tea.NewProgram(d, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Init implements tea.Model. Triggers initial data load.
func (d *Dashboard) Init() tea.Cmd {
	return reloadData(d.dfPath)
}

// Update implements tea.Model.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		d.width, d.height = m.Width, m.Height
		return d, nil
	case tea.KeyMsg:
		return d.handleKey(m)
	case TerminalExecMsg:
		return d.handleTerminalExec(m)
	default:
		return d.broadcast(msg)
	}
}

// broadcast sends a message to all panels and updates dashboard-level state.
func (d *Dashboard) broadcast(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Dashboard-level state flags (needed for key routing).
	switch msg.(type) {
	case CmdStartMsg:
		d.executing = true
	case CmdOutputMsg:
		d.executing = false
		cmds = append(cmds, reloadData(d.dfPath))
	case ConfirmStartMsg:
		d.confirmRemove = true
	case ConfirmCancelMsg:
		d.confirmRemove = false
	}
	if m, ok := msg.(DataReloadMsg); ok {
		d.modules = m.Modules
		d.manifest = m.Manifest
		d.gitChanges = m.GitChanges
	}

	// Broadcast to all panels.
	for i, p := range d.panels {
		updated, cmd := p.Update(msg)
		d.panels[i] = updated
		cmds = append(cmds, cmd)
	}

	// Keep concrete refs in sync with panels slice.
	d.syncPanelRefs()

	return d, tea.Batch(cmds...)
}

// broadcastAndExec broadcasts a message and batches it with an async command.
func (d *Dashboard) broadcastAndExec(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	_, bcmd := d.broadcast(msg)
	return d, tea.Batch(bcmd, cmd)
}

// syncPanelRefs updates concrete panel references after broadcast mutates the slice.
// Panel.Update returns Panel (interface), so after reassignment the concrete
// pointer in the slice may differ from the stored field.
func (d *Dashboard) syncPanelRefs() {
	for _, p := range d.panels {
		switch v := p.(type) {
		case *ChannelStrip:
			d.channel = v
		case *Overview:
			d.overview = v
		case *Scope:
			d.scope = v
		case *Terminal:
			d.terminal = v
		case *Controls:
			d.controls = v
		}
	}
}

// --- Key handling ---

func (d *Dashboard) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := m.String()

	// Confirm mode: only y/n/esc
	if d.confirmRemove {
		return d.handleConfirmKey(key)
	}

	// Executing: only quit
	if d.executing {
		if key == "q" || key == "ctrl+c" {
			return d, tea.Quit
		}
		return d, nil
	}

	// Terminal input mode: ctrl keys -> actions, rest -> terminal
	if d.terminal.InputMode() && d.panels[d.focus] == d.terminal {
		if action, ok := ctrlKeyMap[key]; ok {
			return d.execAction(action)
		}
		if key == "ctrl+a" {
			return d.broadcast(TerminalHintMsg{Text: "提示：请使用 dot add <module> 添加模块"})
		}
		if key == "ctrl+q" || key == "ctrl+c" {
			return d, tea.Quit
		}
		// All other keys go to terminal.
		_, cmd := d.terminal.Update(m)
		return d, cmd
	}

	// Normal mode: check action map first
	if action, ok := normalKeyMap[key]; ok {
		return d.execAction(action)
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
	case "a":
		return d.broadcast(TerminalHintMsg{Text: "提示：请使用 dot add <module> 添加模块"})
	case ":":
		d.setFocus(d.panelIndex(d.terminal))
		_, cmd := d.terminal.Update(m)
		return d, cmd
	case "esc":
		d.setFocus(0) // channel is always index 0
		return d, nil
	}

	return d, nil
}

func (d *Dashboard) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		mod := d.channel.Selected()
		if mod != nil {
			return d.broadcastAndExec(ConfirmCancelMsg{}, execUninstall(d.dfPath, mod.Name))
		}
		return d.broadcast(ConfirmCancelMsg{})
	case "n", "N", "esc":
		return d.broadcast(ConfirmCancelMsg{})
	}
	return d, nil
}

// execAction dispatches a named action from the key maps.
func (d *Dashboard) execAction(action string) (tea.Model, tea.Cmd) {
	mod := d.channel.Selected()
	switch action {
	case "install":
		if mod == nil {
			return d, nil
		}
		return d.broadcastAndExec(CmdStartMsg{}, execInstall(d.dfPath, mod.Name))
	case "push":
		return d.broadcastAndExec(CmdStartMsg{}, execPush(d.dfPath, "tui push"))
	case "doctor":
		if mod == nil {
			return d, nil
		}
		return d.broadcastAndExec(CmdStartMsg{}, execDoctor(d.dfPath, mod.Name))
	case "confirm-remove":
		if mod == nil {
			return d, nil
		}
		return d.broadcast(ConfirmStartMsg{ModuleName: mod.Name})
	case "channel-prev":
		_, cmd := d.channel.Update(tea.KeyMsg{Type: tea.KeyLeft})
		return d, cmd
	case "channel-next":
		_, cmd := d.channel.Update(tea.KeyMsg{Type: tea.KeyRight})
		return d, cmd
	}
	return d, nil
}

// handleTerminalExec uses the extracted command dispatcher.
func (d *Dashboard) handleTerminalExec(m TerminalExecMsg) (tea.Model, tea.Cmd) {
	cmds, msgs, errText := dispatchCommand(m.Input, d.dfPath, d.channel.Selected())
	if errText != "" {
		msgs = append(msgs, TerminalHintMsg{Text: errText})
	}
	var allCmds []tea.Cmd
	for _, msg := range msgs {
		_, cmd := d.broadcast(msg)
		allCmds = append(allCmds, cmd)
	}
	allCmds = append(allCmds, cmds...)
	return d, tea.Batch(allCmds...)
}

// --- Focus management ---

func (d *Dashboard) cycleFocus() {
	next := (d.focus + 1) % len(d.panels)
	d.setFocus(next)
}

func (d *Dashboard) setFocus(idx int) {
	d.panels[d.focus].SetFocus(false)
	d.focus = idx
	d.panels[d.focus].SetFocus(true)
}

func (d *Dashboard) panelIndex(target Panel) int {
	for i, p := range d.panels {
		if p == target {
			return i
		}
	}
	return 0
}
