# Neovim 配置

基于 [LazyVim](https://www.lazyvim.org/) 的模块化 Neovim 配置，采用单一符号链接设计，最小侵入。

**核心特点**：
- 🔗 **单一入口点** - 只需一个符号链接，LazyVim 自动加载
- 🧩 **模块化设计** - 配置和插件分离，自动发现
- 🎨 **Tokyo Night 主题** - 与终端统一配色
- 🚀 **零侵入** - 与 LazyVim 完全隔离，互不干扰

---

## ⚙️ 工作原理

### 加载流程
```
LazyVim 启动
    ↓
扫描 ~/.config/nvim/lua/plugins/*.lua
    ↓
发现 dotfiles.lua（符号链接）
    → 指向 ~/dotfiles/nvim/settings.lua
    ↓
settings.lua 执行：
    1. 扫描 ~/dotfiles/nvim/lua/config/*.lua
       → 加载配置覆盖（keymaps.lua、options.lua）
    2. 扫描 ~/dotfiles/nvim/lua/plugins/*.lua
       → 加载插件规范（tokyonight.lua、autoread.lua）
    ↓
所有配置叠加到 LazyVim 之上
```

### 设计理念

**为什么用单一符号链接？**
- ✅ 最小侵入：只需一个文件，不污染 LazyVim 配置
- ✅ 完全隔离：dotfiles 配置独立管理，随时可移除
- ✅ 自动发现：添加新配置无需修改 settings.lua

**为什么模块化？**
- ✅ 职责分离：配置覆盖和插件规范分开管理
- ✅ 易于维护：每个文件功能单一，修改方便
- ✅ 按需加载：LazyVim 的 lazy loading 机制自动优化

---

## 📁 文件结构

```
nvim/
├── settings.lua         # 统一加载器（符号链接入口）
└── lua/
    ├── config/          # 配置覆盖（LazyVim config/）
    │   ├── keymaps.lua  # 自定义快捷键
    │   └── options.lua  # Neovide GUI 配置
    └── plugins/         # 自定义插件（LazyVim plugins/）
        ├── tokyonight.lua  # Tokyo Night 主题
        └── autoread.lua    # 文件自动刷新
```

---

## 🔧 配置说明

### 🎨 主题
- **Tokyo Night (night 变体)** - 深蓝黑背景 `#1a1b26`
- 与 Kitty 终端主题统一

### ⌨️ 自定义快捷键
| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `Cmd+E` | 切换文件浏览器 | 所有模式可用 |
| `Cmd+/` | 切换浮动终端 | 继承当前目录 |
| `Cmd+T` | 切换右侧终端 | 占屏幕 40% 宽度 |

### 🖥️ Neovide 配置
- **字体**：JetBrainsMono Nerd Font Mono (18pt)
- **光标特效**：Ripple（涟漪效果）
- **性能**：60fps 刷新率，空闲时 5fps
- **输入**：Option 键作为 Meta，支持输入法

### 🔄 文件自动刷新
- **激进模式**：200ms 检测间隔
- **自动触发**：焦点切换、光标移动、定时检查
- **静默加载**：外部修改自动刷新，无需确认

---

## 🚀 安装

### 前置要求

```bash
# 1. 安装 Neovim
brew install neovim

# 2. 安装 LazyVim starter
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git
```

### 应用配置

**只需一个符号链接**：

```bash
ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua
```

### 首次启动

```bash
nvim
```

LazyVim 会自动：
1. 安装 lazy.nvim 插件管理器
2. 加载 dotfiles 配置
3. 下载并安装所有插件（Tokyo Night、autoread 等）

### 验证安装

```vim
" 在 Neovim 中运行
:Lazy
```

应该看到：
- `dotfiles-config-loader` - 配置加载器
- `tokyonight.nvim` - Tokyo Night 主题
- `autoread-config` - 自动刷新插件

---

## ✏️ 自定义指南

### 添加新配置

**添加配置覆盖**：
```bash
# 创建新的配置文件
nvim ~/dotfiles/nvim/lua/config/autocmds.lua
```

settings.lua 会自动发现并加载，无需修改任何文件。

**添加新插件**：
```bash
# 创建新的插件规范文件
nvim ~/dotfiles/nvim/lua/plugins/my-plugin.lua
```

示例插件配置：
```lua
return {
  "author/plugin-name",
  lazy = false,
  opts = {
    -- 插件选项
  },
}
```

### 修改快捷键

编辑 `lua/config/keymaps.lua`：
```lua
vim.keymap.set("n", "<D-k>", "<cmd>echo 'Hello'<cr>", { desc = "My Keymap" })
```

### 修改主题

编辑 `lua/plugins/tokyonight.lua`：
```lua
opts = {
  style = "storm",  -- 改为 storm 变体
  transparent = true,  -- 启用透明背景
}
```

---

## 📝 注意事项

1. **LazyVim 依赖**：必须先安装 LazyVim starter，否则配置无法工作
2. **符号链接路径**：确保符号链接指向正确的路径
3. **插件冲突**：避免与 LazyVim 默认插件冲突
4. **Neovim 版本**：需要 Neovim >= 0.9.0

---

## 🔗 参考

- [LazyVim 官方文档](https://www.lazyvim.org/)
- [LazyVim 配置指南](https://www.lazyvim.org/configuration)
- [Tokyo Night 主题](https://github.com/folke/tokyonight.nvim)

---

[← 返回主 README](../README.md)
