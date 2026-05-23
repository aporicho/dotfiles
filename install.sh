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

# ── 3. 构建 dot CLI 到系统 PATH ──────────
info "构建 dot CLI..."
cd "$DOTFILES_DIR/dot"
if go build -o /usr/local/bin/dot . 2>/dev/null; then
    ok "dot CLI 已安装: /usr/local/bin/dot"
else
    # macOS 上 /usr/local/bin 通常需要 sudo
    sudo go build -o /usr/local/bin/dot .
    ok "dot CLI 已安装: /usr/local/bin/dot"
fi

# ── 4. 安装模块 ──────────────────────────
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

echo ""
ok "全部完成！"
