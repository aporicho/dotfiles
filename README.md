# Dotfiles

个人配置文件和自定义脚本

## 📁 目录结构

### bin/

自定义命令脚本目录

#### nvide

使用 Neovide 打开当前目录或指定路径的启动器

```bash
# 打开当前目录
nvide

# 打开指定文件或目录
nvide /path/to/file
nvide /path/to/directory
```

### nvim/

Neovim 自定义配置（基于 LazyVim）

**设计理念**: 单一符号链接，最小侵入，完全隔离

- **主题**: Tokyo Night (night 变体)
- **自定义快捷键**:
  - `Command+e` - 切换文件浏览器
  - `Command+/` - 切换浮动终端
  - `Command+t` - 切换右侧终端
- **Neovide**: 完整的 GUI 优化配置
- **自动刷新**: 激进模式文件监测

配置结构：
- `settings.lua` - 统一加载器（单一入口）
- `lua/config/` - 配置覆盖（keymaps、options）
- `lua/plugins/` - 自定义插件（主题、功能扩展）

详见 [nvim/README.md](nvim/README.md) 了解工作原理和安装方法

### kitty/

Kitty 终端自定义配置（模块化设计）

- **主题**: Tokyo Night - 与 Neovim 统一视觉体验
- **性能优化**: 0ms 输入延迟、2ms 重绘延迟
- **快捷键**: macOS 风格 Command 键操作
- **Shell 集成**: 命令跳转、输出复制、工作目录跟踪
- **会话管理**: 预设开发/工作环境布局

配置内容：
- `kitty.conf` - 主配置（模块加载器）
- `performance.conf` - 性能优化
- `layouts.conf` - 窗口布局
- `keymaps.conf` - 快捷键配置
- `shell_integration.conf` - Shell 集成
- `modern_features.conf` - 现代功能
- `kitty.zsh` - Zsh 集成脚本
- `themes/` - 主题配色
- `sessions/` - 会话文件

详见 [kitty/README.md](kitty/README.md) 了解安装方法和功能说明

### neovide/

Neovide GUI 配置

- 无边框窗口模式
- 启动时最大化
- 与 Neovim 配置集成

配置文件位于 `~/.config/neovide/`（符号链接）

### zsh/

Zsh Shell 配置（模块化设计）

- **主配置**: `.zshrc` - 主入口文件
- **模块目录**: `config/` - 包含各个功能模块
  - `env.zsh` - 环境变量配置
  - `options.zsh` - Shell 选项配置
  - `history.zsh` - 历史记录配置
  - `completion.zsh` - 补全系统配置
  - `aliases.zsh` - 别名定义

配置文件通过符号链接：
- `.zshrc` → `~/.zshrc`
- `config/` → `~/.zsh/`

## 🚀 安装

### 1. 克隆仓库
```bash
git clone <your-repo-url> ~/dotfiles
```

### 2. Neovim 配置
**重要**: 需要先安装 LazyVim starter

```bash
# 1. 安装 LazyVim starter
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git

# 2. 创建单一符号链接（无侵入设计）
ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua

# 3. 启动 Neovim，自动安装插件
nvim
```

详见 [nvim/README.md](nvim/README.md) 了解工作原理

### 3. Kitty 终端配置

```bash
# 安装 Kitty（如未安装）
brew install --cask kitty

# 使用符号链接
ln -s ~/dotfiles/kitty ~/.config/kitty

# 验证配置
~/dotfiles/kitty/verify.sh
```

详见 [kitty/README.md](kitty/README.md)

### 4. 其他配置
```bash
# Neovide GUI
ln -s ~/dotfiles/neovide ~/.config/neovide

# Zsh
ln -s ~/dotfiles/zsh/.zshrc ~/.zshrc
ln -s ~/dotfiles/zsh/config ~/.zsh

# 重新加载 shell
source ~/.zshrc
```

**注意**:
- PATH 配置已包含在 `zsh/config/env.zsh` 中，会自动加载
- 自定义脚本路径: `~/dotfiles/bin`

## 🔄 备份说明

所有配置文件通过符号链接管理，修改会自动同步到 git 仓库。原始备份保存在：
- `~/.config/nvim.backup`
- `~/.config/kitty.backup`
- `~/.config/neovide.backup`
- `~/.zshrc.backup`
- `~/.zsh.backup`
