# dot CLI

跨平台 dotfiles 模块管理工具。

---

## 命令

| 命令 | 说明 |
|------|------|
| `dot` | 无子命令时进入 TUI 仪表盘（详见 `tui.md`） |
| `dot status` | 显示所有模块安装状态 |
| `dot pull` | 从远端拉取代码更新（仅 git pull） |
| `dot install <name>` | 安装指定模块（创建符号链接+依赖+secrets） |
| `dot install --all` | 安装全部模块 |
| `dot install` | 无参数时 TUI 选择器 |
| `dot uninstall <name>` | 卸载模块（删除符号链接，模块文件保留） |
| `dot remove <name>` | 彻底删除模块（从 modules/ 目录移除文件） |
| `dot push` | 提交并推送变更 |
| `dot push -m "msg"` | 自定义提交信息 |
| `dot add <path>` | 纳入新配置文件为模块（复制+链接） |
| `dot add <path> --name <name>` | 指定模块名 |
| `dot doctor` | 检查符号链接和 secrets 健康 |
| `dot clean` | 清理：备份文件 + 孤立链接 + 无效 manifest 记录 |

---

## 命令职责划分

### pull vs install

| | pull | install |
|--|------|---------|
| 做什么 | `git pull` 拉取远端代码 | 创建符号链接、安装依赖、解密 secrets |
| 操作对象 | 整个仓库 | 指定模块或全部模块 |
| 关系 | 拉取了不一定安装 | 安装不依赖 pull |

### uninstall vs remove

| | uninstall | remove |
|--|-----------|--------|
| 做什么 | 删除符号链接、恢复备份 | 从 modules/ 目录删除模块文件 |
| 模块文件 | 保留在 modules/ | 删除 |
| 可逆性 | 可以重新 install | 不可逆（需重新 add） |
| 前提 | — | 必须先 uninstall |

### add 机制

- 复制文件到 `modules/<name>/`（非移动）
- 自动生成 `module.toml`
- 创建符号链接（原位置 → modules/ 内的文件）
- 支持 `--generate` 参数让 AI 生成 module.toml 草稿 **（未实现）**

### clean 范围

| 清理项 | 说明 |
|--------|------|
| 备份文件 | `*.backup.*` 文件 |
| 孤立链接 | 指向不存在目标的符号链接 |
| 无效 manifest | manifest 中已删除模块的记录 |

---

## 数据存储

| 数据 | 位置 |
|------|------|
| 模块文件 | `~/dotfiles/modules/` |
| 安装记录 | `~/.local/share/dot/manifest.json` |
| dot CLI 源码 | `~/dotfiles/dot/` |

---

## Manifest

JSON 格式，记录已安装模块及其链接。

```json
{
  "version": 1,
  "dotfiles_path": "/Users/xxx/dotfiles",
  "modules": {
    "zsh": {
      "installed_at": "2026-03-21T...",
      "links": [
        {"source": "/path/to/.zshrc", "target": "~/.zshrc", "backup": ""}
      ]
    }
  }
}
```

---

## 技术栈

| 组件 | 技术 |
|------|------|
| CLI 框架 | cobra |
| TUI 框架 | bubbletea + lipgloss |
| 配置解析 | BurntSushi/toml |
| 加密 | filippo.io/age |
| 语言 | Go 1.26 |
