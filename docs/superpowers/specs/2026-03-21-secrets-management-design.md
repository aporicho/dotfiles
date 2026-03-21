# Secrets Management Design

## Problem

敏感环境变量（GitHub token、Anthropic token 等）硬编码在 `modules/zsh/.zshrc` 中，会随 git 提交暴露。需要一种方案让秘密能安全地提交到 git 并在所有机器间同步。

## Solution

用 age 对称加密（passphrase 模式）管理秘密文件。明文 gitignore 忽略，加密后的 `.age` 文件提交到 git。dot CLI 在 pull/push 时自动处理加解密。

## File Structure

```
modules/zsh/
├── secrets.env            # 明文，gitignore 忽略，权限 0600
├── secrets.env.age        # age 加密后，提交到 git
├── config/env.zsh         # 末尾 source secrets
└── module.toml            # 新增 [[secrets]] 段
```

`secrets.env` 格式：

```bash
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_..."
export ANTHROPIC_BASE_URL="https://claudelike.online/api"
export ANTHROPIC_AUTH_TOKEN="cr_..."
```

## module.toml Extension

新增 `[[secrets]]` 段声明加密文件：

```toml
[[secrets]]
source = "secrets.env"
encrypted = "secrets.env.age"
target = "~/.zsh/secrets.env"
```

任何模块都可以声明自己的 secrets，不限于 zsh。`[[secrets]]` 不替代 `[[links]]`——模块仍需至少一个 `[[links]]`。

## Shell Integration

从 `.zshrc` 删除硬编码 token，在 `config/env.zsh` 末尾加载：

```bash
[ -f ~/.zsh/secrets.env ] && source ~/.zsh/secrets.env
```

## Encryption

使用 Go 库 `filippo.io/age`，scrypt recipient（基于 passphrase）：

```go
recipient, _ := age.NewScryptRecipient(passphrase)
identity, _ := age.NewScryptIdentity(passphrase)
```

选择 passphrase 模式而非公钥模式，因为所有机器共享同一份秘密。

**所有模块共享同一个 passphrase**，keychain 中只存一个 `dot-secrets` 条目。

## Passphrase Storage

优先存入系统 keychain，不可用时回退到交互输入。

- **macOS**: `security add-generic-password -s "dot-secrets" -a "$USER" -w`
- **Linux**: `secret-tool store --label="dot-secrets" app dot`
- **WSL / 无桌面**: 每次交互输入

## CLI Behavior

### dot pull

1. 拉取仓库后扫描所有模块的 `[[secrets]]`
2. 存在 `.age` 但无明文 → 获取 passphrase → 解密生成明文（权限 `0600`）
3. 创建符号链接到 target 路径
4. 首次成功后提示保存 passphrase 到 keychain

离线场景：git pull 失败时仍正常处理本地已有的 `.age` 文件。

### dot push

1. 扫描 `[[secrets]]`，解密 `.age` 后与明文内容比较（age 加密非确定性，不能直接比较密文 hash）
2. 内容有变更 → 获取 passphrase（keychain 优先）→ 加密生成新 `.age`
3. 将 `.age` 加入 git 提交，明文被 gitignore 忽略
4. 若 `.age` 不存在但明文存在 → 视为首次加密，触发 passphrase 设置流程

注意：变更检测本身需要 passphrase 来解密 `.age`。如果 keychain 中有 passphrase 则静默完成；否则需要交互输入。

### dot doctor

检查项：
- 明文文件是否存在（已解密）
- 符号链接是否正常
- 明文与 `.age` 是否一致（解密 `.age` 后与明文做内容比对）
- 明文文件权限是否为 `0600`

## Interactive UX

所有交互融入 bubbletea TUI 风格。

**首次加密**（push 时发现未加密的明文）：
- 提示设置 passphrase + 确认
- 加密后提示保存到 keychain

**新机器 pull**：
- 提示输入 passphrase
- 解密成功后提示保存到 keychain

**日常操作**（keychain 中已有 passphrase）：
- 静默加解密，用户零交互

**passphrase 错误**：
- 最多重试 3 次，之后退出

**keychain 不可用**：
- 提示每次需手动输入，不阻塞流程

## .gitignore Changes

项目根 `.gitignore` 追加：

```
# Decrypted secrets
secrets.env
```

固定命名约定：所有模块的秘密文件统一命名为 `secrets.env`，一条 gitignore 规则覆盖所有模块。

## Migration

1. 从 `.zshrc` 移除硬编码 token
2. 创建 `modules/zsh/secrets.env` 存放这些变量
3. `config/env.zsh` 末尾加 source 行
4. `module.toml` 加 `[[secrets]]` 段
5. `.gitignore` 追加 `secrets.env`
6. 首次 `dot push` 触发加密流程
7. 用 `git filter-repo` 或 BFG 清理 git 历史中的明文 token
8. 撤销已泄��的 token 并生成新的

## Out of Scope

以下功能不在本次实��范围，可作为后续迭代：

- `dot secrets` 子命令（手动 encrypt/decrypt/edit）
- passphrase 轮换命令
- per-module 独立 passphrase

当前 passphrase 轮换的手动路径：解密所有 secrets → 删除 `.age` 文件 → 清除 keychain 条目 → 用新 passphrase 重新 `dot push`。

## Dependencies

- `filippo.io/age` — Go 原生 age 加密库，编译进 dot CLI，无运行时外部依赖
