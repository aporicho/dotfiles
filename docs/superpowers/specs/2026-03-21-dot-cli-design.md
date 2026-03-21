# dot — Dotfiles 管理 CLI 工具设计

## 概述

自建 Go CLI 工具，用于跨平台（macOS / Linux / WSL）管理 dotfiles 配置。核心理念：

- **模块化**：每个工具的配置是一个独立模块，按需安装/卸载
- **双向同步**：本地变更一键推送到远端，远端更新一键拉取并按需应用
- **约定驱动**：模块遵循统一约定，Claude Code 可以按约定自动编写模块

## 仓库结构

```
dotfiles/
├── dot/                        # CLI 工具源码（Go）
│   ├── main.go
│   ├── go.mod
│   └── ...
├── modules/                    # 所有配置模块
│   ├── zsh/
│   │   ├── module.toml
│   │   ├── .zshrc
│   │   └── config/
│   ├── kitty/
│   │   ├── module.toml
│   │   ├── kitty.conf
│   │   └── ...
│   ├── nvim/
│   │   ├── module.toml
│   │   ├── settings.lua
│   │   └── lua/...
│   ├── neovide/
│   │   ├── module.toml
│   │   └── config.toml
│   ├── yazi/
│   │   ├── module.toml
│   │   └── yazi.toml
│   ├── git/
│   │   ├── module.toml
│   │   └── ignore
│   └── bin/
│       ├── module.toml
│       └── nvide
├── install.sh                  # 引导脚本：安装 Go → build dot → dot pull
└── README.md
```

## 模块约定（Module Convention）

### module.toml 规范

一个模块 = `modules/` 下的一个目录 + 一个 `module.toml` 文件。

```toml
# === 基本信息（必填）===
name = "kitty"
description = "Kitty terminal emulator configuration"

# === 平台过滤（可选）===
# 省略表示全平台。可选值：darwin, linux
platforms = ["darwin", "linux"]

# === 模块依赖（可选）===
# 安装时自动先安装依赖模块。循环依赖是非法的，工具检测到时报错。
requires = ["zsh"]

# === 符号链接规则（至少一条）===

# 整个模块目录链接到目标位置
[[links]]
source = "."
target = "~/.config/kitty"

# 也可以链接单个文件
[[links]]
source = "settings.lua"
target = "~/.config/nvim/lua/plugins/dotfiles.lua"

# 平台特定链接
[[links]]
source = "kitty-darwin.conf"
target = "~/.config/kitty/platform.conf"
platforms = ["darwin"]

# === 排除文件（可选）===
# 当 source = "." 时，排除不需要链接的文件。
# 默认排除：["module.toml", "README*", "LICENSE*", ".gitignore"]
# 设为 [] 则不排除任何文件。
# 当 exclude 生效时，不做目录级符号链接，而是逐个链接目录内容。
exclude = ["module.toml", "README*"]

# === 系统依赖（可选）===
[deps.darwin]
brew = ["kitty"]

[deps.linux]
apt = ["kitty"]

# === 钩子脚本（可选）===
# 在 shell 中执行，工作目录为模块目录。
# 返回非零退出码时视为失败，中止该模块的后续步骤。
[hooks]
pre_install = ""
post_install = ""
pre_remove = ""
post_remove = ""
```

### 约定规则

1. **一个目录一个模块**，目录名即模块名
2. **`module.toml` 必须存在**，没有它的目录不被识别为模块
3. **`source = "."`** 表示链接模块目录内容（受 exclude 过滤）。具体行为：遍历模块目录下的第一层条目，排除 exclude 匹配项后，对每个剩余条目创建符号链接。子目录作为整体链接，不递归展开
4. **多个 `[[links]]`** 可共存，支持一个模块链接到多个位置
5. **`platforms`** 出现在顶层 = 模块级过滤（该平台不显示此模块）；出现在 `[[links]]` 内 = 仅该条链接的平台过滤
6. **`requires`** 安装时自动先安装依赖模块，拓扑排序，循环依赖报错
7. **钩子失败中止**：hook 返回非零退出码时，中止该模块安装/卸载
8. **目标父目录自动创建**：工具在创建符号链接前自动 `mkdir -p` 确保父目录存在
9. **冲突处理**：目标路径已存在且非符号链接时，备份为 `*.backup.{timestamp}`；已是符号链接时直接替换
10. **exclude 模式匹配**：使用 Go `filepath.Match` 语法（支持 `*`、`?`、`[abc]`），仅匹配文件名，不支持递归 `**`
11. **安装失败回滚**：hook 失败或链接创建中途出错时，回滚该模块已创建的所有符号链接，不写入 manifest
12. **反向依赖检查**：卸载模块时检查是否有其他已安装模块依赖它，如有则提示用户确认

### 各模块 module.toml 示例

**zsh**（多链接目标）：
```toml
name = "zsh"
description = "Zsh shell configuration"

[[links]]
source = ".zshrc"
target = "~/.zshrc"

[[links]]
source = "config"
target = "~/.zsh"

[deps.darwin]
brew = ["zsh-autosuggestions", "zsh-syntax-highlighting", "zsh-completions", "starship", "fzf"]

[deps.linux]
apt = ["zsh", "fzf"]

[hooks]
post_install = "echo 'Restart your shell or run: source ~/.zshrc'"
```

**nvim**（单文件注入）：
```toml
name = "nvim"
description = "Neovim configuration injection via LazyVim plugin"

[[links]]
source = "settings.lua"
target = "~/.config/nvim/lua/plugins/dotfiles.lua"

[[links]]
source = "lua/config/keymaps.lua"
target = "~/.config/nvim/lua/config/keymaps.lua"

[[links]]
source = "lua/config/options.lua"
target = "~/.config/nvim/lua/config/options.lua"

[[links]]
source = "lua/plugins/autoread.lua"
target = "~/.config/nvim/lua/plugins/autoread.lua"

[[links]]
source = "lua/plugins/tokyonight.lua"
target = "~/.config/nvim/lua/plugins/tokyonight.lua"

[deps.darwin]
brew = ["neovim"]

[deps.linux]
apt = ["neovim"]

[hooks]
pre_install = """
if [ ! -d ~/.config/nvim ]; then
  echo 'LazyVim not found. Please install LazyVim first:'
  echo '  git clone https://github.com/LazyVim/starter ~/.config/nvim'
  exit 1
fi
"""
```

**git**（单文件到嵌套目录）：
```toml
name = "git"
description = "Git global configuration"

[[links]]
source = "ignore"
target = "~/.config/git/ignore"
```

**bin**（可执行脚本）：
```toml
name = "bin"
description = "Custom executable scripts"

[[links]]
source = "."
target = "~/bin"
exclude = ["module.toml", "README*"]
```

**kitty**：
```toml
name = "kitty"
description = "Kitty terminal emulator configuration"
platforms = ["darwin", "linux"]

[[links]]
source = "."
target = "~/.config/kitty"
exclude = ["module.toml", "README*"]

[deps.darwin]
brew = ["kitty"]

[deps.linux]
apt = ["kitty"]
```

**neovide**：
```toml
name = "neovide"
description = "Neovide GUI client configuration"
platforms = ["darwin", "linux"]

[[links]]
source = "."
target = "~/.config/neovide"
exclude = ["module.toml", "README*"]

[deps.darwin]
brew = ["neovide"]
```

**yazi**：
```toml
name = "yazi"
description = "Yazi file manager configuration"
platforms = ["darwin", "linux"]

[[links]]
source = "."
target = "~/.config/yazi"
exclude = ["module.toml", "README*"]

[deps.darwin]
brew = ["yazi"]

[deps.linux]
apt = ["yazi"]
```

## CLI 命令设计

### 命令总览

```
dot status                          # 查看所有模块状态
dot push [-m "message"]             # 同步本地变更 → commit & push
dot pull                            # TUI 勾选要安装/更新的模块
dot pull <module...>                # 直接安装/更新指定模块
dot pull --all                      # 安装/更新全部模块
dot add <path> [--name <name>]      # 把本地配置纳入管理，生成模块
dot remove <module...>              # 卸载模块（移除符号链接）
dot doctor                          # 检查符号链接健康状态
```

### dot status

扫描 `modules/` 目录，检查每个模块的安装状态和变更情况。

```
$ dot status

模块状态：
  zsh        ✓ 已安装 · 无变更
  kitty      ✓ 已安装 · kitty.conf 已修改
  nvim       ✓ 已安装 · 无变更
  neovide    ✗ 未安装
  yazi       ✗ 未安装
  git        ✓ 已安装 · 无变更
  bin        ✓ 已安装 · 无变更
```

实现逻辑：
1. 遍历 `modules/` 下所有含 `module.toml` 的目录
2. 检查平台过滤（不适用当前平台的模块标注"不适用"）
3. 检查 manifest 判断是否已安装
4. 对已安装模块，通过符号链接检查配置文件是否有本地变更

### dot push

将已安装模块的本地变更同步回仓库并推送。

```
$ dot push

检测到变更：
  kitty: kitty.conf 已修改
  zsh: aliases.zsh 已修改

提交并推送？[Y/n] y
✓ 已提交: update kitty and zsh config
✓ 已推送到远端
```

实现逻辑：
1. 在 dotfiles 仓库目录执行 `git status` 检测变更
2. 展示变更摘要，按模块分组
3. 确认后 `git add modules/ → git commit → git push`（仅提交 modules/ 下的变更，非模块文件需单独处理）

### dot pull

从远端拉取更新，按需安装/更新模块。

**无参数 — TUI 模式：**
```
$ dot pull

┌ 选择要安装/更新的模块：
│
│ [x] zsh        ✓ 已安装 · 有更新
│ [x] kitty      ✓ 已安装 · 有更新
│ [ ] nvim       ✓ 已安装 · 无变更
│ [ ] neovide    ✗ 未安装
│ [ ] yazi       ✗ 未安装
│ [x] git        ✗ 未安装（新模块）
│ [ ] bin        ✓ 已安装 · 无变更
│
│ ↑↓ 移动 · 空格 勾选 · 回车 确认
└

✓ zsh: 已更新
✓ kitty: 已更新
✓ git: 已安装
```

**有参数 — 直接执行：**
```
$ dot pull zsh kitty        # 安装/更新指定模块
$ dot pull --all            # 安装/更新全部模块
```

实现逻辑：
1. `git pull` 拉取远端更新
2. 解析所有模块，过滤当前平台
3. TUI 或命令行指定要操作的模块
4. 对每个选中的模块：
   a. 检查并安装 requires 依赖
   b. 执行 pre_install hook
   c. 安装系统依赖（brew/apt）
   d. 创建符号链接（处理 exclude、冲突备份）
   e. 执行 post_install hook
   f. 记录到 manifest

### dot add \<path\> [--name \<module-name\>]

将本地配置纳入 dotfiles 管理。支持目录和单文件两种场景。

**目录场景：**
```
$ dot add ~/.config/yazi

检测到 yazi 配置：
  yazi.toml (1.2KB)
  keymap.toml (0.8KB)
  theme.toml (2.1KB)

创建模块 modules/yazi？[Y/n] y
✓ 文件已复制到 modules/yazi/
✓ module.toml 已生成
✓ 符号链接已创建
```

**单文件场景：**
```
$ dot add ~/.zshrc --name zsh

创建模块 modules/zsh？[Y/n] y
✓ 文件已复制到 modules/zsh/.zshrc
✓ module.toml 已生成
✓ 符号链接已创建
```

实现逻辑：
1. 判断路径是目录还是文件：
   - **目录**：模块名取最后一级目录名（如 `~/.config/yazi` → `yazi`）
   - **单文件**：必须通过 `--name` 指定模块名，未指定���报错提示
2. 如果 `modules/<name>/` 已存在，报错退出（提示：`模块 <name> 已存在，如需更新请直接编辑`）
3. 复制配置文件到 `modules/<name>/`
4. 生成 `module.toml`（推断 target 为原路径，目录场景使用 `source = "."` + 默认 exclude）
5. 用符号链接替换原配置（备份原文件）
6. 记录到 manifest

### dot remove \<module...\>

卸载模块，移除符号链接。

```
$ dot remove kitty

确认卸载 kitty？[Y/n] y
✓ 执行 pre_remove hook
✓ 移除符号链接: ~/.config/kitty/*
✓ 恢复备份: ~/.config/kitty.backup.20260321
✓ 执行 post_remove hook
✓ kitty 已卸载
```

实现逻辑：
1. 读取 manifest 获取该模块创建的所有符号链接
2. 执行 pre_remove hook
3. 逐个删除符号链接
4. 如果存在备份，恢复备份
5. 执行 post_remove hook
6. 从 manifest 中移除记录

## 安装清单（Manifest）

工具在 `~/.local/share/dot/manifest.json` 维护安装状态。source 路径相对于 `dotfiles_path` 解析。

```json
{
  "version": 1,
  "dotfiles_path": "/Users/aporicho/dotfiles",
  "modules": {
    "zsh": {
      "installed_at": "2026-03-21T10:00:00Z",
      "links": [
        { "source": "modules/zsh/.zshrc", "target": "/Users/aporicho/.zshrc" },
        { "source": "modules/zsh/config", "target": "/Users/aporicho/.zsh" }
      ]
    },
    "kitty": {
      "installed_at": "2026-03-21T10:00:00Z",
      "links": [
        { "source": "modules/kitty/kitty.conf", "target": "/Users/aporicho/.config/kitty/kitty.conf" },
        { "source": "modules/kitty/themes", "target": "/Users/aporicho/.config/kitty/themes" }
      ]
    }
  }
}
```

manifest 的作用：
- `dot status` 判断模块是否已安装
- `dot remove` 知道要删除哪些符号链接
- 记录 dotfiles 仓库路径，避免硬编码

## 技术选型

- **语言**：Go
- **TUI 框架**：[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) + [bubbles](https://github.com/charmbracelet/bubbles)（多选列表组件）
- **TOML 解析**：[BurntSushi/toml](https://github.com/BurntSushi/toml)
- **CLI 框架**：[spf13/cobra](https://github.com/spf13/cobra)
- **构建产物**：单二进制文件，无外部依赖

## 引导安装（install.sh）

首次安装的引导脚本，负责：
1. 克隆 dotfiles 仓库到 `~/dotfiles`
2. 安装 Go（如果没有）
3. 在仓库内 `cd dot && go build -o ~/bin/dot`
4. 执行 `dot pull` 进入 TUI 选择模块

## dot push 的 commit message

自动生成格式：`update <module1>, <module2> config`。支持 `-m` 参数自定义：

```
$ dot push -m "fix kitty font size"
```

## dot doctor

检查所有已安装模块的健康状态。

```
$ dot doctor

检查模块健康状态...
  zsh        ✓ 正常
  kitty      ✗ 符号链接断裂: ~/.config/kitty/kitty.conf
  nvim       ✓ 正常
  git        ✗ 符号链接被替换为普通文件: ~/.config/git/ignore

发现 2 个问题。运行 dot pull kitty git 修复。
```

实现逻辑：
1. 遍历 manifest 中所有已安装模块
2. 检查每个符号链接：是否存在、是否指向正确的源文件
3. 输出损坏的链接并提供修复建议

## 迁移计划

从当前仓库结构迁移到新结构的步骤：

### 前置条件

当前存在两个仓库：`~/dotfiles`（实际使用）和 `~/Desktop/dotfiles`（推到 GitHub）。以 `~/dotfiles` 为准（它有最新的本地配置变更），合并 `~/Desktop/dotfiles` 独有的内容（install.sh）。

### 迁移步骤

1. **合并两个仓库**
   - 将 `~/Desktop/dotfiles` 独有的 `install.sh` 合入 `~/dotfiles`
   - 提交 `~/dotfiles` 的所有未提交变更
   - 以 `~/dotfiles` 为唯一仓库，废弃 `~/Desktop/dotfiles`

2. **重组目录结构**
   - `git mv zsh modules/zsh`
   - `git mv kitty modules/kitty`
   - `git mv nvim modules/nvim`
   - `git mv neovide modules/neovide`
   - `git mv bin modules/bin`
   - 新建 `modules/yazi/`、`modules/git/` 并从本地复制配置
   - 为每个模块创建 `module.toml`

3. **构建 dot CLI**
   - 创建 `dot/` 目录，初始化 Go 项目
   - 实现核心命令
   - `go build -o ~/bin/dot`

4. **更新符号链接**
   - 运行 `dot pull --all` 重建所有符号链接（指向新的 `modules/` 路径）

5. **更新 install.sh**
   - 改写为引导脚本（克隆仓库 → build dot → dot pull）

6. **验证并推送**
   - `dot doctor` 检查所有链接正常
   - `dot push` 推送到远端

## 后续迭代（v2+）

以下能力不在 v1 范围内，预留设计空间：

- **模板/变量替换**：`source = "kitty.conf.tmpl"` + 数据文件，按机器生成不同配置
- **条件扩展**：`when = { hostname = "work-*" }`，超越平台的条件匹配
- **glob source**：`source = "lua/**/*.lua"`，解决 nvim 多文件逐个列举的繁琐
- **dry-run 预览**：`dot pull --dry-run` 展示将要执行的操作但不实际执行
- **文件权限标记**：`mode = "755"`，显式控制权限
- **加密文件**：敏感配置（SSH key 等）的加密存储
- **dot diff**：`dot pull` 前预览远端变更内容，安全保障
- **dot clean**：清理累积的 `*.backup.*` 备份文件
