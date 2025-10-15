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

Neovim 配置（基于 LazyVim）

- **主题**: Tokyo Night
- **插件管理**: lazy.nvim
- **LSP**: 完整的 Language Server Protocol 支持
- **自定义快捷键**:
  - `Command+e` - 切换文件浏览器
  - `Command+/` - 切换浮动终端
  - `Command+t` - 切换右侧终端

配置文件位于 `~/.config/nvim/`（符号链接）

### neovide/

Neovide GUI 配置

- 无边框窗口模式
- 启动时最大化
- 与 Neovim 配置集成

配置文件位于 `~/.config/neovide/`（符号链接）

## 🚀 安装

```bash
# 克隆仓库
git clone <your-repo-url> ~/dotfiles

# 创建符号链接
ln -s ~/dotfiles/nvim ~/.config/nvim
ln -s ~/dotfiles/neovide ~/.config/neovide

# 添加 bin 到 PATH（在 ~/.zsh/env.zsh 中）
export PATH="$HOME/dotfiles/bin:$PATH"

# 重新加载配置
source ~/.zshrc
```

## 🔄 备份说明

所有配置文件通过符号链接管理，修改会自动同步到 git 仓库。原始备份保存在：
- `~/.config/nvim.backup`
- `~/.config/neovide.backup`
