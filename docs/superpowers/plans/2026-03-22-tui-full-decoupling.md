# TUI Full Decoupling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Decouple the TUI Dashboard from panel internals by making all state changes flow through bubbletea messages, eliminating direct setter calls, and extracting command dispatch.

**Architecture:** Dashboard becomes a pure message router that broadcasts `tea.Msg` to all panels. Panels self-manage state by responding to messages in their own `Update()`. Command parsing is extracted to a pure function in `command_dispatch.go`. The Panel interface gains `Weight() int` so `buildPanelRow` can compute column widths without knowing concrete types.

**Tech Stack:** Go, bubbletea (Elm architecture), lipgloss

**Spec:** `docs/superpowers/specs/2026-03-22-tui-full-decoupling-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `dot/internal/tui/messages.go` | Modify | All message type definitions (centralized) |
| `dot/internal/tui/panel.go` | Modify | Panel interface with Weight() |
| `dot/internal/tui/panel_controls.go` | Modify | Controls panel — remove setters, respond to messages |
| `dot/internal/tui/panel_terminal.go` | Modify | Terminal panel — privatize AppendOutput, respond to TerminalHintMsg |
| `dot/internal/tui/panel_channel.go` | Modify | ChannelStrip — simplified constructor, Update handles DataReloadMsg |
| `dot/internal/tui/panel_overview.go` | Modify | Overview — simplified constructor, add Weight() |
| `dot/internal/tui/panel_scope.go` | Modify | Scope — simplified constructor, add Weight() |
| `dot/internal/tui/command_dispatch.go` | Create | Pure function for terminal command parsing |
| `dot/internal/tui/dashboard.go` | Rewrite | Message router + broadcast + keyActions mapping |
| `dot/internal/tui/dashboard_view.go` | Modify | buildPanelRow uses []Panel + Weight |

---

### Task 1: Centralize Messages

**Files:**
- Modify: `dot/internal/tui/messages.go`
- Modify: `dot/internal/tui/panel_terminal.go:13-16` (remove TerminalExecMsg definition)

- [ ] **Step 1: Add new message types to messages.go**

Add after the existing `DataReloadMsg` definition:

```go
// ConfirmStartMsg signals the dashboard entered remove-confirmation mode.
type ConfirmStartMsg struct{ ModuleName string }

// ConfirmCancelMsg signals the dashboard exited remove-confirmation mode.
type ConfirmCancelMsg struct{}

// TerminalHintMsg carries a hint/info line to display in the terminal panel.
type TerminalHintMsg struct{ Text string }
```

- [ ] **Step 2: Move TerminalExecMsg from panel_terminal.go to messages.go**

Remove lines 13-16 from `panel_terminal.go`:

```go
// TerminalExecMsg is sent when the user submits a command in the terminal panel.
type TerminalExecMsg struct {
	Input string
}
```

Add to `messages.go` in the user interaction section:

```go
// TerminalExecMsg is sent when the user submits a command in the terminal panel.
type TerminalExecMsg struct{ Input string }
```

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/...`
Expected: SUCCESS (no behavior change, just moved definitions)

- [ ] **Step 4: Commit**

```bash
git add dot/internal/tui/messages.go dot/internal/tui/panel_terminal.go
git commit -m "refactor: centralize all message types in messages.go"
```

---

### Task 2: Extend Panel Interface with Weight()

**Files:**
- Modify: `dot/internal/tui/panel.go`
- Modify: `dot/internal/tui/panel_overview.go`
- Modify: `dot/internal/tui/panel_scope.go`
- Modify: `dot/internal/tui/panel_terminal.go`
- Modify: `dot/internal/tui/panel_channel.go`
- Modify: `dot/internal/tui/panel_controls.go`

- [ ] **Step 1: Add Weight() to Panel interface**

Replace the entire `panel.go` content:

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

// Panel is the interface all dashboard panels implement.
type Panel interface {
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View(width, height int) string
	Focused() bool
	SetFocus(bool)
	Weight() int // column width weight for buildPanelRow (0 = not a middle panel)
}
```

- [ ] **Step 2: Add Weight() to each panel**

Add to `panel_overview.go` after `SetFocus`:

```go
// Weight implements Panel.
func (o *Overview) Weight() int { return 1 }
```

Add to `panel_scope.go` after `SetFocus`:

```go
// Weight implements Panel.
func (s *Scope) Weight() int { return 2 }
```

Add to `panel_terminal.go` after `SetFocus`:

```go
// Weight implements Panel.
func (t *Terminal) Weight() int { return 2 }
```

Add to `panel_channel.go` after `SetFocus`:

```go
// Weight implements Panel.
func (cs *ChannelStrip) Weight() int { return 0 }
```

Add to `panel_controls.go` after `SetFocus`:

```go
// Weight implements Panel.
func (c *Controls) Weight() int { return 0 }
```

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/...`
Expected: SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dot/internal/tui/panel.go dot/internal/tui/panel_overview.go dot/internal/tui/panel_scope.go dot/internal/tui/panel_terminal.go dot/internal/tui/panel_channel.go dot/internal/tui/panel_controls.go
git commit -m "refactor: add Weight() to Panel interface for column width calculation"
```

---

### Task 3: Controls Panel — Simplify Constructor, Remove Setters, Respond to Messages

**Files:**
- Modify: `dot/internal/tui/panel_controls.go`

- [ ] **Step 1: Simplify constructor to take only theme**

Replace the existing `NewControls`:

```go
func NewControls(theme Theme) *Controls {
	return &Controls{styles: NewStyles(theme), theme: theme}
}
```

- [ ] **Step 2: Remove setter methods and update Update()**

Delete these three lines:

```go
func (c *Controls) SetExecuting(v bool)   { c.executing = v }
func (c *Controls) SetConfirming(v bool)  { c.confirming = v }
func (c *Controls) SetConfirmName(n string) { c.confirmName = n }
```

Replace the existing `Update` method with:

```go
func (c *Controls) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch m := msg.(type) {
	case CmdStartMsg:
		c.executing = true
	case CmdOutputMsg:
		c.executing = false
	case ConfirmStartMsg:
		c.confirming = true
		c.confirmName = m.ModuleName
	case ConfirmCancelMsg:
		c.confirming = false
		c.confirmName = ""
	}
	return c, nil
}
```

- [ ] **Step 3: Verify compilation fails (dashboard.go still calls setters)**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/... 2>&1 | head -20`
Expected: FAIL with `SetExecuting undefined`, `SetConfirming undefined`, etc. This confirms nothing else secretly depends on these methods.

Note: We'll fix the call sites in Task 8 (Dashboard rewrite). The compilation will remain broken until then.

- [ ] **Step 4: Commit**

```bash
git add dot/internal/tui/panel_controls.go
git commit -m "refactor: Controls responds to messages, remove setter methods

Note: dashboard.go call sites not yet updated — will compile after Task 8."
```

---

### Task 4: Terminal Panel — Simplify Constructor, Privatize AppendOutput, Add TerminalHintMsg

**Files:**
- Modify: `dot/internal/tui/panel_terminal.go`

- [ ] **Step 1: Simplify constructor to take only theme**

Replace the existing `NewTerminal`:

```go
// NewTerminal constructs a Terminal panel.
func NewTerminal(theme Theme) *Terminal {
	vp := viewport.New(0, 0)
	return &Terminal{
		styles:       NewStyles(theme),
		theme:        theme,
		viewport:     vp,
		historyIndex: -1,
	}
}
```

- [ ] **Step 2: Rename AppendOutput to appendOutput (lowercase)**

In `panel_terminal.go`, change:

```go
// AppendOutput adds lines to the terminal output history programmatically.
func (t *Terminal) AppendOutput(text string) {
```

To:

```go
// appendOutput adds lines to the terminal output history.
func (t *Terminal) appendOutput(text string) {
```

- [ ] **Step 3: Add TerminalHintMsg handling to Update()**

Add a new case in the Update switch, after the `CmdOutputMsg` case:

```go
	case TerminalHintMsg:
		t.appendOutput(m.Text)
		return t, nil
```

The full Update becomes:

```go
func (t *Terminal) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch m := msg.(type) {
	case CmdOutputMsg:
		isErr := m.Err != nil
		for _, line := range strings.Split(m.Output, "\n") {
			t.lines = append(t.lines, outputLine{
				text:      line,
				timestamp: time.Now(),
				isError:   isErr,
			})
		}
		if len(t.lines) > maxLines {
			t.lines = t.lines[len(t.lines)-maxLines:]
		}
		t.syncViewport()
		t.viewport.GotoBottom()
		return t, nil

	case TerminalHintMsg:
		t.appendOutput(m.Text)
		return t, nil

	case tea.KeyMsg:
		if !t.focused {
			return t, nil
		}
		if t.inputMode {
			return t.handleInputMode(m)
		}
		return t.handleNavMode(m)
	}
	return t, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add dot/internal/tui/panel_terminal.go
git commit -m "refactor: Terminal simplified constructor, responds to TerminalHintMsg, privatize appendOutput

Note: dashboard.go call sites not yet updated — will compile after Task 8."
```

---

### Task 5: ChannelStrip — Simplified Constructor, Update Handles DataReloadMsg

**Files:**
- Modify: `dot/internal/tui/panel_channel.go`

- [ ] **Step 1: Simplify constructor**

Replace the existing `NewChannelStrip` function:

```go
// NewChannelStrip constructs a ChannelStrip. Data arrives via DataReloadMsg.
func NewChannelStrip(theme Theme) *ChannelStrip {
	return &ChannelStrip{
		styles: NewStyles(theme),
		theme:  theme,
	}
}
```

- [ ] **Step 2: Add DataReloadMsg handling to Update**

Replace the existing `Update` method to handle both DataReloadMsg and key navigation:

```go
// Update implements Panel. Handles DataReloadMsg for data and ←→ for navigation.
func (cs *ChannelStrip) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch m := msg.(type) {
	case DataReloadMsg:
		cs.modules = m.Modules
		cs.manifest = m.Manifest
		cs.gitChanges = m.GitChanges
		// Preserve selection if still valid, else reset.
		if cs.selected >= len(cs.modules) {
			cs.selected = cs.firstNavigable()
		}
		// Emit selection so Scope updates.
		return cs, cs.emitSelected()

	case tea.KeyMsg:
		if !cs.focused {
			return cs, nil
		}
		switch m.String() {
		case "left", "h":
			cs.movePrev()
			return cs, cs.emitSelected()
		case "right", "l":
			cs.moveNext()
			return cs, cs.emitSelected()
		}
	}
	return cs, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/panel_channel.go
git commit -m "refactor: ChannelStrip constructor takes only theme, Update handles DataReloadMsg

Note: dashboard.go call sites not yet updated — will compile after Task 8."
```

---

### Task 6: Overview and Scope — Simplified Constructors, Weight

**Files:**
- Modify: `dot/internal/tui/panel_overview.go`
- Modify: `dot/internal/tui/panel_scope.go`

- [ ] **Step 1: Simplify Overview constructor**

Replace `NewOverview`:

```go
// NewOverview constructs an Overview panel. Data arrives via DataReloadMsg.
func NewOverview(theme Theme, dfPath string) *Overview {
	return &Overview{
		dfPath: dfPath,
		styles: NewStyles(theme),
		theme:  theme,
	}
}
```

- [ ] **Step 2: Simplify Scope constructor**

Replace `NewScope`:

```go
// NewScope constructs a Scope panel. Data arrives via DataReloadMsg and ModuleSelectedMsg.
func NewScope(theme Theme) *Scope {
	return &Scope{
		styles: NewStyles(theme),
		theme:  theme,
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/panel_overview.go dot/internal/tui/panel_scope.go
git commit -m "refactor: simplify Overview/Scope constructors, data via messages

Note: dashboard.go call sites not yet updated — will compile after Task 8."
```

---

### Task 7: Create command_dispatch.go

**Files:**
- Create: `dot/internal/tui/command_dispatch.go`

- [ ] **Step 1: Create the file**

```go
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
		return nil, nil, "请使用 dot remove <module> 命��行操作（不可逆操作，TUI 不支持）"

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
```

- [ ] **Step 2: Commit**

```bash
git add dot/internal/tui/command_dispatch.go
git commit -m "refactor: extract command dispatch to pure function

Note: not yet wired in — will compile after Task 8."
```

---

### Task 8: Rewrite Dashboard — Message Router + keyActions + broadcast

This is the big task. Dashboard becomes a thin message router.

**Files:**
- Rewrite: `dot/internal/tui/dashboard.go`

- [ ] **Step 1: Rewrite dashboard.go**

Replace the entire file with:

```go
package tui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/aporicho/dotfiles/dot/internal/git"
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

	// Terminal input mode: ctrl keys → actions, rest → terminal
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
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/...`
Expected: SUCCESS — all setter call sites are gone, all panels respond to messages.

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/dashboard.go
git commit -m "refactor: Dashboard is now a pure message router

- broadcast() sends all messages to all panels
- keyActions mapping eliminates normal/ctrl mode duplication
- execAction() dispatches named actions
- handleTerminalExec() uses extracted dispatchCommand()
- No more direct setter calls on panels"
```

---

### Task 9: Refactor dashboard_view.go — buildPanelRow Uses []Panel + Weight

**Files:**
- Modify: `dot/internal/tui/dashboard_view.go`

- [ ] **Step 1: Rewrite buildPanelRow and update View**

Replace the entire `dashboard_view.go`:

```go
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (d *Dashboard) View() string {
	if d.width < 60 || d.height < 12 {
		return "终端过小，请调整窗口大小（至少 60×12）"
	}

	lay := ComputeLayout(d.width, d.height)
	borderSty := lipgloss.NewStyle().Foreground(lipgloss.Color(d.theme.PanelBorder()))

	var rows []Row

	// Row 1: Channel Strip
	chipRow := d.channel.BuildRow(lay.TotalW)
	rows = append(rows, chipRow)

	// Row 2: Middle panels (uses Weight for column ratios)
	var middlePanels []Panel
	if lay.ShowOverview {
		middlePanels = []Panel{d.overview, d.scope, d.terminal}
	} else {
		middlePanels = []Panel{d.scope, d.terminal}
	}
	panelRow := buildPanelRow(lay.TotalW, lay.PanelH, middlePanels)
	rows = append(rows, panelRow)

	// Row 3: Controls
	if lay.ShowControls {
		rows = append(rows, d.controls.BuildRow(lay.TotalW, panelRow))
	}

	// Row 4: Footer
	if lay.ShowFooter {
		rows = append(rows, buildFooterRow(lay.TotalW, d.theme))
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(d.theme.Dimmed())).
		Render("\uf054 CHANNEL STRIP")

	return RenderUnifiedFrame(borderSty, title, d.width, rows)
}

// buildPanelRow produces the middle-panel Row using Panel.Weight() for column ratios.
func buildPanelRow(totalW, panelH int, panels []Panel) Row {
	n := len(panels)
	pw := totalW - (n + 1) // border chars
	if pw < n {
		pw = n
	}

	totalWeight := 0
	for _, p := range panels {
		totalWeight += p.Weight()
	}
	if totalWeight == 0 {
		totalWeight = n // fallback: equal weights
	}

	cols := make([]int, n)
	used := 0
	for i, p := range panels {
		cols[i] = pw * p.Weight() / totalWeight
		used += cols[i]
	}
	// Give rounding remainder to the middle panel.
	cols[n/2] += pw - used

	contents := make([]string, n)
	for i, p := range panels {
		contents[i] = p.View(cols[i], panelH)
	}
	return Row{Cols: cols, Contents: contents, Height: panelH}
}

// buildFooterRow produces a single-column footer Row.
func buildFooterRow(totalW int, theme Theme) Row {
	w := totalW - 2
	if w < 1 {
		w = 1
	}
	hint := "←→ module · : terminal · esc back · p install · P push · d doctor · x uninstall · q quit"
	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.FooterFg())).
		Width(w).
		Render(hint)
	return Row{Cols: []int{w}, Contents: []string{content}, Height: 1}
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/...`
Expected: SUCCESS

- [ ] **Step 3: Commit**

```bash
git add dot/internal/tui/dashboard_view.go
git commit -m "refactor: buildPanelRow uses []Panel + Weight for column ratios"
```

---

### Task 10: Final Verification

- [ ] **Step 1: Build the entire project**

Run: `cd /Users/aporicho/dotfiles && go build ./dot/...`
Expected: SUCCESS with zero errors

- [ ] **Step 2: Run existing tests**

Run: `cd /Users/aporicho/dotfiles && go test ./dot/...`
Expected: All tests pass

- [ ] **Step 3: Manual smoke test**

Run: `cd /Users/aporicho/dotfiles && go run ./dot tui`

Verify:
1. TUI launches, shows channel strip with module chips
2. ←→ navigates modules, Scope updates
3. Press `:` to enter terminal mode, type `install <module>`, press Enter
4. Controls shows "执行中..." during execution
5. Press `x` to trigger confirm mode, `n` to cancel
6. Press `q` to quit
7. Resize terminal — layout adapts (overview appears/disappears at width 100)

- [ ] **Step 4: Verify decoupling criteria**

Check each criterion from the spec:
- Adding a command only requires `command_dispatch.go` ✓
- Changing button text only requires `panel_controls.go` ✓
- Changing panel width only requires that panel's `Weight()` ✓
- Changing keybinding only requires the keyMap in `dashboard.go` ✓

- [ ] **Step 5: Final commit if any fixes were needed**

Only if steps 1-3 revealed issues that required fixes.
