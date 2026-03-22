# 实现对齐设计

将代码实现对齐到设计文档。分两批执行：CLI 命令重构 → TUI 仪表盘对齐。

---

## 第一批：CLI 命令重构

### C1: pull / install 分离

当前 `pull` 同时执行 git pull 和模块安装，违反设计中的职责分离。

| 命令 | 改后职责 | 参数 |
|------|---------|------|
| `dot pull` | 仅 `git pull` 拉取远端代码 | 无 |
| `dot install [modules...]` | 安装模块（符号链接+依赖+secrets） | `--all` 安装全部 |
| `dot install` | 无参数时 TUI 选择器 | — |

**变更：**
- 创建 `cmd/install.go`：从 pull.go 移出安装逻辑
- 简化 `cmd/pull.go`：仅保留 git pull
- TUI 仪表盘 PULL 按钮改为调用 install 逻辑

### C2+C3: uninstall / remove 分离

当前 `remove` 做的是 uninstall 的事（删符号链接、保留文件）。

| 命令 | 职责 | 可逆性 |
|------|------|--------|
| `dot uninstall <modules...>` | 删除符号链接，恢复备份，保留模块文件 | 可重新 install |
| `dot remove <modules...>` | 从 `modules/` 目录删除模块文件 | 不可逆 |

**变更：**
- 重命名 `cmd/remove.go` → `cmd/uninstall.go`，命令名改为 `uninstall`
- 创建新 `cmd/remove.go`：检查已 uninstall → 删除 `modules/<name>/` 目录
- 重命名 `internal/ops/remove.go` → `internal/ops/uninstall.go`
- 创建新 `internal/ops/remove.go`：模块文件删除逻辑
- TUI 仪表盘 REMOVE 按钮改为调用 uninstall 逻辑

**remove 前置条件：**
- 模块必须未安装（已 uninstall），否则提示先 uninstall
- 二次确认："将永久删除 modules/<name>/ 目录，确认？"

### C4: clean 补全

当前 `clean` 仅清理备份文件，缺两项。

| 清理项 | 当前状态 | 说明 |
|--------|---------|------|
| 备份文件 | ✓ 已实现 | `*.backup.*` 文件 |
| 孤立链接 | ✗ 缺失 | manifest 中记录的链接，目标已不存在或源已不存在 |
| 无效 manifest | ✗ 缺失 | manifest 中记录的模块，`modules/` 目录中已不存在 |

**变更：**
- `cmd/clean.go` 添加孤立链接扫描和清理
- `cmd/clean.go` 添加无效 manifest 记录清理
- 三项分别列出、分别确认

---

## 第二批：TUI 仪表盘对齐

### T1: 芯片固定尺寸

| 属性 | 当前 | 改为 |
|------|------|------|
| ChipW | 动态计算 `(totalW - n - 1) / n` | 固定 8 |
| ChipH | `ChipW / 2` | 固定 4 |
| 溢出处理 | 无 | 横向滚动 |

**变更文件：** `layout.go`

### T2: 滚动指示符

- Channel Strip 左右边缘显示 ◂ / ▸ 指示符
- 仅在芯片溢出且对应方向可滚动时显示

**变更文件：** `panel_channel.go`

### T3: 健康度公式

当前：`healthy_links / total_links × 100`
改为：`(healthy_links + healthy_secrets) / (total_links + total_secrets) × 100`

**变更文件：** `panel_overview.go`

### T4: Overview 补充字段

| 字段 | 数据来源 |
|------|---------|
| shell | `$SHELL` 环境变量，取 basename |
| pkg | 检测 brew / apt 可用性 |
| font | 固定值 "JetBrainsMono NF" |

**变更文件：** `panel_overview.go`

### T5: sync 完整显示

当前：显示 "X change(s)" 或 "clean"
改为：`git rev-list --count HEAD..@{u}` / `@{u}..HEAD` 获取 behind/ahead，显示 "2↑ 3↓"

**变更文件：** `panel_overview.go`

### T6: Terminal 命令历史

- 添加 `[]string` 历史缓冲区
- ↑ 向前翻阅历史，↓ 向后翻阅
- 执行命令后追加到历史

**变更文件：** `panel_terminal.go`

### T7: Remove 二次确认

1. 按 `x` → Controls 栏变为 "确认删除 [模块名]？Y/N"
2. 按 `Y` → 执行 uninstall
3. 按 `N` / `Esc` → 取消，恢复正常按钮

**变更文件：** `dashboard.go`, `panel_controls.go`

### T8: Terminal 输入模式 Ctrl 快捷键

Terminal 输入模式下，操作快捷键需加 Ctrl 前缀：

| 正常模式 | 输入模式 |
|---------|---------|
| p | Ctrl+p |
| P | Ctrl+P |
| d | Ctrl+d |
| x | Ctrl+x |
| a | Ctrl+a |
| q | Ctrl+q |

**变更文件：** `dashboard.go`

### T9: 响应式降级

终端尺寸不足时按优先级隐藏面板：

| 优先级 | 隐藏面板 | 触发条件 |
|--------|---------|---------|
| 1（最先隐藏） | Footer | 高度 < 24 |
| 2 | Overview | 宽度 < 100 |
| 3 | Controls | 高度 < 20 |

核心面板（Channel Strip + Scope + Terminal）始终保留。

**变更文件：** `layout.go`, `dashboard_view.go`

---

## 实施顺序

1. **第一批 CLI：** C2+C3 (uninstall/remove) → C1 (pull/install) → C4 (clean)
2. **编译验证 + 测试**
3. **第二批 TUI：** T1+T2 (芯片+滚动) → T3+T4+T5 (Overview) → T6 (历史) → T7+T8 (交互) → T9 (降级)
4. **编译验证 + 手动测试**
