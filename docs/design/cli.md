# dot CLI

跨平台 dotfiles 模块管理工具。

---

## 命令

| 命令 | 说明 |
|------|------|
| `dot` | 无子命令时进入 TUI 仪表盘（详见 `tui.md`） |
| `dot status` | 显示所有模块安装状态 |
| `dot pull` | 安装/更新模块（TUI 选择或指定名称） |
| `dot pull --all` | 安装全部模块 |
| `dot push` | 提交并推送变更 |
| `dot push -m "msg"` | 自定义提交信息 |
| `dot add <path>` | 纳入新配置文件为模块 |
| `dot add <path> --name <name>` | 指定模块名 |
| `dot remove <name>` | 卸载模块 |
| `dot doctor` | 检查符号链接健康 |
| `dot clean` | 清理备份文件 |

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
