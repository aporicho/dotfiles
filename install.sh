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
DOT_BIN="/usr/local/bin/dot"
RELEASE_BASE="https://github.com/aporicho/dotfiles/releases/download/latest"

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

# ── 2. 下载预编译的 dot CLI ───────────────
# 检测系统架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) fail "不支持的架构: $ARCH" ;;
esac

# macOS on Apple Silicon → arm64, Intel Mac → amd64
PLATFORM="${OS}-${ARCH}"
DOWNLOAD_URL="${RELEASE_BASE}/dot-${PLATFORM}"

info "下载 dot CLI (${PLATFORM})..."

# 确保目标目录存在
BIN_DIR="$(dirname "$DOT_BIN")"
if [ ! -d "$BIN_DIR" ]; then
    if command -v sudo &>/dev/null; then
        sudo mkdir -p "$BIN_DIR"
    else
        mkdir -p "$BIN_DIR"
    fi
fi

# 下载二进制（GitHub Releases 在某些网络下可能不通）
if ! curl -fsSL --max-time 30 "$DOWNLOAD_URL" -o /tmp/dot 2>/dev/null; then
    warn "GitHub Releases 下载失败，尝试从原始仓库下载..."
    # 备用：直接通过 raw.githubusercontent.com 走不同 CDN
    curl -fsSL --max-time 60 "https://github.com/aporicho/dotfiles/releases/download/latest/dot-${PLATFORM}" -o /tmp/dot || {
        fail "下载失败，请检查网络或手动安装: $DOWNLOAD_URL"
    }
fi

# 安装到系统 PATH
if [ -w "$BIN_DIR" ]; then
    mv /tmp/dot "$DOT_BIN" && chmod +x "$DOT_BIN"
elif command -v sudo &>/dev/null; then
    # 非交互环境用 sudo -S 从 stdin 读密码（如果设了密码）
    if sudo -n true 2>/dev/null; then
        # sudo 不需要密码（NOPASSWD）
        sudo mv /tmp/dot "$DOT_BIN" && sudo chmod +x "$DOT_BIN"
    else
        # sudo 需要密码 — 提示用户手动执行
        warn "需要管理员权限，请手动执行:"
        echo "  sudo mv /tmp/dot $DOT_BIN"
        echo "  sudo chmod +x $DOT_BIN"
        echo "  或者设置 sudo NOPASSWD"
    fi
fi
ok "dot CLI 已安装: $DOT_BIN"

# ── 3. 安装模块 ──────────────────────────
echo ""
cd "$DOTFILES_DIR"

if [ "${DOT_MODULES:-}" = "all" ]; then
    info "安装全部模块..."
    dot install --all
elif [ -n "${DOT_MODULES:-}" ]; then
    info "安装模块: $DOT_MODULES"
    # shellcheck disable=SC2086
    dot install $DOT_MODULES
else
    echo -e "${GREEN}  选择要安装的模块${NC}"
    echo ""
    dot install
fi

echo ""
ok "全部完成！"
