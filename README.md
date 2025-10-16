# Dotfiles

现代化的 macOS 开发环境配置，采用模块化设计和符号链接管理，统一 Tokyo Night 主题。

**核心特点**：
- 🔗 **符号链接管理** - 配置文件集中管理，修改自动同步
- 🧩 **模块化设计** - 各配置独立维护，按需加载
- 🎨 **统一主题** - Tokyo Night 配色贯穿所有工具
- 🚀 **最小侵入** - 不覆盖原始配置，完全可逆

---

## ⚙️ 工作原理

### 符号链接管理
所有配置通过符号链接连接到系统配置目录，优点：
- 配置文件实际存储在 `~/dotfiles`，通过 Git 版本控制
- 修改配置文件立即生效，自动同步到仓库
- 完全可逆，删除符号链接即可恢复

### 配置加载机制

**Neovim (LazyVim)**：
```
~/.config/nvim/lua/plugins/dotfiles.lua  →  ~/dotfiles/nvim/settings.lua
                                              ↓
                            自动加载 lua/config/*.lua（配置覆盖）
                            自动加载 lua/plugins/*.lua（插件规范）
```
- 单一符号链接，LazyVim 启动时自动发现并加载
- `settings.lua` 动态扫描 config/ 和 plugins/ 目录
- 与 LazyVim 完全隔离，互不干扰

**Zsh**：
```
~/.zshrc  →  ~/dotfiles/zsh/.zshrc
              ↓
      依次加载 ~/.zsh/*.zsh 模块
      (env → options → history → completion → aliases)
```
- 模块化加载，每个功能独立文件
- 按顺序加载确保依赖关系正确

**Kitty**：
```
~/.config/kitty  →  ~/dotfiles/kitty/
                      ↓
            kitty.conf 加载各模块配置
            (performance/layouts/keymaps/themes...)
```
- 主配置作为加载器，include 各功能模块

---

## 📦 前置要求

### 系统环境
- **macOS** (测试于 macOS 14.x+)
- **Homebrew** - [安装方法](https://brew.sh/)

### 必需工具
```bash
# 终端和编辑器
brew install --cask kitty neovide
brew install neovim

# 字体（所有配置统一使用）
brew install --cask font-jetbrains-mono-nerd-font

# Shell 增强（Zsh 自带，只需安装插件）
brew install zsh-autosuggestions zsh-syntax-highlighting zsh-completions
```

### 可选工具
```bash
# 提示符和工具
brew install starship fzf

# Kitty 增强
brew install imagemagick  # 图片预览支持
```

---

## 🚀 快速开始

### 1. 克隆仓库
```bash
git clone https://github.com/your-username/dotfiles.git ~/dotfiles
cd ~/dotfiles
```

### 2. 创建符号链接

**Neovim**（需先安装 LazyVim starter）：
```bash
# 安装 LazyVim starter（如果还没有）
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git

# 创建符号链接
ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua
```

**其他配置**：
```bash
# Kitty
ln -s ~/dotfiles/kitty ~/.config/kitty

# Zsh
ln -s ~/dotfiles/zsh/.zshrc ~/.zshrc
ln -s ~/dotfiles/zsh/config ~/.zsh

# Neovide
ln -s ~/dotfiles/neovide ~/.config/neovide
```

### 3. 重新加载配置
```bash
# 重新加载 Zsh
source ~/.zshrc

# 启动 Neovim（自动安装插件）
nvim

# 重新加载 Kitty（在 Kitty 中按 Cmd+Shift+R）
```

---

## 📁 目录结构

```
dotfiles/
├── nvim/              # Neovim 配置（基于 LazyVim）
│   ├── settings.lua   # 统一加载器（符号链接入口）
│   └── lua/
│       ├── config/    # 配置覆盖（keymaps/options）
│       └── plugins/   # 自定义插件（主题/功能扩展）
│
├── kitty/             # Kitty 终端配置
│   ├── kitty.conf     # 主配置（模块加载器）
│   ├── *.conf         # 功能模块（性能/布局/快捷键）
│   ├── themes/        # Tokyo Night 主题配色
│   └── sessions/      # 预设会话（dev/work）
│
├── zsh/               # Zsh Shell 配置
│   ├── .zshrc         # 主入口文件
│   └── config/        # 配置模块（env/aliases/completion）
│
├── neovide/           # Neovide GUI 配置
│   └── config.toml    # 窗口和性能设置
│
└── bin/               # 自定义脚本
    └── nvide          # Neovide 启动器
```

**详细说明**：
- [nvim/README.md](nvim/README.md) - Neovim 配置详解
- [kitty/README.md](kitty/README.md) - Kitty 终端配置详解

---

## 🔧 详细安装

### Neovim

**重要**：本配置基于 [LazyVim](https://www.lazyvim.org/)，需先安装 LazyVim starter。

```bash
# 1. 安装 LazyVim starter
git clone https://github.com/LazyVim/starter ~/.config/nvim
rm -rf ~/.config/nvim/.git

# 2. 创建符号链接（单一入口点）
ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua

# 3. 启动 Neovim，自动安装所有插件
nvim
```

**配置内容**：
- **主题**：Tokyo Night (night 变体)
- **快捷键**：`Cmd+E`（文件浏览器）、`Cmd+/`（浮动终端）、`Cmd+T`（右侧终端）
- **Neovide**：字体、光标特效、性能优化
- **自动刷新**：激进模式文件监测（200ms）

详见 [nvim/README.md](nvim/README.md)

---

### Kitty

```bash
# 1. 安装 Kitty（如未安装）
brew install --cask kitty

# 2. 创建符号链接
ln -s ~/dotfiles/kitty ~/.config/kitty

# 3. 验证配置（可选）
~/dotfiles/kitty/verify.sh

# 4. 重新加载 Kitty
# 在 Kitty 中按 Cmd+Shift+R
```

**配置内容**：
- **主题**：Tokyo Night - 与 Neovim 统一
- **性能**：0ms 输入延迟、2ms 重绘延迟
- **快捷键**：macOS 风格 Command 键操作
- **Shell 集成**：命令跳转、输出复制、目录跟踪
- **会话管理**：开发/工作环境预设

详见 [kitty/README.md](kitty/README.md)

---

### Zsh

```bash
# 1. 创建符号链接
ln -s ~/dotfiles/zsh/.zshrc ~/.zshrc
ln -s ~/dotfiles/zsh/config ~/.zsh

# 2. 安装依赖（如未安装）
brew install zsh-autosuggestions zsh-syntax-highlighting zsh-completions starship

# 3. 重新加载 Shell
source ~/.zshrc
# 或使用别名
reload
```

**配置内容**：
- **模块化**：环境变量、选项、历史、补全、别名分离
- **增强**：自动建议、语法高亮、智能补全
- **提示符**：Starship（可选）
- **PATH**：自动添加 `~/dotfiles/bin`

---

### Neovide

```bash
# 1. 安装 Neovide（如未安装）
brew install --cask neovide

# 2. 创建符号链接
ln -s ~/dotfiles/neovide ~/.config/neovide

# 3. 使用 nvide 命令启动
nvide                    # 打开当前目录
nvide /path/to/file      # 打开指定文件
```

**配置内容**：
- **窗口**：无边框模式、启动时最大化
- **集成**：自动读取 Neovim 配置

---

## ✅ 验证配置

### 快速验证所有配置

```bash
# 1. 验证符号链接
ls -la ~/.config/nvim/lua/plugins/dotfiles.lua  # 应指向 ~/dotfiles/nvim/settings.lua
ls -la ~/.config/kitty                           # 应指向 ~/dotfiles/kitty
ls -la ~/.zshrc                                  # 应指向 ~/dotfiles/zsh/.zshrc

# 2. 验证 Zsh 加载
echo $PATH | grep dotfiles  # 应包含 ~/dotfiles/bin

# 3. 验证 Kitty 配置
kitty --debug-config  # 检查配置语法

# 4. 验证 Neovim 插件
nvim -c ":Lazy"  # 查看插件加载状态
```

### 常见问题检查

**Neovim 插件未加载**：
```bash
# 检查符号链接是否正确
ls -la ~/.config/nvim/lua/plugins/dotfiles.lua

# 查看 LazyVim 日志
nvim -c ":Lazy log"
```

**Kitty 快捷键不工作**：
- 检查是否与系统快捷键冲突
- 系统偏好设置 → 键盘 → 快捷键

**Zsh 模块未加载**：
```bash
# 检查 ~/.zsh 目录
ls -la ~/.zsh

# 重新加载
source ~/.zshrc
```

---

## 🔄 备份说明

### 自动备份
首次创建符号链接时，建议备份原始配置：

```bash
# 备份原始配置（如果存在）
mv ~/.config/nvim ~/.config/nvim.backup
mv ~/.config/kitty ~/.config/kitty.backup
mv ~/.config/neovide ~/.config/neovide.backup
mv ~/.zshrc ~/.zshrc.backup
mv ~/.zsh ~/.zsh.backup
```

### 恢复原始配置
```bash
# 删除符号链接
rm ~/.config/nvim/lua/plugins/dotfiles.lua
rm ~/.config/kitty
rm ~/.config/neovide
rm ~/.zshrc
rm ~/.zsh

# 恢复备份
mv ~/.config/nvim.backup ~/.config/nvim
mv ~/.config/kitty.backup ~/.config/kitty
mv ~/.config/neovide.backup ~/.config/neovide
mv ~/.zshrc.backup ~/.zshrc
mv ~/.zsh.backup ~/.zsh
```

### Git 管理
所有配置通过符号链接管理，修改会自动同步到仓库：

```bash
cd ~/dotfiles

# 查看更改
git status
git diff

# 提交更改
git add .
git commit -m "更新配置"
git push
```

---

## 🎨 主题说明

所有配置统一使用 **Tokyo Night (night 变体)** 主题：

- **背景色**：`#1a1b26`（深蓝黑）
- **前景色**：`#c0caf5`（浅蓝白）
- **强调色**：蓝色、青色、紫色系

配置位置：
- Neovim: `nvim/lua/plugins/tokyonight.lua`
- Kitty: `kitty/themes/tokyo-night-colors.conf`

---

## 📝 注意事项

1. **LazyVim 依赖**：Neovim 配置必须基于 LazyVim，否则无法正常工作
2. **符号链接管理**：修改配置文件会直接影响 Git 仓库，建议频繁提交
3. **PATH 配置**：`~/dotfiles/bin` 已在 Zsh 配置中添加到 PATH
4. **字体要求**：所有配置使用 JetBrainsMono Nerd Font，必须安装
5. **版本兼容**：测试于 macOS 14.x、Neovim 0.10+、Kitty 0.36+

---

## 🔗 参考资源

- [LazyVim 官方文档](https://www.lazyvim.org/)
- [Kitty 终端文档](https://sw.kovidgoyal.net/kitty/)
- [Tokyo Night 主题](https://github.com/folke/tokyonight.nvim)
- [Starship 提示符](https://starship.rs/)

---

**最后更新**：2025-10-17
