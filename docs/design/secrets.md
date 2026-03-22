# Secrets 管理

用 age 对称加密管理敏感环境变量。明文 gitignore 忽略，加密后的 `.age` 文件提交到 git。

---

## 文件结构

| 文件 | 用途 | 是否提交 |
|------|------|---------|
| `secrets.env` | 明文，shell export 格式 | 否（gitignore） |
| `secrets.env.age` | age 加密后 | 是 |
| `module.toml` `[[secrets]]` | 声明加密文件 | 是 |

---

## module.toml 声明

```toml
[[secrets]]
source = "secrets.env"
encrypted = "secrets.env.age"
target = "~/.zsh/secrets.env"
```

- 任何模块都可以声明 `[[secrets]]`
- `[[secrets]]` 不替代 `[[links]]`

---

## 加密方式

| 属性 | 值 |
|------|-----|
| 库 | filippo.io/age（Go 原生） |
| 模式 | scrypt passphrase（对称加密） |
| 所有模块 | 共享同一个 passphrase |
| 非确定性 | 同一明文每次加密结果不同 |

---

## Passphrase 存储

| 平台 | 方式 |
|------|------|
| macOS | Keychain（`security` CLI） |
| Linux | secret-service（`secret-tool` CLI） |
| WSL/无桌面 | 每次交互输入 |

---

## CLI 行为

| 命令 | secrets 行为 |
|------|-------------|
| `dot install` | 有 `.age` 无明文 → 解密；成功后提示保存 keychain |
| `dot push` | 明文有变更 → 加密生成新 `.age`；首次加密需设置 passphrase |
| `dot doctor` | 检查：明文存在、权限 0600、符号链接正常、内容与 `.age` 一致 |

---

## 明文文件权限

- 解密后的 `secrets.env` 权限固定 `0600`
- `dot doctor` 检查权限是否正确

---

## Passphrase 交互

| 场景 | 行为 |
|------|------|
| 首次加密 | 输入 + 确认，提示保存 keychain |
| 新机器 install | 输入，成功后提示保存 keychain |
| 日常操作（keychain 有） | 静默，零交互 |
| passphrase 错误 | 最多重试 3 次 |
| keychain 不可用 | 每次手动输入 |

---

## 不在范围内

- `dot secrets` 子命令 **（未实现）**
- passphrase 轮换命令 **（未实现）**
- per-module 独立 passphrase **（未实现）**
