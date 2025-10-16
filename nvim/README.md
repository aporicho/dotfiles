# Neovim 自定义配置

这是我的 Neovim 自定义配置，基于 [LazyVim](https://www.lazyvim.org/)。

**重要**: 此目录只包含我的个人配置，不包含 LazyVim 的默认文件。

## 📁 目录结构

```
nvim/
├── settings.lua         # 统一加载器（会被符号链接到 LazyVim）
└── lua/
    ├── config/          # LazyVim 配置覆盖
    │   ├── keymaps.lua  # 自定义快捷键
    │   └── options.lua  # Neovide GUI 配置
    └── plugins/         # 自定义插件
        ├── autoread.lua # 文件自动刷新
        └── tokyonight.lua # Tokyo Night 主题
```

**设计理念**：
- `settings.lua` 作为统一入口，自动加载 `config/` 和 `plugins/` 中的所有内容
- 只需要**一个符号链接**，最小侵入 LazyVim 配置
- 所有自定义配置都在 dotfiles 中管理，与 LazyVim 完全隔离

## ✨ 主要配置

### 🎨 主题
- **Tokyo Night** (night 变体) - 与 Kitty 终端主题一致

### ⌨️ 自定义快捷键
- `Command+e` - 切换文件浏览器
- `Command+/` - 切换浮动终端
- `Command+t` - 切换右侧终端

### 🖥️ Neovide 配置
- 字体: JetBrainsMono Nerd Font Mono (18pt)
- 光标特效: Ripple
- 完整的性能和输入优化

### 🔄 文件自动刷新
- 激进模式自动刷新，实时监测文件变化

## 🚀 安装

### 前置要求
1. 安装 Neovim (>= 0.9.0)
2. 安装 LazyVim starter

```bash
# 使用 LazyVim starter 初始化 Neovim 配置
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git
```

### 应用自定义配置

**只需要一个符号链接**：

```bash
ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua
```

### 首次启动
启动 Neovim，LazyVim 会自动：
1. 安装 lazy.nvim 插件管理器
2. 通过 `dotfiles.lua` 加载你的所有自定义配置
3. 下载并安装所有插件

```bash
nvim
```

### 验证安装

启动 Neovim 后，运行以下命令验证：
```vim
:Lazy
```
你应该能看到 `dotfiles-config-loader` 和你的自定义插件（Tokyo Night、autoread）

## 📝 说明

- **无侵入设计**：只需一个符号链接，不会覆盖 LazyVim 核心文件
- **完全隔离**：dotfiles 配置与 LazyVim 独立，互不干扰
- **自动加载**：`settings.lua` 会自动发现并加载 `config/` 和 `plugins/` 中的所有文件
- **易于维护**：添加新配置无需修改 settings.lua，直接在对应目录创建 `.lua` 文件即可

### 工作原理

1. LazyVim 启动时会自动加载 `~/.config/nvim/lua/plugins/*.lua`
2. `dotfiles.lua`（符号链接）指向 `~/dotfiles/nvim/settings.lua`
3. `settings.lua` 执行时会：
   - 加载 `~/dotfiles/nvim/lua/config/*.lua`（配置覆盖）
   - 加载 `~/dotfiles/nvim/lua/plugins/*.lua`（插件规范）
4. 所有配置叠加到 LazyVim 之上

## 🔗 参考

- [LazyVim 官方文档](https://www.lazyvim.org/)
- [LazyVim GitHub](https://github.com/LazyVim/LazyVim)
