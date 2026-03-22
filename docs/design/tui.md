# TUI 仪表盘

裸 `dot` 命令进入，合成器风格多面板仪表盘。

---

## 布局

```
┌────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┐
│        │        │        │        │        │        │        │        │
│  BIN   │  GIT   │ KITTY  │NEOVIDE │  NVIM  │  YAZI  │  ZSH   │ + ADD  │
│   ●    │   ●    │   ●    │   ●    │   ●    │   ●    │  ●     │   a    │
│        │        │        │        │        │        │        │        │
├────────┴──┬─────┴────────┴────────┴────────┴──┬─────┴────────┴────────┤
│ OVERVIEW  │ SCOPE                             │ TERMINAL              │
│           │                                   │                       │
├───────────┼─────────┬───────────┬─────────────┼───────────────────────┤
│ INSTALL p │ PUSH P  │ DOCTOR d  │  UNINSTALL x│                       │
├───────────┴─────────┴───────────┴─────────────┴───────────────────────┤
│ ←→ module · : terminal · p install · P push · d doctor · x uninstall · q │
└───────────────────────────────────────────────────────────────────────┘
```

---

## 面板

| 面板 | 位置 | 内容 |
|------|------|------|
| Channel Strip | 顶部 | 模块芯片网格 |
| Overview | 中左 | 全局状态：健康度、统计、系统环境 |
| Scope | 中央 | 选中模块详情：进度条、链接路径、依赖 |
| Terminal | 中右 | 命令输出 + 命令输入（仅 dot 子命令） |
| Controls | 底部 | 操作按钮 |
| Footer | 最底 | 快捷键提示 |

---

## 芯片

| 属性 | 值 |
|------|-----|
| 内容尺寸 | 8 列 × 4 行（固定，绝对不变） |
| 含边框尺寸 | 10 × 6 |
| 形状 | 视觉正方形（���端字符 1:2 比例） |
| 内容排布 | 名字 + LED 垂直居中 |
| 选中状态 | 反色（lipgloss Reverse） |
| 溢出处理 | 横向滚动，显示滚动指示 |
| 启动默认 | 选中第一个模块 |

---

## 面板尺寸

| 面板 | 宽度 | 高度 |
|------|------|------|
| Overview | 总宽 1/4 | 自适应剩余空间 |
| Scope | 总宽 2/4 | 同 Overview |
| Terminal | 总宽 1/4 | 同 Overview |
| Controls | 4 等分 | 1 行 |
| Footer | 全宽 | 1 行 |

- 三个中间面板高度相同

---

## Overview 内容

### 健康度

- 计算公式：`(healthy_links + healthy_secrets) / (total_links + total_secrets) × 100`
- 远端有更新但本地未拉取时也算不健康
- 显示为进度条 + 百分比

### 统计

| 数字 | 说明 |
|------|------|
| modules | 模块总数 |
| installed | 已安装数 |
| secrets | secrets 总数 |
| changed | git 变更文件数 |

### 系统环境

| 项目 | 示例 |
|------|------|
| os | darwin |
| shell | zsh |
| pkg | brew |
| branch | main |
| sync | clean / 2 behind / 3 ahead |
| key | available / unavailable |
| font | JetBrainsMono NF |

---

## Scope 内容

- 启动时默认选中第一个模块
- Meter box（进度条区域）保留内部边框
- Signal Path 数据来自 manifest 记录（展开后的实际链接，非 module.toml 原始定义）

---

## 边框

| 属性 | 值 |
|------|-----|
| 结构 | 一个外框 + 内部分割线 |
| 风格 | 直角（不用圆角） |
| 外框 | ┌ ┐ └ ┘ ─ │ |
| 行分割 | ├ ─ ┤ |
| 不同列数行连接 | ┴（上行竖线结束）┬（下行竖线开始）┼（交叉）|
| 颜色 | ANSI "8"（dimmed，跟随终端主题）|

---

## 操作栏

| 按钮 | 图标 | 快捷键 |
|------|------|--------|
| INSTALL | \uf0ed | p |
| PUSH | \uf0ee | P |
| DOCTOR | \uf21e | d |
| UNINSTALL | \uf1f8 | x（需二次确认） |

- 4 列等分
- 各列之间有竖线分割
- 高度 1 行
- 执行期��显示 spinner

### Uninstall 确认

按 x 后操作栏变为"确认卸载？Y/N"，再按 Y 才执行。

---

## Add 操作

- 按 a 打开 bubbletea filepicker 文件浏览器
- 选择目录后用目录名作为模块名
- 自动执行 add 流程

---

## 交互

### 快捷键

| 操作 | 非输入模式 | Terminal 输入模式 |
|------|-----------|-----------------|
| 切换模块 | ←→ | Ctrl+←→ |
| 聚焦 Terminal | `:` | — |
| 退出 Terminal | — | esc |
| 翻历史命令 | — | ↑↓ |
| Install | p | Ctrl+p |
| Push | P | Ctrl+u |
| Doctor | d | Ctrl+d |
| Uninstall | x | Ctrl+x |
| Add | a | Ctrl+a |
| 退出 | q | Ctrl+q |

### 规则

- 非输入模式：单键触发快捷键
- Terminal 输入模式：所有按键都是输入，快捷键需加 Ctrl
- Terminal 支持 ↑↓ 翻历史命令

---

## 执行反馈

- 命令输出显示在 Terminal 面板
- 执行期间操作栏显示 spinner
- 执行完成后自动刷新所有面板数据（模块、manifest、git status）

---

## 响应式

| 情况 | 处理 |
|------|------|
| 芯片溢出 | 横向滚动 |
| 终端过小 | 降级渲染，按优先级隐藏面板 |

### 降级顺序（从最先隐藏到最后隐藏）

1. Footer
2. Overview
3. Controls
4. Channel Strip / Scope / Terminal 始终保留

---

## 主题

详见 `theme.md`

| 属性 | 值 |
|------|-----|
| 配色 | ANSI 0-15 调色板，跟随终端主题 |
| 图标 | Nerd Font only |
