# Neovide 配置

Neovide GUI 客户端的配置文件，提供无边框窗口和启动优化。

**核心特点**：
- 🖥️ **无边框窗口** - 隐藏标题栏和控制按钮
- 🚀 **启动优化** - 自动最大化窗口
- 🔗 **Neovim 集成** - 自动读取 Neovim 配置
- 🎨 **视觉统一** - 使用 Neovim 的 Tokyo Night 主题

---

## ⚙️ 工作原理

### 配置加载流程

```
Neovide 启动
    ↓
读取 ~/.config/neovide/config.toml
    ↓
应用窗口设置：
    - frame = "buttonless"（无边框）
    - title-hidden = true（隐藏标题）
    - maximized = true（启动时最大化）
    ↓
启动 Neovim
    ↓
Neovim 读取 ~/.config/nvim/lua/config/options.lua
    ↓
检测到 vim.g.neovide 变量
    ↓
应用 Neovide 专属配置：
    - 字体、光标特效、性能设置等
    ↓
Neovide + Neovim 完全就绪
```

### 配置分离

**为什么分两处配置？**
- `~/.config/neovide/config.toml` - Neovide **自身**的设置（窗口、启动）
- `~/.config/nvim/lua/config/options.lua` - Neovim **在 Neovide 中**的设置（字体、光标）

这样设计的好处：
- ✅ Neovim 配置在终端和 GUI 都能用
- ✅ Neovide 配置只影响 GUI 外观
- ✅ 职责清晰，互不干扰

---

## 📁 文件结构

```
neovide/
└── config.toml    # Neovide GUI 配置文件
```

---

## 🔧 配置说明

### config.toml

**窗口框架设置**：
```toml
frame = "buttonless"    # 窗口类型
```
可选值：
- `"buttonless"` - 无边框，隐藏控制按钮（当前使用）
- `"transparent"` - 完全透明边框
- `"none"` - 无装饰
- `"full"` - 完整窗口框架（默认）

**标题设置**：
```toml
title-hidden = true     # 隐藏标题文字（macOS 专用）
```

**窗口大小设置**：
```toml
maximized = true        # 启动时最大化
fullscreen = false      # 启动时不全屏
```

### Neovim 中的 Neovide 配置

在 `~/dotfiles/nvim/lua/config/options.lua` 中：

**字体配置**：
```lua
vim.o.guifont = "JetBrainsMono Nerd Font Mono:h18"
```

**光标特效**：
```lua
vim.g.neovide_cursor_vfx_mode = "ripple"  # 涟漪效果
vim.g.neovide_cursor_animation_length = 0.05
vim.g.neovide_cursor_trail_size = 0.3
```

可选特效：
- `"railgun"` - 轨道炮
- `"torpedo"` - 鱼雷
- `"pixiedust"` - 像素尘
- `"sonicboom"` - 音爆
- `"ripple"` - 涟漪（当前使用）
- `"wireframe"` - 线框
- `""` - 无特效

**性能设置**：
```lua
vim.g.neovide_refresh_rate = 60          # 刷新率 60fps
vim.g.neovide_refresh_rate_idle = 5      # 空闲时 5fps
vim.g.neovide_no_idle = false            # 允许空闲优化
```

**输入设置**：
```lua
vim.g.neovide_input_macos_option_key_is_meta = "both"  # Option 键作为 Meta
vim.g.neovide_input_ime = true                         # 启用输入法
vim.g.neovide_hide_mouse_when_typing = true           # 输入时隐藏鼠标
```

---

## 🚀 安装

### 前置要求

```bash
# 1. 安装 Neovide
brew install --cask neovide

# 2. 确保 Neovim 已配置
# 参见 ~/dotfiles/nvim/README.md
```

### 应用配置

```bash
# 创建符号链接
ln -s ~/dotfiles/neovide ~/.config/neovide
```

### 启动 Neovide

**方法 1：使用 nvide 命令**（推荐）
```bash
# 打开当前目录
nvide

# 打开指定文件或目录
nvide /path/to/file
nvide /path/to/directory
```

**方法 2：直接启动**
```bash
# macOS
open -a Neovide

# 或命令行
neovide
```

---

## ✏️ 自定义指南

### 调整窗口样式

编辑 `config.toml`：

**使用透明边框**：
```toml
frame = "transparent"
```

**启用全屏启动**：
```toml
fullscreen = true
```

**不自动最大化**：
```toml
maximized = false
```

### 调整字体大小

编辑 `~/dotfiles/nvim/lua/config/options.lua`：
```lua
-- 在 if vim.g.neovide then 块中修改
vim.o.guifont = "JetBrainsMono Nerd Font Mono:h20"  # 改为 20pt
```

### 更换字体

```lua
-- 使用其他 Nerd Font
vim.o.guifont = "FiraCode Nerd Font Mono:h18"
vim.o.guifont = "Hack Nerd Font Mono:h18"
```

### 调整光标特效

```lua
-- 更换特效
vim.g.neovide_cursor_vfx_mode = "pixiedust"

-- 调整动画速度
vim.g.neovide_cursor_animation_length = 0.1  # 更慢

-- 关闭特效
vim.g.neovide_cursor_vfx_mode = ""
```

### 性能优化

**提升刷新率**（高性能 Mac）：
```lua
vim.g.neovide_refresh_rate = 120  # 120fps
```

**降低性能消耗**（低配 Mac）：
```lua
vim.g.neovide_refresh_rate = 30            # 30fps
vim.g.neovide_refresh_rate_idle = 1        # 空闲 1fps
vim.g.neovide_cursor_animation_length = 0  # 关闭光标动画
vim.g.neovide_cursor_vfx_mode = ""         # 关闭光标特效
```

---

## 💡 使用技巧

### nvide 启动器

`nvide` 是自定义脚本，位于 `~/dotfiles/bin/nvide`：

```bash
#!/bin/zsh
if [ $# -eq 0 ]; then
  open -na Neovide --args "$(pwd)"
else
  open -na Neovide --args "$@"
fi
```

**优点**：
- 独立于终端运行（不会随终端关闭）
- 支持多实例（`-n` 参数）
- 自动传递参数给 Neovide

### 快捷键

Neovide 支持所有 Neovim 快捷键，加上自定义的：
- `Cmd+E` - 切换文件浏览器
- `Cmd+/` - 切换浮动终端
- `Cmd+T` - 切换右侧终端

（详见 `~/dotfiles/nvim/README.md`）

### 从命令行打开文件

```bash
# 在 Neovide 中打开文件
nvide ~/.zshrc

# 在 Neovide 中打开项目
nvide ~/my-project
```

---

## 📝 注意事项

1. **依赖 Neovim 配置**：Neovide 使用 Neovim 配置，必须先配置 Neovim
2. **macOS 专用**：某些选项（如 `title-hidden`）仅在 macOS 上有效
3. **字体要求**：必须安装 Nerd Font，否则图标显示异常
4. **性能影响**：光标特效和高刷新率会消耗 GPU 资源
5. **多实例**：使用 `nvide` 启动支持多实例，直接启动 Neovide 不支持

---

## 🔗 参考

- [Neovide 官方文档](https://neovide.dev/)
- [Neovide 配置参考](https://neovide.dev/configuration.html)
- [Neovim 配置](../nvim/README.md)

---

[← 返回主 README](../README.md)
