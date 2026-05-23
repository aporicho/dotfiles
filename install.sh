#!/usr/bin/env bash
set -euo pipefail

# ============================================
# Dotfiles 引导安装脚本
# Usage:
#   交互式选择模块:
#     curl -fsSL https://raw.githubusercontent.com/aporicho/dotfiles/main/install.sh | bash
#
#   全自动安装指定模块（推荐）:
#     curl -fsSL https://raw.githubusercontent.com/aporicho/dotfiles/main/install.sh | DOT_MODULES="kitty zsh" bash
#
#   安装全部模块:
#     curl -fsSL https://raw.githubusercontent.com/aporicho/dotfiles/main/install.sh | DOT_MODULES="all" bash
# ============================================

DOTFILES_DIR="$HOME/dotfiles"
REPO_URL="https://github.com/aporicho/dotfiles.git"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Dotfiles 引导安装${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ── 1. 克隆或更新仓库 ─────────────────────
if [ -d "$DOTFILES_DIR" ]; then
    warn "$DOTFILES_DIR 已存在，拉取最新代码..."
    cd "$DOTFILES_DIR" && git pull
else
    info "克隆 dotfiles 仓库..."
    git clone "$REPO_URL" "$DOTFILES_DIR"
fi
ok "dotfiles 就绪: $DOTFILES_DIR"

# ── 2. 安装 Go（如果没有）─────────────────
if ! command -v go &>/dev/null; then
    info "安装 Go..."
    if [[ "$(uname)" == "Darwin" ]]; then
        if ! command -v brew &>/dev/null; then
            info "先安装 Homebrew..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            if [[ -f /opt/homebrew/bin/brew ]]; then
                eval "$(/opt/homebrew/bin/brew shellenv)"
            fi
        fi
        brew install go
    else
        sudo apt-get update && sudo apt-get install -y golang
    fi
    ok "Go 安装完成"
else
    ok "Go 已安装: $(go version)"
fi

# ── 3. 构建 dot CLI ──────────────────────
info "构建 dot CLI..."
mkdir -p "$HOME/bin"
cd "$DOTFILES_DIR/dot"
go build -o "$HOME/bin/dot" .
ok "dot CLI 已构建: $HOME/bin/dot"

# ── 4. 永久添加 ~/bin 到 PATH ────────────
if ! grep -q 'export PATH="$HOME/bin:$PATH"' "$HOME/.zshenv" 2>/dev/null; then
    echo 'export PATH="$HOME/bin:$PATH"' >> "$HOME/.zshenv"
    export PATH="$HOME/bin:$PATH"
    ok "已将 ~/bin 永久添加到 PATH（~/.zshenv）"
else
    ok "~/bin 已在 PATH 中"
fi

# ── 5. 安装模块 ──────────────────────────
echo ""
cd "$DOTFILES_DIR"

if [ "${DOT_MODULES:-}" = "all" ]; then
    info "安装全部模块..."
    dot pull --all
elif [ -n "${DOT_MODULES:-}" ]; then
    info "安装模块: $DOT_MODULES"
    # shellcheck disable=SC2086
    dot pull $DOT_MODULES
else
    echo -e "${GREEN}  选择要安装的模块${NC}"
    echo ""
    dot pull
fi

# ── 6. 确保 dot 命令立即可用 ─────────────
# 尝试创建全局 symlink（macOS /usr/local/bin 默认在 PATH 中）
if command -v dot &>/dev/null; then
    ok "dot 命令已可用"
else
    if [[ "$(uname)" == "Darwin" ]] && [ -w /usr/local/bin ] 2>/dev/null; then
        ln -sf "$HOME/bin/dot" /usr/local/bin/dot
        ok "已创建全局 symlink: /usr/local/bin/dot"
    elif [[ "$(uname)" == "Darwin" ]] && command -v sudo &>/dev/null; then
        sudo ln -sf "$HOME/bin/dot" /usr/local/bin/dot 2>/dev/null && \
            ok "已创建全局 symlink: /usr/local/bin/dot" || \
            warn "无法写入 /usr/local/bin（无权限）"
    fi
fi

echo ""
ok "全部完成！"
echo ""
echo -e "${YELLOW}━━━ 使用提示 ━━━${NC}"
if command -v dot &>/dev/null; then
    echo "  直接运行:  dot pull"
else
    echo "  完整路径:   $HOME/bin/dot pull"
    echo "  或执行一下让 PATH 生效:"
    echo "    source ~/.zshenv"
    echo "  然后就能直接用:  dot pull"
fi
echo "  重启终端后也会自动生效（~/.zshenv）"
