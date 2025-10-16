# 自定义脚本

存放自定义命令行工具和脚本的目录，所有脚本通过 PATH 全局可用。

**核心特点**：
- 🛠️ **全局可用** - 已添加到 PATH，可在任何位置执行
- 📝 **易于扩展** - 添加新脚本即可使用
- 🔒 **可执行权限** - 所有脚本已设置执行权限
- 🐚 **Zsh 优化** - 针对 Zsh Shell 编写

---

## ⚙️ 工作原理

### PATH 机制

```
Zsh 启动
    ↓
加载 ~/.zsh/env.zsh
    ↓
设置 PATH：
    export PATH="$HOME/dotfiles/bin:$PATH"
    ↓
~/dotfiles/bin 中的所有可执行文件
    ↓
可在任何位置直接执行（无需完整路径）
```

### 为什么放在 bin/ ？

- ✅ **符合惯例**：Unix/Linux 标准目录结构
- ✅ **集中管理**：所有自定义脚本在一处
- ✅ **版本控制**：随 dotfiles 一起管理
- ✅ **易于同步**：换机器只需克隆仓库

---

## 📁 文件结构

```
bin/
└── nvide    # Neovide GUI 启动器
```

---

## 🔧 脚本说明

### nvide - Neovide 启动器

**功能**：在独立进程中启动 Neovide GUI，不依赖终端。

**实现**：
```bash
#!/bin/zsh
# 使用 macOS 'open' 命令独立启动 Neovide

if [ $# -eq 0 ]; then
  # 无参数：打开当前目录
  open -na Neovide --args "$(pwd)"
else
  # 有参数：传递给 Neovide
  open -na Neovide --args "$@"
fi
```

**参数说明**：
- `-n` - 打开新实例（允许多窗口）
- `-a Neovide` - 指定应用名称
- `--args` - 后续参数传递给应用

**使用方法**：
```bash
# 打开当前目录
nvide

# 打开指定文件
nvide ~/.zshrc

# 打开指定目录
nvide ~/projects/my-app

# 打开多个文件
nvide file1.txt file2.txt
```

**为什么不直接用 neovide 命令？**
- ❌ `neovide` - 绑定终端，关闭终端会退出
- ✅ `nvide` - 独立进程，终端可以关闭
- ✅ `nvide` - 支持多实例（同时打开多个窗口）
- ✅ `nvide` - 传递当前目录，更符合使用习惯

---

## 🚀 安装

### 前置要求

脚本已包含在 dotfiles 中，只需确保：
- Zsh 配置已应用（`~/.zsh/env.zsh` 已加载）
- PATH 包含 `~/dotfiles/bin`

### 验证安装

```bash
# 1. 检查 PATH
echo $PATH | grep dotfiles
# 应输出包含 /Users/xxx/dotfiles/bin 的路径

# 2. 检查脚本位置
which nvide
# 应输出: /Users/xxx/dotfiles/bin/nvide

# 3. 测试执行
nvide --help  # 或任何参数，验证脚本可执行
```

---

## ✏️ 添加新脚本

### 1. 创建脚本文件

```bash
# 在 bin/ 目录创建新脚本
nvim ~/dotfiles/bin/my-script

# 或使用 touch 创建
touch ~/dotfiles/bin/my-script
```

### 2. 编写脚本内容

**示例 1：简单命令脚本**
```bash
#!/bin/zsh
# my-script - 我的自定义脚本

echo "Hello from my-script!"
echo "Arguments: $@"
```

**示例 2：带参数处理的脚本**
```bash
#!/bin/zsh
# git-quick - 快速 Git 提交

if [ $# -eq 0 ]; then
  echo "Usage: git-quick <commit-message>"
  exit 1
fi

git add .
git commit -m "$*"
git push
```

**示例 3：交互式脚本**
```bash
#!/bin/zsh
# project-open - 快速打开项目

PROJECTS_DIR=~/projects

echo "Available projects:"
ls -1 "$PROJECTS_DIR"

echo -n "\nEnter project name: "
read project

if [ -d "$PROJECTS_DIR/$project" ]; then
  cd "$PROJECTS_DIR/$project"
  nvide
else
  echo "Project not found: $project"
fi
```

### 3. 设置执行权限

```bash
chmod +x ~/dotfiles/bin/my-script
```

### 4. 测试脚本

```bash
# 直接执行（因为已在 PATH 中）
my-script

# 或使用完整路径
~/dotfiles/bin/my-script
```

### 5. 提交到 Git

```bash
cd ~/dotfiles
git add bin/my-script
git commit -m "Add my-script"
git push
```

---

## 💡 最佳实践

### 脚本命名

- ✅ **使用小写**：`my-script`（推荐）
- ✅ **用连字符分隔**：`git-quick`、`project-open`
- ❌ **避免下划线**：`my_script`（不推荐）
- ❌ **避免空格**：`my script`（无效）

### Shebang 选择

```bash
#!/bin/zsh       # Zsh 脚本（推荐，支持 Zsh 特性）
#!/bin/bash      # Bash 脚本（兼容性更好）
#!/usr/bin/env zsh   # 可移植 Zsh（推荐跨平台）
```

### 错误处理

```bash
#!/bin/zsh
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时报错
set -o pipefail  # 管道中任何命令失败都返回失败

# 你的脚本逻辑
```

### 帮助信息

```bash
#!/bin/zsh

# 检查 --help 参数
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
  cat <<EOF
Usage: my-script [options] <args>

Description:
  My custom script does something useful.

Options:
  -h, --help    Show this help message

Examples:
  my-script foo
  my-script --option bar
EOF
  exit 0
fi

# 你的脚本逻辑
```

---

## 📝 注意事项

1. **执行权限**：新脚本必须设置 `chmod +x`
2. **Shebang**：第一行必须是 `#!/bin/zsh` 或其他解释器
3. **PATH 优先级**：`~/dotfiles/bin` 在 PATH 前面，优先于系统命令
4. **名称冲突**：避免与系统命令同名（如 `ls`、`cd`）
5. **跨平台**：如需跨平台，使用 `#!/usr/bin/env zsh` 而非 `#!/bin/zsh`

---

## 🔗 参考

- [Zsh 脚本编写指南](http://zsh.sourceforge.net/Guide/zshguide.html)
- [Bash 脚本教程](https://www.gnu.org/software/bash/manual/)
- [Shell 脚本最佳实践](https://google.github.io/styleguide/shellguide.html)

---

[← 返回主 README](../README.md)
