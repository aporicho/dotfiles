# Kitty 终端配置

这是我的 Kitty 终端自定义配置，采用模块化设计，与 Neovim/Zsh 配置风格一致。

**配置特点**：
- 🎨 **Tokyo Night 主题** - 与 Neovim 统一视觉体验
- ⚡ **性能优化** - 0ms 输入延迟、2ms 重绘延迟
- ⌨️ **macOS 风格快捷键** - Command 键操作，符合系统习惯
- 🔄 **Shell 集成** - 命令跳转、输出复制、工作目录跟踪
- 📁 **会话管理** - 预设开发/工作环境布局
- 🔧 **模块化配置** - 功能分离，易于维护

---

## ⚙️ 工作原理

### 模块化加载机制

```
Kitty 启动
    ↓
读取 ~/.config/kitty/kitty.conf（主配置）
    ↓
kitty.conf 使用 include 指令按顺序加载：
    1. themes/style.conf        # 通用样式（字体、窗口）
       ├→ tokyo-night-colors.conf  # Tokyo Night 配色
    2. performance.conf         # 性能优化
    3. layouts.conf            # 布局管理
    4. keymaps.conf            # 快捷键映射
    5. shell_integration.conf   # Shell 集成
    6. modern_features.conf    # 现代功能
    ↓
所有模块配置合并生效
```

### 设计理念

**为什么模块化？**
- ✅ 功能分离：每个配置文件职责单一，易于查找和修改
- ✅ 可维护性：修改某个功能不影响其他配置
- ✅ 可复用性：可以单独使用某些模块（如主题）

**配置加载顺序**
1. 样式优先：字体、窗口、主题先加载，确保视觉正确
2. 性能其次：性能优化早加载，提升整体体验
3. 功能最后：布局、快捷键、集成功能后加载

---

## 📁 目录结构

```
kitty/
├── kitty.conf                    # 主配置文件（模块加载器）
├── performance.conf              # 性能优化配置
├── layouts.conf                  # 窗口布局管理
├── keymaps.conf                  # 快捷键配置
├── shell_integration.conf        # Shell 集成
├── modern_features.conf          # 现代功能（通知、URL 检测）
├── kitty.zsh                     # Zsh 集成脚本
├── verify.sh                     # 配置验证脚本
├── themes/
│   ├── tokyo-night-colors.conf   # Tokyo Night 配色方案
│   └── style.conf                # 通用样式（字体、窗口、标签栏）
└── sessions/
    ├── dev.session               # 开发环境会话
    └── work.session              # 工作环境会话
```

## ✨ 主要功能

### 🎨 主题与字体
- **Tokyo Night (night 变体)** - 与 Neovim 主题一致
  - 背景色：`#1a1b26`（深蓝黑）
  - 前景色：`#c0caf5`（浅蓝白）
- **JetBrainsMono Nerd Font Mono (18pt)** - 与 Neovim/Neovide 统一
- **编程连字** - 美化 `=>`, `!=`, `>=` 等符号
- **Display P3 颜色空间** - macOS 优化，颜色更鲜艳

### ⌨️ 快捷键配置

#### 窗口管理
- `Cmd+Enter` - 新建窗口（继承当前目录）
- `Cmd+T` - 新建标签（继承当前目录）
- `Cmd+N` - 新建 OS 窗口
- `Cmd+W` - 关闭窗口
- `Cmd+Shift+W` - 关闭标签

#### 窗口分割
- `Cmd+D` - 垂直分割（左右）
- `Cmd+Shift+D` - 水平分割（上下）
- `Cmd+Z` - 最大化/还原当前窗口（切换 stack 布局）

#### 窗口导航
- `Cmd+Left/Right/Up/Down` - 窗口间导航
- `Cmd+1~9` - 切换到指定标签

#### 窗口调整
- `Cmd+Ctrl+Left/Right` - 调整窗口宽度
- `Cmd+Ctrl+Up/Down` - 调整窗口高度

#### 布局管理
- `Cmd+L` - 切换布局（splits/stack/tall/fat/grid/horizontal/vertical）
- `Cmd+Shift+L` - 返回上次使用的布局

#### 文本操作
- `Cmd+K` - 清屏到光标位置（需 shell 集成）
- `Cmd+Shift+K` - 清除整个终端
- `Cmd+Ctrl+K` - 清除滚动历史
- `Cmd+H` - 在编辑器中查看滚动历史

#### 滚动控制
- `Cmd+Shift+Up/Down` - 向上/下滚动一行
- `Cmd+Shift+Page Up/Down` - 向上/下滚动一页
- `Cmd+Shift+Home/End` - 滚动到顶部/底部
- `Cmd+Shift+Z/X` - 跳转到上一个/下一个提示符（需 shell 集成）

#### 其他
- `Cmd+F` - 搜索滚动历史
- `Cmd+Shift+R` - 重新加载配置
- `Cmd+=/−` - 调整字体大小
- `Cmd+0` - 重置字体大小

### ⚡ 性能优化

**极致性能配置** (`performance.conf`)：
- **输入延迟**：0ms（默认 3ms）- 最低延迟
- **重绘延迟**：2ms（默认 10ms）- 快速响应
- **显示器同步**：启用 - 防止画面撕裂
- **滚动缓冲区**：10,000 行

**渲染优化**：
- GPU 缓存字形（自动启用）
- macOS 字体加粗：0.1（提升可读性）

### 🔄 Shell 集成

启用 Kitty Shell 集成后可使用：
- **命令提示符跳转**：快速定位到上一条/下一条命令
- **输出复制**：复制上一条命令的输出
- **智能光标**：提示符处自动切换为 beam 光标
- **工作目录跟踪**：新窗口/标签自动继承当前目录

### 🐚 Zsh 集成函数

在 Zsh 中可使用以下函数（来自 `kitty.zsh`）：

```bash
# 列出可用会话
ksession

# 加载指定会话
ksession dev     # 加载开发环境会话
ksession work    # 加载工作环境会话

# 重新加载 Kitty 配置
kreload

# 在 Kitty 窗口中打开文件
kopen file.txt

# 在当前目录打开新 Kitty 窗口
kcd

# 在新标签打开
kt

# SSH 增强（使用 Kitty kitten）
kssh user@host
```

**别名**：
- `kcd` - 在当前目录打开新 OS 窗口
- `kt` - 在当前目录打开新标签
- `kssh` - 使用 Kitty SSH kitten（支持图像传输）

### 📁 会话管理

会话文件保存在 `sessions/` 目录：

#### **dev.session** - 开发环境
- 标签 1：主开发（3 窗口分割）
- 标签 2：服务器
- 标签 3：系统监控（htop）

#### **work.session** - 工作环境
- 标签 1：终端（Documents）
- 标签 2：日志（/var/log）
- 标签 3：笔记（Notes）

**使用方法**：
```bash
# 方法 1：启动时加载会话
kitty --session ~/.config/kitty/sessions/dev.session

# 方法 2：在 Zsh 中切换会话
ksession dev
```

**自定义会话**：
在 `sessions/` 目录创建 `.session` 文件，使用以下语法：
```
# 设置布局
layout splits

# 创建标签
new_tab 标签名
cd ~/path/to/dir
launch zsh

# 分割窗口
launch --location=vsplit zsh
launch --location=hsplit htop
```

## 🚀 安装

### 前置要求
- Kitty 终端（>= 0.36）
- Zsh Shell
- JetBrainsMono Nerd Font

```bash
# macOS 安装 Kitty
brew install --cask kitty

# 安装字体
brew tap homebrew/cask-fonts
brew install font-jetbrains-mono-nerd-font
```

### 应用配置

```bash
# 使用符号链接（推荐，已完成）
ln -s ~/dotfiles/kitty ~/.config/kitty

# 重新加载配置
# 在 Kitty 中按 Cmd+Shift+R
```

### Zsh 集成

确保 `.zshrc` 中包含：
```bash
# Kitty 终端集成
if [[ -f ~/.config/kitty/kitty.zsh ]]; then
    source ~/.config/kitty/kitty.zsh
fi
```

然后重新加载 Shell：
```bash
source ~/.zshrc
# 或使用别名
reload
```

### 验证配置

运行验证脚本：
```bash
~/dotfiles/kitty/verify.sh
```

验证内容：
- ✅ Kitty 版本信息
- ✅ 配置文件完整性
- ✅ 会话文件检查
- ✅ 配置语法检查
- ✅ Zsh 集成检查

## 🎨 主题定制

### 更换配色方案

编辑 `themes/style.conf`，修改第 9 行：
```bash
# 当前使用 Tokyo Night
include tokyo-night-colors.conf

# 更换为其他主题（需要先创建配色文件）
# include catppuccin.conf
# include dracula.conf
```

### 自定义颜色

编辑 `themes/tokyo-night-colors.conf`，修改颜色值：
```bash
background #1a1b26      # 背景色
foreground #c0caf5      # 前景色
cursor #c0caf5          # 光标颜色
# ... 其他颜色
```

### 调整字体

编辑 `themes/style.conf`：
```bash
font_family JetBrainsMono Nerd Font Mono
font_size 18.0          # 调整字体大小
```

## 🔧 配置说明

### 窗口样式

- **窗口装饰**：`titlebar-only`（隐藏标题栏，保留圆角）
- **窗口内边距**：8px
- **边框宽度**：1pt
- **标签栏位置**：顶部
- **标签栏样式**：Powerline 圆角

### 光标样式

- **形状**：beam（竖线）
- **粗细**：1.5
- **闪烁间隔**：0.5s
- **停止闪烁**：15s 无操作后

### 鼠标与剪贴板

- **自动隐藏鼠标**：2 秒无操作
- **选中即复制**：启用
- **智能删除空格**：启用

### macOS 专属

- **Option 键**：作为 Alt 键
- **颜色空间**：Display P3（更鲜艳）
- **字体加粗**：0.1（提升可读性）
- **退出行为**：关闭最后窗口时退出应用

## 🐛 故障排除

### 配置不生效

```bash
# 检查配置语法
kitty --config ~/.config/kitty/kitty.conf --debug-config

# 重新加载配置
# 在 Kitty 中按 Cmd+Shift+R
```

### 快捷键不工作

1. 检查是否与系统快捷键冲突
2. 查看 `keymaps.conf` 中的映射
3. 使用 `Cmd+,` 打开配置文件调试

### 字体显示异常

```bash
# 检查字体是否安装
kitty + list-fonts | grep JetBrains

# 如未安装
brew install font-jetbrains-mono-nerd-font
```

### Shell 集成不工作

```bash
# 检查 shell_integration 是否启用
grep "shell_integration" ~/.config/kitty/shell_integration.conf

# 应该显示：
# shell_integration enabled
```

### Zsh 函数不可用

```bash
# 检查 kitty.zsh 是否被加载
grep "kitty.zsh" ~/.zshrc

# 检查是否在 Kitty 终端中
echo $TERM  # 应该输出: xterm-kitty
```

## 📚 参考资源

- [Kitty 官方文档](https://sw.kovidgoyal.net/kitty/)
- [Tokyo Night 主题](https://github.com/folke/tokyonight.nvim)
- [Kitty Shell 集成](https://sw.kovidgoyal.net/kitty/shell-integration/)
- [Kitty Kittens](https://sw.kovidgoyal.net/kitty/kittens/)

## 🔄 更新日志

### 2025-10-16 - 集成到 dotfiles
- ✅ 迁移到模块化 dotfiles 仓库
- ✅ 创建符号链接管理
- ✅ 添加完整文档
- ✅ 集成 Zsh 增强函数

### 2025-10-15 - 初始优化
- ✅ 实现模块化配置结构
- ✅ Tokyo Night 主题集成
- ✅ 性能优化（0ms 输入延迟）
- ✅ macOS 快捷键优化
- ✅ Shell 集成启用
- ✅ 会话管理功能

---

享受你的 Kitty 终端配置！

---

[← 返回主 README](../README.md)
