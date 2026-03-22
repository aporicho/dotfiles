# TUI 全面解耦设计

日期: 2026-03-22
状态: 草案
前置: [TUI 布局解耦](2026-03-22-tui-layout-decoupling-design.md)（已完成）

## 问题

布局解耦完成后，TUI 仍有 11 个耦合点。最严重的两个：

1. **Dashboard 是 God Object** — 直接持有 5 个面板引用，57+ 次直接方法调用，任何面板行为变化都要改 dashboard.go
2. **Controls 绕过 Panel 接口** — `SetExecuting()`/`SetConfirming()`/`SetConfirmName()` 20+ 次调用不走 `Update(msg)` 机制

其他耦合：消息定义分散、数据重载不一致（ChannelStrip 被重建 vs 其他走 Update）、Controls 对齐依赖 panelRow 列结构、buildPanelRow 直接调用具体面板类型的 View、handleKey 中 normal/terminal mode 命令逻辑重复、Styles+Theme 冗余传参。

## 方案选择

**方案 A：纯消息驱动（Elm 架构最大化）** — 已选定

彻底发挥 bubbletea 的 Elm 模式。所有面板状态变更通过 `tea.Msg`，Dashboard 变为纯消息路由器。

淘汰的方案：
- 方案 B（事件总线 + 订阅）：对 bubbletea 过度抽象，反射机制不够 Go 惯用
- 方案 C（Dashboard 拆分为子组合体）：没有从根本上解决 setter 绕过接口的问题

## 设计

### 1. 消息体系重构

`messages.go` 统一定义所有消息类型：

```go
// === 用户交互 ===
type ModuleSelectedMsg struct { Index int; Module *module.Module }
type TerminalExecMsg  struct { Input string }           // 从 panel_terminal.go 移入

// === 命令生命周期 ===
type CmdStartMsg   struct{}
type CmdOutputMsg  struct { Output string; Err error }

// === 面板状态 ===
type ConfirmStartMsg  struct { ModuleName string }      // 新增：替代 SetConfirming+SetConfirmName
type ConfirmCancelMsg struct{}                           // 新增：替代 SetConfirming(false)
type TerminalHintMsg  struct { Text string }             // 新增：替代 AppendOutput()

// === 数据 ===
type DataReloadMsg struct {
    Modules    []*module.Module
    Manifest   *manifest.Manifest
    GitChanges []string
}
```

关键变化：
- `TerminalExecMsg` 从 panel_terminal.go 移到 messages.go
- 新增 `ConfirmStartMsg` / `ConfirmCancelMsg` 替代 Controls 的 setter 方法
- 新增 `TerminalHintMsg` 替代 `AppendOutput()` 直接调用
- `CmdStartMsg` 和 `CmdOutputMsg` 已自然传达执行状态，无需额外 setter

### 2. Panel 接口扩展

```go
type Panel interface {
    Update(msg tea.Msg) (Panel, tea.Cmd)
    View(width, height int) string
    Focused() bool
    SetFocus(bool)
    Weight() int  // 新增：宽度权重，用于 buildPanelRow 计算列宽
}
```

各面板 Weight 值：Overview=1, Scope=2, Terminal=2, ChannelStrip=0, Controls=0。

注意：当前 2 列模式是 Scope:Terminal = 1:1（各 pw/2）。为保持该比例，Terminal.Weight 必须等于 Scope.Weight=2。3 列模式比例从当前 1:2:1 变为 1:2:2（Overview 面板只在宽屏>=100 显示，占比减小可接受）。

### 3. Dashboard 瘦身为纯消息路由器

Dashboard.Update() 核心逻辑变为「收到消息 → 广播给所有面板」：

```go
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m := msg.(type) {
    case tea.WindowSizeMsg:
        d.width, d.height = m.Width, m.Height
        return d, nil
    case tea.KeyMsg:
        return d.handleKey(m)
    default:
        return d.broadcast(msg)
    }
}
```

`broadcast(msg)` 方法：

```go
func (d *Dashboard) broadcast(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Dashboard 级状态更新（按键路由需要这些标志）
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
    case DataReloadMsg:
        m := msg.(DataReloadMsg)
        d.modules, d.manifest, d.gitChanges = m.Modules, m.Manifest, m.GitChanges
    }

    // 广播给所有面板
    for i, p := range d.panels {
        updated, cmd := p.Update(msg)
        d.panels[i] = updated
        cmds = append(cmds, cmd)
    }
    return d, tea.Batch(cmds...)
}
```

`panels` 辅助字段：

```go
type Dashboard struct {
    channel  *ChannelStrip  // 保留具体引用：Selected() 和 SelectedIndex() 需要
    panels   []Panel        // 所有面板有序列表，用于 broadcast 和焦点切换

    focus         int       // 当前焦点索引（对应 panels 切片）
    dfPath        string
    modules       []*module.Module
    manifest      *manifest.Manifest
    gitChanges    []string
    executing     bool
    confirmRemove bool
    width, height int
    theme         Theme
}
```

Dashboard 不再调用任何面板的非接口方法（SetExecuting 等全部删除）。只保留 `channel` 具体引用以访问 `Selected()`/`SelectedIndex()`。

### 4. handleKey 重构：keyActions 映射表

用映射表消除 normal mode 和 terminal input mode 的命令重复：

```go
var normalKeyMap = map[string]string{
    "p": "install",
    "P": "push",
    "d": "doctor",
    "x": "confirm-remove",
}

var ctrlKeyMap = map[string]string{
    "ctrl+p":     "install",
    "ctrl+u":     "push",
    "ctrl+d":     "doctor",
    "ctrl+x":     "confirm-remove",
    "ctrl+left":  "channel-prev",  // 保留：terminal input mode 下切换模块
    "ctrl+right": "channel-next",  // 保留：terminal input mode 下切换模块
}
```

Normal mode 查 `normalKeyMap`，terminal input mode 查 `ctrlKeyMap`。两者调用同一个 `execAction(name)`：

```go
func (d *Dashboard) execAction(action string) (tea.Model, tea.Cmd) {
    mod := d.channel.Selected()
    switch action {
    case "install":
        if mod == nil { return d, nil }
        return d.broadcastAndExec(CmdStartMsg{}, execInstall(d.dfPath, mod.Name))
    case "push":
        return d.broadcastAndExec(CmdStartMsg{}, execPush(d.dfPath, "tui push"))
    case "doctor":
        if mod == nil { return d, nil }
        return d.broadcastAndExec(CmdStartMsg{}, execDoctor(d.dfPath, mod.Name))
    case "confirm-remove":
        if mod == nil { return d, nil }
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

// broadcastAndExec 广播消息并附带一个异步命令
func (d *Dashboard) broadcastAndExec(msg tea.Msg, cmd tea.Cmd) (tea.Model, tea.Cmd) {
    _, bcmd := d.broadcast(msg)
    return d, tea.Batch(bcmd, cmd)
}
```

handleKey 结构：

```go
func (d *Dashboard) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
    key := m.String()

    if d.confirmRemove {
        return d.handleConfirmKey(key)
    }
    if d.executing {
        if key == "q" || key == "ctrl+c" { return d, tea.Quit }
        return d, nil
    }

    // Terminal input mode
    if tm, ok := d.panels[d.focus].(*Terminal); ok && tm.InputMode() {
        if action, ok := ctrlKeyMap[key]; ok {
            return d.execAction(action)
        }
        if key == "ctrl+a" {
            return d.broadcast(TerminalHintMsg{Text: "提示：请使用 dot add <module> 添加模块"})
        }
        if key == "ctrl+q" || key == "ctrl+c" {
            return d, tea.Quit
        }
        // 其他键给 Terminal 处理
        _, cmd := d.panels[d.focus].Update(m)
        return d, cmd
    }

    // Normal mode
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
        d.setFocusToTerminal()
        _, cmd := d.panels[d.focus].Update(m)
        return d, cmd
    case "esc":
        d.setFocusToChannel()
        return d, nil
    }
    return d, nil
}
```

### 5. 命令分发提取

从 dashboard.go 提取 `handleTerminalExec` 到独立的 `command_dispatch.go`：

```go
// command_dispatch.go

// dispatchCommand 解析终端命令输入，返回要执行的 tea.Cmd 和要广播的 tea.Msg 列表。
func dispatchCommand(input string, dfPath string, selectedMod *module.Module) ([]tea.Cmd, []tea.Msg, string) {
    parts := strings.Fields(input)
    if len(parts) == 0 {
        return nil, nil, ""
    }
    cmd, args := parts[0], parts[1:]

    switch cmd {
    case "install":
        name := resolveModName(args, selectedMod)
        if name == "" { return nil, nil, "错误：未指定模块" }
        return []tea.Cmd{execInstall(dfPath, name)},
               []tea.Msg{CmdStartMsg{}}, ""

    case "pull":
        return []tea.Cmd{execPull(dfPath)},
               []tea.Msg{CmdStartMsg{}}, ""

    case "push":
        msg := "tui push"
        if len(args) > 0 { msg = strings.Join(args, " ") }
        return []tea.Cmd{execPush(dfPath, msg)},
               []tea.Msg{CmdStartMsg{}}, ""

    case "doctor":
        name := resolveModName(args, selectedMod)
        if name == "" { return nil, nil, "错误：未指定模块" }
        return []tea.Cmd{execDoctor(dfPath, name)},
               []tea.Msg{CmdStartMsg{}}, ""

    case "uninstall":
        name := resolveModName(args, selectedMod)
        if name == "" { return nil, nil, "错误：未指定模块" }
        return []tea.Cmd{execUninstall(dfPath, name)},
               []tea.Msg{CmdStartMsg{}}, ""

    case "remove":
        return nil, nil, "请使用 dot remove <module> 命令行操作（不可逆操作，TUI 不支持）"

    default:
        return nil, nil, "未知命令: " + input
    }
}

func resolveModName(args []string, selected *module.Module) string {
    if len(args) > 0 { return args[0] }
    if selected != nil { return selected.Name }
    return ""
}
```

Dashboard 处理 TerminalExecMsg 的方式：

```go
case TerminalExecMsg:
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
```

### 6. 面板自治

#### Controls

删除 `SetExecuting()`、`SetConfirming()`、`SetConfirmName()` 三个 setter。Update 响应消息：

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

#### Terminal

`AppendOutput` 降为私有方法 `appendOutput`。外部通过消息注入：

```go
func (t *Terminal) Update(msg tea.Msg) (Panel, tea.Cmd) {
    switch m := msg.(type) {
    case CmdOutputMsg:
        // 已有逻辑不变
    case TerminalHintMsg:
        t.appendOutput(m.Text)
        return t, nil
    case tea.KeyMsg:
        // 已有逻辑不变
    }
    return t, nil
}
```

#### ChannelStrip

构造器简化为只接收 theme。数据通过 DataReloadMsg 传入：

```go
func NewChannelStrip(theme Theme) *ChannelStrip {
    return &ChannelStrip{
        theme:  theme,
        styles: NewStyles(theme),
    }
}

func (cs *ChannelStrip) Update(msg tea.Msg) (Panel, tea.Cmd) {
    switch m := msg.(type) {
    case DataReloadMsg:
        cs.modules = m.Modules
        cs.manifest = m.Manifest
        cs.gitChanges = m.GitChanges
        cs.rebuildChips()
        if cs.selected >= len(cs.modules) {
            cs.selected = max(0, len(cs.modules)-1)
        }
        // 触发模块选中事件
        if len(cs.modules) > 0 {
            return cs, func() tea.Msg {
                return ModuleSelectedMsg{Index: cs.selected, Module: cs.modules[cs.selected]}
            }
        }
        return cs, nil
    case tea.KeyMsg:
        // 已有左右导航逻辑
    }
    return cs, nil
}
```

#### Overview, Scope

构造器简化，增加 Weight()：

```go
func NewOverview(theme Theme, dfPath string) *Overview {
    return &Overview{theme: theme, styles: NewStyles(theme), dfPath: dfPath}
}
func (o *Overview) Weight() int { return 1 }

func NewScope(theme Theme) *Scope {
    return &Scope{theme: theme, styles: NewStyles(theme)}
}
func (s *Scope) Weight() int { return 2 }
```

### 7. View 层解耦

buildPanelRow 泛化为接收 []Panel，用 Weight 计算列宽：

```go
func buildPanelRow(totalW, panelH int, panels []Panel) Row {
    n := len(panels)
    pw := totalW - (n + 1)
    if pw < n { pw = n }

    totalWeight := 0
    for _, p := range panels { totalWeight += p.Weight() }

    cols := make([]int, n)
    used := 0
    for i, p := range panels {
        cols[i] = pw * p.Weight() / totalWeight
        used += cols[i]
    }
    cols[n/2] += pw - used  // 舍入误差给中间面板

    contents := make([]string, n)
    for i, p := range panels { contents[i] = p.View(cols[i], panelH) }
    return Row{Cols: cols, Contents: contents, Height: panelH}
}
```

Dashboard.View() 调用：

```go
var middlePanels []Panel
if lay.ShowOverview {
    middlePanels = []Panel{d.overview(), d.scope(), d.terminal()}
} else {
    middlePanels = []Panel{d.scope(), d.terminal()}
}
panelRow := buildPanelRow(lay.TotalW, lay.PanelH, middlePanels)
```

Dashboard.View() 中直接使用具名字段引用（`d.overview`、`d.scope`、`d.terminal`），不引入额外的辅助方法。Dashboard 保留这些具名字段是因为 View 组装需要知道哪些面板属于中间行。

### 8. 数据流统一

构造时只传 theme，数据全走消息：

```go
func (d *Dashboard) Init() tea.Cmd {
    return reloadData(d.dfPath)
}
```

DataReloadMsg 到达时，所有面板通过 broadcast 统一更新。初始化和重载走同一条路径。

注意：Init() 返回异步 cmd，首帧渲染时 DataReloadMsg 尚未到达，面板数据为空。各面板的 View() 应优雅处理空数据状态（显示空白或占位文本）。当前面板已经能处理空数据（modules=nil 时显示空列表），所以这不是新问题。

### 9. handleConfirmKey

从当前 dashboard.go 228-244 行提取：

```go
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
```

注意：确认后同时广播 ConfirmCancelMsg（清除确认状态）并��行卸载命令。

## 文件改动范围

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `messages.go` | 重构 | 集中所有消息类型，新增 ConfirmStartMsg/ConfirmCancelMsg/TerminalHintMsg |
| `panel.go` | 扩展 | Panel 接口增加 Weight() |
| `dashboard.go` | 重写 | 瘦身为消息路由器 + keyActions 映射 + broadcast |
| `dashboard_view.go` | 重构 | buildPanelRow 泛化为 []Panel + Weight |
| `panel_controls.go` | 重构 | 删除 setter，Update 响应消息 |
| `panel_terminal.go` | 重构 | TerminalExecMsg 移走，AppendOutput 私有化，响应 TerminalHintMsg |
| `panel_channel.go` | 重构 | 构造器简化，Update 处理 DataReloadMsg |
| `panel_overview.go` | 小改 | 构造器简化，加 Weight() |
| `panel_scope.go` | 小改 | 构造器简化，加 Weight() |
| `command_dispatch.go` | 新建 | 从 dashboard.go 提取命令分发逻辑 |
| `commands.go` | 不变 | |
| `layout.go` | 不变 | |
| `frame.go` | 不变 | |

## 验证标准

重构完成后，以下操作应该只涉及 1 个文件：

| 操作 | 只改的文件 |
|------|-----------|
| 新增一个命令（如 "status"） | `command_dispatch.go` |
| 新增一个面板 | 新建 `panel_xxx.go` + Dashboard 面板注册 |
| 改变面板宽度比例 | 对应 `panel_xxx.go` 的 `Weight()` |
| 改 Controls 按钮文案 | `panel_controls.go` |
| 新增一种消息类型 | `messages.go` + 响应的面板 |
| 改快捷键绑定 | `dashboard.go` 的 keyMap |

## 风险点

1. **ChannelStrip.Selected()** — Dashboard 仍需具体引用获取当前选中模块，这是合理的特例
2. **Controls.BuildRow(panelRow)** — 对齐逻辑仍依赖 panelRow 结构，这是视觉对齐的必然要求，不属于逻辑耦合
3. **编译兼容性** — 大刀阔斧改完需要一次性编译通过，无中间可编译态
