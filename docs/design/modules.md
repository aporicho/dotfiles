# 模块系统

dot CLI 的核心概念。每个模块管理一组配置文件的符号链接。

---

## 模块结构

```
modules/
├── zsh/
│   ├── module.toml      # 模块定义
│   ├── .zshrc           # 配置文件
│   ├── config/          # 配置目录
│   └── secrets.env      # 加密文件（gitignore）
├── kitty/
│   ├── module.toml
│   └── kitty.conf
└── ...
```

---

## module.toml

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模块名 |
| `description` | string | 否 | 简短描述 |
| `platforms` | []string | 否 | 平台过滤，空=全平台 |
| `requires` | []string | 否 | 依赖的其他模块 |
| `exclude` | []string | 否 | `source = "."` 时排除的文件 |
| `[[links]]` | 数组 | 是 | 符号链接规则 |
| `[[secrets]]` | 数组 | 否 | 加密文件规则（详见 `secrets.md`）|
| `[deps.darwin]` | 对象 | 否 | macOS 依赖 |
| `[deps.linux]` | 对象 | 否 | Linux 依赖 |
| `[hooks]` | 对象 | 否 | 生命周期钩子 |

---

## 链接规则

```toml
[[links]]
source = ".zshrc"        # 相对模块目录
target = "~/.zshrc"      # 目标路径（~ 展开）
platforms = ["darwin"]    # 可选平台过滤
```

- `source = "."` 展开为目录内所有文件（排除 `exclude` 匹配的）
- 目标已存��时备份为 `{target}.backup.{timestamp}`
- 目标已是符号链接时直接替换（无备份）

---

## 系统依赖

```toml
[deps.darwin]
brew = ["starship", "fzf"]

[deps.linux]
apt = ["zsh", "fzf"]
```

---

## 钩子

```toml
[hooks]
pre_install = "echo before"
post_install = "echo after"
pre_remove = "echo before"
post_remove = "echo after"
```

- 在模块目录下执行
- stdout/stderr 直接转发

---

## 已有模块

| 模块 | 说明 | 平台 |
|------|------|------|
| zsh | Zsh 配置 + secrets | 全平台 |
| kitty | Kitty 终端 | macOS / Linux |
| nvim | Neovim 插件注入 | 全平台 |
| neovide | Neovide GUI | macOS / Linux |
| yazi | Yazi 文件管理器 | macOS / Linux |
| git | Git 全局 ignore | 全平台 |
| bin | 自定义脚本 | 全平台 |
