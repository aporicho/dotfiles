# TUI Dashboard Design

## Problem

dot CLI 目前是纯子命令模式，没有统一的交互界面。用户需要记忆命令，无法一眼看到所有模块状态。需要一个 TUI 总控制台，裸 `dot`（无子命令）时进入。

## Solution

合成器风格的多面板 TUI 仪表盘。五个面板 + 全宽控制栏，用 bubbletea + lipgloss 实现，Tokyo Night 配色。

## Layout

```
┌──────────────────────────────────────────────────────┐
│ DOT  [ZSH●🔐] [KITTY●] [NVIM●⚡] ... [+ ADD a]      │  Channel Strip
├──────────┬──────────────────┬─────────────────────────┤
│ Overview │ Scope: ZSH       │ Terminal                │
│          │                  │                         │
│ Health   │ LNK ████████ 2/2 │ 22:14:32                │
│ ██████98%│ SEC ████████ 🔐  │ $ dot doctor zsh        │
│          │ GIT ████████ ok  │   ✓ links healthy       │
│ 7 mod    │                  │   ✓ secrets consistent  │
│ 7 inst   │ Signal Path      │   All passed ✓          │
│ 1 secret │ ● .zshrc → ~    │                         │
│ 1 changed│ ● config/ → ~   │                         │
│          │ ● secrets 🔐    │                         │
│ main ✓   │                  │                         │
│ darwin   │ Deps (brew)      │                         │
│ keychain │ starship · fzf   │ dot ▸ pull zsh▋         │
├──────────┴──────────────────┴─────────────────────────┤
│ ⬇ PULL p    │  ⬆ PUSH P    │  ⚕ DOCTOR d  │  ✕ RM x │  Controls
├───────────────────────────────────────────────────────┤
│ ←→ module · tab panel · : terminal · esc back · q quit│  Footer
└───────────────────────────────────────────────────────┘
```

## Panels

### 1. Channel Strip（顶部全宽）

模块通道条。每个模块一个芯片，显示名称 + LED 状态灯 + 标记（🔐 secrets / ⚡ changes）。末尾是 `+ ADD a` 虚线槽位。

- ←→ 键切换选中模块，所有面板联动
- 选中模块高亮（亮色边框 + 背景色变化，通过 lipgloss Border + Background 实现）
- LED 颜色：绿=healthy / 黄=有变更 / 红=异常
- 模块过多时横向截断，显示 `◂ ▸` 滚动指示

不匹配当前平台的模块灰色显示但保留在 Strip 中（不可选中）。

### 2. Overview（左侧固定宽度）

全局状态面板，不随选中模块变化。

内容：
- Health 进度条（所有模块健康百分比）
- 四宫格统计：modules / installed / secrets / changed
- System 信息：git branch + sync status / platform / keychain status

Health 计算：(healthy_links + healthy_secrets) / total_expected * 100

### 3. Scope（中间）

选中模块的详细信息，随 Channel Strip 切换联动。

内容分两部分：
- **Meter 摘要条**（顶部）：LNK / SEC / GIT 三根进度条，快速了解模块三个维度的健康度
- **详情**：Signal Path（所有链接及状态）、Deps（系统依赖）

Meter 条说明：
- LNK：healthy_links / total_links
- SEC：有 secrets 时显示加解密状态，无 secrets 时隐藏
- GIT：模块目录下的 git 变更情况（clean / N files changed）

### 4. Terminal（右侧）

操作输出面板 + 命令输入。

- 上方：操作历史输出（带��间戳，自动滚动，可上翻），缓冲区保留最近 1000 行
- 下方：`dot ▸` 输入提示符
- 按 `:` 聚焦输入模式，可键入 dot 子命令参数（如 `pull zsh`、`doctor`）
- 快捷键（p/P/d/x）执行时，命令和输出也显示在这里
- `esc` 退出输入模式
- 错误输出以红色（`#f7768e`）显示

Terminal 输入只接受简单参数（不启动交互式子 TUI），复杂交互场景退出 TUI 走 CLI。

### 5. Controls（底部全宽）

四个等宽操作按钮横排，作用于当前选中模块：

| 按钮 | 快捷键 | 作用 |
|------|--------|------|
| ⬇ PULL | p | 拉取/安装选中模块 |
| ⬆ PUSH | P | 提交并推送变更 |
| ⚕ DOCTOR | d | 检查选中模块健康 |
| ✕ REMOVE | x | 卸载选中模块（需确认） |

快捷键全局可用（不需���聚焦 Controls 面板）。执行期间快捷键禁用，Controls 显示执行中状态（spinner），完成后重新启用。

## Navigation

| 操作 | 按键 |
|------|------|
| 切换选中模块 | ←→ |
| 切换面板焦点 | tab |
| 聚焦 Terminal 输入 | `:` |
| 退出 Terminal 输入 | esc |
| 执行操作 | p / P / d / x（全局） |
| 添加模块 | a（退出 TUI，运行 `dot add` 的 CLI 流程） |
| 退出 TUI | q |

## Entry Point

裸 `dot` 命令（无子命令）进入 TUI。有子命令时（如 `dot pull zsh`）走普通 CLI 模式不变。

实现方式：在 `root.go` 的 `rootCmd` 上设置 `RunE`，检测无子命令时启动 TUI。

## Command Execution Architecture

### 核心问题

现有的 `runPull/runPush/runDoctor/runRemove` 直接写 `os.Stdout` 和读 `os.Stdin`，且部分命令内部启动独立的 bubbletea 程序（picker、passphrase）。这些在嵌套 TUI 中无法工作。

### 解决方案：提取核心逻辑为 io.Writer 函数

将每个命令拆为两层：
1. **核心逻辑函数**（新增）：接受 `io.Writer` 参数，所有输出写入 writer，不读 stdin，不启动 TUI。交互式确认改为参数控制（如 `force bool`）。
2. **cobra handler**（现有）：调用核心逻辑，使用 `os.Stdout`，保持 CLI 兼容。

```go
// 新增：核心逻辑，输出到 writer
func pullModule(dfPath string, modName string, w io.Writer) error { ... }
func pushChanges(dfPath string, msg string, w io.Writer) error { ... }
func doctorCheck(dfPath string, modName string, w io.Writer) error { ... }
func removeModule(dfPath string, modName string, w io.Writer) error { ... }

// 现有 cobra handler 改为调用核心逻辑
func runPull(cmd *cobra.Command, args []string) error {
    return pullModule(dfPath, name, os.Stdout)
}
```

Dashboard TUI 调用核心逻辑，传入 `bytes.Buffer` 捕获输出：

```go
func execDotCmd(fn func(io.Writer) error) tea.Cmd {
    return func() tea.Msg {
        var buf bytes.Buffer
        err := fn(&buf)
        return CmdOutputMsg{Output: buf.String(), Err: err}
    }
}
```

### 交互式操作处理

| 操作 | CLI 模式 | TUI 模式 |
|------|---------|---------|
| pull（指定模块） | 正常执行 | 核心逻辑直接执行，输出到 Terminal |
| pull（TUI picker） | 启动 picker | 不需要——Channel Strip 已选中模块 |
| push（确认提示） | stdin Y/n | Controls 按钮即确认，无需二次确认 |
| push（passphrase） | 启动 passphrase TUI | 退出 dashboard，运行 CLI `dot push`，完成后重新进入 |
| remove（确认） | stdin Y/n | 按 x 后 Controls 变为 "确认删除？ Y/N" |
| add | 需要路径参数 | 按 a 退出 TUI，提示用户运行 `dot add <path>` |

原则：**简单操作在 TUI 内完成，需要复杂交互（passphrase、文件路径）的退出 TUI 走 CLI**。

### 并发控制

- 同时只能执行一个命令（全局锁）
- 执行期间：快捷键禁用，Controls 按钮显示 spinner + "执行中..."
- 执行完成后：自动重新加载数据（`module.LoadAll` + `manifest.Load` + `gitpkg.Status`），所有面板刷新

## Data Flow

1. 启动时：`module.LoadAll` + `manifest.Load` + `gitpkg.Status` 获取所有数据
2. 面板渲染：基于加载的数据渲染各面板
3. 操作执行：快捷键 → `tea.Cmd` 异步调用核心逻辑 → 输出作为 `CmdOutputMsg` 发回 → Terminal 显示输出 → 自动 reload 数据 → 面板刷新

## Responsive Layout

- **最小终端尺寸**：80 列 × 24 行。窗口小于此尺寸时显示提示 "终端过小，请调整窗口大小（至少 80×24）"
- **面板尺寸分配**：
  - Overview：固定 20 列
  - Scope：自适应，占剩余宽度的 50%
  - Terminal：自适应，占剩余宽度的 50%
  - Channel Strip / Controls / Footer：全宽
- **`tea.WindowSizeMsg`**：Dashboard 监听并重新计算各面板尺寸，传入 `View(width, height)`

## Edge Cases

| 场景 | 处理 |
|------|------|
| 零模块（首次使���） | Channel Strip 只显示 `+ ADD a`，Scope 显示 "无模块，按 a 添加"，快捷键 p/P/d/x 禁用 |
| 模块加载部分失败 | 跳过损坏的模块，Terminal 显示警告，其他模块正常展示 |
| Channel Strip 溢出 | 横向滚动，显示 `◂` / `▸` 方向指示 |
| Terminal 输出过长 | 保留最近 1000 行，自动滚动到底部，可用 PgUp/PgDn 上翻 |
| Doctor 单模块 vs 全局 | Controls 的 DOCTOR 只检查选中模块；全局 doctor 通过 Terminal 输入 `doctor` 执行 |

## Visual Style

- 配色：Tokyo Night（与 kitty/nvim 统一）
- 背景：`#0f0f14`（深色）/ `#1a1b26`（面板）/ `#1e2030`（高亮）
- 边框：`#3b4261`（普通）/ 模块主色（选中，通过 lipgloss Border 实现）
- 文字：`#c0caf5`（正文）/ `#565f89`（次要）
- 状态色：`#9ece6a`（健康/绿）/ `#e0af68`（警告/黄）/ `#f7768e`（错误/红）
- 强调色：`#7aa2f7`（蓝/选中）/ `#bb9af7`（紫/push）

## Architecture

```
dot/internal/tui/
├── dashboard.go          # 主 Model，组合所有面板，路由消息
├── dashboard_view.go     # 主 View 渲染，布局计算
├── panel_channel.go      # Channel Strip 面板
├── panel_overview.go     # Overview 面板
├── panel_scope.go        # Scope 面板（含 meters）
├── panel_terminal.go     # Terminal 面板（输出 + 输入）
├── panel_controls.go     # Controls 面板
├── styles.go             # Tokyo Night 主题样式常量
├── commands.go           # execDotCmd + CmdOutputMsg + 核心逻辑调用
├── picker.go             # 已有：模块选择器（CLI 模式用）
├── passphrase.go         # 已有：密码输入（CLI 模式用）
```

### 关键类型

```go
// Panel 接口
type Panel interface {
    Update(msg tea.Msg) (Panel, tea.Cmd)
    View(width, height int) string
    Focused() bool
    SetFocus(bool)
}

// Dashboard 主模型
type Dashboard struct {
    channels   ChannelStrip
    overview   Overview
    scope      Scope
    terminal   Terminal
    controls   Controls
    focusIdx   int
    modules    []*module.Module
    manifest   *manifest.Manifest
    gitChanges []string
    selected   int
    executing  bool         // 命令执行中标记
    width      int          // 终端宽度
    height     int          // 终端高度
}

// 命令执行消息
type CmdOutputMsg struct {
    Output string
    Err    error
}

type CmdStartMsg struct{}
type DataReloadMsg struct {
    Modules    []*module.Module
    Manifest   *manifest.Manifest
    GitChanges []string
}
```

### 命令重构

需要重构的函数（提取核心逻辑）：

| 现有函数 | 新增核心函数 | 说明 |
|---------|------------|------|
| `runPull` | `pullModule(dfPath, modName string, w io.Writer) error` | 跳过 picker/passphrase 交互 |
| `runPush` | `pushChanges(dfPath, msg string, w io.Writer) error` | 跳过确认提示 |
| `runDoctor` | `doctorCheck(dfPath, modName string, w io.Writer) error` | 支持单模块过滤 |
| `runRemove` | `removeModule(dfPath, modName string, w io.Writer) error` | 跳过确认提示 |

现有 cobra handler 改为调用核心函数 + `os.Stdout`，保持 CLI 完全兼容。

## Dependencies

- `github.com/charmbracelet/bubbletea` — 已有
- `github.com/charmbracelet/lipgloss` — 已有
- 无新外部依赖
