# Dotfiles

跨平台（macOS / Linux / WSL）开发环境配置，使用自建 `dot` CLI 工具管理，模块化按需加载。

## 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/aporicho/dotfiles/main/install.sh | bash
```

脚本会自动克隆仓库、构建 `dot` CLI、进入 TUI 选择要安装的模块。

## dot CLI

```bash
dot status                      # 查看所有模块状态
dot pull                        # TUI 勾选安装/更新模块
dot pull zsh kitty              # 直接安装指定模块
dot pull --all                  # 安装全部模块
dot push                        # 提交并推送本地变更
dot push -m "fix font size"     # 自定义提交信息
dot add ~/.config/starship.toml --name starship  # 纳入新配置
dot remove kitty                # 卸载模块
dot doctor                      # 检查符号链接健康
dot clean                       # 清理备份文件
```

## 模块

| 模块 | 说明 | 平台 |
|------|------|------|
| zsh | Zsh 配置（别名、补全、历史、环境变量） | 全平台 |
| kitty | Kitty 终端（Tokyo Night 主题、零延迟） | macOS / Linux |
| nvim | Neovim 插件注入（基于 LazyVim） | 全平台 |
| neovide | Neovide GUI 配置 | macOS / Linux |
| yazi | Yazi 文件管理器 | macOS / Linux |
| git | Git 全局 ignore | 全平台 |
| bin | 自定义脚本（nvide 启动器等） | 全平台 |

## 仓库结构

```
dotfiles/
├── dot/                    # dot CLI 工具源码（Go）
├── modules/                # 配置模块
│   ├── zsh/                #   .zshrc + config/
│   ├── kitty/              #   kitty.conf + themes/
│   ├── nvim/               #   settings.lua + lua/
│   ├── neovide/            #   config.toml
│   ├── yazi/               #   yazi.toml
│   ├── git/                #   ignore
│   └── bin/                #   nvide
├── install.sh              # 引导安装脚本
└── README.md
```

每个模块目录包含一个 `module.toml` 描述文件，定义符号链接规则、平台过滤、系统依赖和钩子脚本。

## 添加新模块

### 方式一：dot add

```bash
dot add ~/.config/starship.toml --name starship
```

自动复制文件、生成 `module.toml`、创建符号链接。

### 方式二：手动创建

在 `modules/` 下新建目录，编写 `module.toml`：

```toml
name = "starship"
description = "Starship prompt configuration"

[[links]]
source = "starship.toml"
target = "~/.config/starship.toml"

[deps.darwin]
brew = ["starship"]
```

然后 `dot pull starship` 安装。

## 日常使用

```bash
# 改了配置后，推送到远端
dot push

# 在另一台机器上拉取更新
dot pull

# 新机器只装需要的模块
dot pull zsh nvim git bin

# 检查配置是否正常
dot doctor
```

## 技术栈

- **dot CLI**: Go + cobra + bubbletea
- **主题**: Tokyo Night（统一 Neovim / Kitty）
- **字体**: JetBrainsMono Nerd Font
