# Dotfiles

个人配置文件和自定义脚本

## bin/

自定义命令脚本目录

### nvide

使用 Neovide 打开当前目录或指定路径的启动器

```bash
# 打开当前目录
nvide

# 打开指定文件或目录
nvide /path/to/file
nvide /path/to/directory
```

## 安装

```bash
# 克隆仓库
git clone <your-repo-url> ~/dotfiles

# 添加到 PATH（在 ~/.zsh/env.zsh 中）
export PATH="$HOME/dotfiles/bin:$PATH"

# 重新加载配置
source ~/.zshrc
```
