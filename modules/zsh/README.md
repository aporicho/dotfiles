# Zsh 配置

模块化的 Zsh Shell 配置，采用功能分离设计，易于维护和扩展。

**核心特点**：
- 🧩 **模块化设计** - 按功能分离配置文件
- 🚀 **性能优化** - 补全缓存、延迟加载
- 🎨 **增强功能** - 自动建议、语法高亮、智能补全
- 🔧 **易于扩展** - 添加新模块无需修改主配置

---

## ⚙️ 工作原理

### 模块加载流程

```
Zsh 启动
    ↓
读取 ~/.zshrc（主入口）
    ↓
按顺序加载 ~/.zsh/*.zsh 模块：
    1. env.zsh         # 环境变量（PATH、编辑器、语言）
    2. options.zsh     # Shell 选项（导航、补全、纠错）
    3. history.zsh     # 历史记录（配置、去重、共享）
    4. completion.zsh  # 补全系统（缓存、样式）
    5. aliases.zsh     # 别名定义（命令简写）
    ↓
加载第三方工具：
    - FZF（模糊查找）
    - zsh-autosuggestions（自动建议）
    - zsh-syntax-highlighting（语法高亮）
    - Starship（提示符）
    - Kitty 集成（终端函数）
    ↓
Zsh 就绪
```

### 设计理念

**为什么模块化？**
- ✅ 职责分离：每个模块只负责一个功能领域
- ✅ 易于维护：修改某个功能不影响其他配置
- ✅ 按需加载：可以选择性禁用某些模块
- ✅ 清晰结构：新手也能快速理解配置组织

**加载顺序重要性**
1. **env.zsh 必须最先**：其他模块可能依赖环境变量
2. **options.zsh 次之**：Shell 行为影响后续模块
3. **completion.zsh 在 options 后**：需要某些选项已设置
4. **aliases.zsh 最后**：避免干扰其他配置

---

## 📁 文件结构

```
zsh/
├── .zshrc           # 主入口文件（模块加载器）
└── config/          # 配置模块目录
    ├── env.zsh         # 环境变量配置
    ├── options.zsh     # Shell 选项配置
    ├── history.zsh     # 历史记录配置
    ├── completion.zsh  # 补全系统配置
    └── aliases.zsh     # 别名定义
```

---

## 🔧 配置说明

### 1. env.zsh - 环境变量

**编辑器设置**：
```bash
export EDITOR='vim'
export VISUAL='vim'
```

**语言和编码**：
```bash
export LANG='en_US.UTF-8'
export LC_ALL='en_US.UTF-8'
```

**PATH 扩展**：
```bash
export PATH="$HOME/dotfiles/bin:$PATH"  # 添加自定义脚本目录
```

**FZF 配置**：
- 高度：40%，反向布局，带边框
- 可自定义颜色方案匹配终端主题

---

### 2. options.zsh - Shell 选项

**目录导航**：
| 选项 | 功能 |
|------|------|
| `AUTO_CD` | 输入目录名直接跳转，无需 `cd` |
| `AUTO_PUSHD` | 自动将目录压入栈 |
| `PUSHD_IGNORE_DUPS` | 不压入重复目录 |

**模式匹配**：
- `EXTENDED_GLOB` - 启用扩展 glob（如 `^file` 匹配除 file 外的所有文件）
- `NO_NOMATCH` - 无匹配时保持模式不变，不报错

**命令纠错**：
- `CORRECT` - 自动纠正拼写错误的命令
- `NO_CORRECT_ALL` - 不纠正参数，只纠正命令

**作业控制**：
- `NOTIFY` - 后台作业状态立即报告
- `NO_HUP` - Shell 退出时不杀死后台作业

---

### 3. history.zsh - 历史记录

**基本配置**：
```bash
HISTFILE=~/.zsh_history    # 历史文件位置
HISTSIZE=10000             # 内存中保存 10,000 条命令
SAVEHIST=10000             # 文件中保存 10,000 条命令
```

**历史行为**：
| 选项 | 功能 |
|------|------|
| `EXTENDED_HISTORY` | 记录时间戳和执行时间 |
| `SHARE_HISTORY` | 所有会话共享历史记录 |
| `INC_APPEND_HISTORY` | 立即写入历史文件 |
| `HIST_IGNORE_ALL_DUPS` | 删除旧的重复条目 |
| `HIST_IGNORE_SPACE` | 空格开头的命令不记录 |
| `HIST_REDUCE_BLANKS` | 删除多余空格 |

---

### 4. completion.zsh - 补全系统

**性能优化**：
- **缓存机制**：每天只重新生成一次 compdump
- **缓存位置**：`~/.zsh/cache`

**补全样式**：
| 配置 | 效果 |
|------|------|
| 大小写不敏感 | `cd doc` 可匹配 `Documents/` |
| 菜单选择 | 使用方向键选择补全项 |
| 彩色显示 | 补全列表使用颜色（同 ls） |
| 分组显示 | 按类别分组补全结果 |

---

### 5. aliases.zsh - 别名定义

**ls 增强**：
```bash
ll      # ls -lah（详细列表）
la      # ls -A（显示隐藏文件）
l       # ls -CF（分类显示）
```

**目录导航**：
```bash
..      # cd ..
...     # cd ../..
....    # cd ../../..
-       # cd -（返回上次目录）
```

**Git 快捷键**：
```bash
gs      # git status
ga      # git add
gc      # git commit
gp      # git push
gl      # git pull
gd      # git diff
glog    # git log --oneline --graph
```

**安全操作**：
```bash
cp      # cp -i（覆盖前确认）
mv      # mv -i
rm      # rm -i
```

**macOS 专属**：
```bash
show/hide           # 显示/隐藏 Finder 中的隐藏文件
flushdns           # 刷新 DNS 缓存
hidedesktop/showdesktop  # 隐藏/显示桌面图标
```

---

## 🚀 安装

### 前置要求

```bash
# Zsh（macOS 自带）
zsh --version  # 确认已安装

# Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 必需插件
brew install zsh-autosuggestions zsh-syntax-highlighting zsh-completions

# 可选工具
brew install starship fzf  # 提示符和模糊查找
```

### 应用配置

```bash
# 1. 备份现有配置（如果存在）
[ -f ~/.zshrc ] && mv ~/.zshrc ~/.zshrc.backup
[ -d ~/.zsh ] && mv ~/.zsh ~/.zsh.backup

# 2. 创建符号链接
ln -s ~/dotfiles/zsh/.zshrc ~/.zshrc
ln -s ~/dotfiles/zsh/config ~/.zsh

# 3. 重新加载 Shell
source ~/.zshrc
```

### 验证配置

```bash
# 1. 检查符号链接
ls -la ~/.zshrc ~/.zsh

# 2. 验证 PATH
echo $PATH | grep dotfiles  # 应包含 ~/dotfiles/bin

# 3. 测试别名
ll  # 应显示详细文件列表

# 4. 测试补全
git [Tab]  # 应显示 git 子命令补全

# 5. 测试自动建议
# 输入之前用过的命令，应出现灰色建议
```

---

## ✏️ 自定义指南

### 添加新模块

创建新的配置文件：
```bash
nvim ~/.zsh/my-config.zsh
```

在 `.zshrc` 中加载：
```bash
# 在合适位置添加
source ~/.zsh/my-config.zsh
```

### 添加自定义别名

编辑 `config/aliases.zsh`：
```bash
# 添加到文件末尾
alias mycommand='echo "Hello World"'
alias dc='docker-compose'
alias k='kubectl'
```

重新加载：
```bash
source ~/.zshrc
# 或使用别名
reload
```

### 添加环境变量

编辑 `config/env.zsh`：
```bash
# 添加自定义 PATH
export PATH="$HOME/my-tools:$PATH"

# 添加其他环境变量
export MY_VAR="value"
```

### 自定义提示符（不使用 Starship）

在 `.zshrc` 中注释掉 Starship：
```bash
# eval "$(starship init zsh)"
```

添加自定义提示符（在 `config/` 创建 `prompt.zsh`）：
```bash
# 简单提示符
PROMPT='%F{blue}%~%f %# '

# 带 Git 分支的提示符
autoload -Uz vcs_info
precmd() { vcs_info }
zstyle ':vcs_info:git:*' formats '%F{yellow}(%b)%f '
setopt PROMPT_SUBST
PROMPT='%F{blue}%~%f ${vcs_info_msg_0_}%# '
```

---

## 📝 注意事项

1. **加载顺序**：不要随意调整模块加载顺序，可能导致功能失效
2. **第三方工具**：某些功能需要安装额外工具（FZF、Starship 等）
3. **Kitty 集成**：`kitty.zsh` 由 Kitty 配置提供，不在此目录
4. **性能**：补全缓存提升性能，但首次启动可能较慢
5. **兼容性**：配置针对 Zsh 5.8+，可能不兼容旧版本

---

## 🔗 参考

- [Zsh 官方文档](http://zsh.sourceforge.net/Doc/)
- [Oh My Zsh](https://ohmyz.sh/)（可参考但本配置不依赖）
- [Starship 提示符](https://starship.rs/)
- [FZF](https://github.com/junegunn/fzf)

---

[← 返回主 README](../README.md)
