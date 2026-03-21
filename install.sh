#!/bin/bash
set -e

# ============================================
# Dotfiles 一键安装脚本
# Usage: curl -fsSL https://raw.githubusercontent.com/aporicho/dotfiles/main/install.sh | bash
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

# ============================================
# 1. 安装 Homebrew
# ============================================
install_homebrew() {
    if command -v brew &>/dev/null; then
        ok "Homebrew 已安装"
    else
        info "安装 Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        # Apple Silicon 需要手动添加 PATH
        if [[ -f /opt/homebrew/bin/brew ]]; then
            eval "$(/opt/homebrew/bin/brew shellenv)"
        fi
        ok "Homebrew 安装完成"
    fi
}

# ============================================
# 2. 安装依赖
# ============================================
install_deps() {
    info "安装命令行工具..."
    brew install neovim starship fzf \
        zsh-autosuggestions zsh-syntax-highlighting zsh-completions

    info "安装字体..."
    brew install --cask font-jetbrains-mono-nerd-font

    # 如果有显示器环境，安装 GUI 应用
    if [[ -n "$DISPLAY" ]] || system_profiler SPDisplaysDataType 2>/dev/null | grep -q "Resolution"; then
        info "检测到显示环境，安装 GUI 应用..."
        brew install --cask kitty neovide
    else
        warn "未检测到显示环境，跳过 kitty/neovide 安装（可手动执行: brew install --cask kitty neovide）"
    fi

    ok "依赖安装完成"
}

# ============================================
# 3. 克隆 dotfiles
# ============================================
clone_dotfiles() {
    if [[ -d "$DOTFILES_DIR" ]]; then
        warn "$DOTFILES_DIR 已存在，拉取最新代码..."
        git -C "$DOTFILES_DIR" pull
    else
        info "克隆 dotfiles..."
        git clone "$REPO_URL" "$DOTFILES_DIR"
    fi
    ok "dotfiles 就绪: $DOTFILES_DIR"
}

# ============================================
# 4. 备份已有配置
# ============================================
backup() {
    local target="$1"
    if [[ -e "$target" && ! -L "$target" ]]; then
        local backup="${target}.backup.$(date +%Y%m%d%H%M%S)"
        warn "备份 $target -> $backup"
        mv "$target" "$backup"
    elif [[ -L "$target" ]]; then
        rm -f "$target"
    fi
}

# ============================================
# 5. 创建符号链接
# ============================================
setup_links() {
    info "创建符号链接..."
    mkdir -p ~/.config

    # Zsh
    backup ~/.zshrc
    backup ~/.zsh
    ln -s "$DOTFILES_DIR/zsh/.zshrc" ~/.zshrc
    ln -s "$DOTFILES_DIR/zsh/config" ~/.zsh
    ok "Zsh 配置已链接"

    # Kitty
    backup ~/.config/kitty
    ln -s "$DOTFILES_DIR/kitty" ~/.config/kitty
    ok "Kitty 配置已链接"

    # Neovide
    backup ~/.config/neovide
    ln -s "$DOTFILES_DIR/neovide" ~/.config/neovide
    ok "Neovide 配置已链接"

    # Neovim (LazyVim)
    if [[ ! -d ~/.config/nvim ]]; then
        info "安装 LazyVim starter..."
        git clone https://github.com/LazyVim/starter ~/.config/nvim
        rm -rf ~/.config/nvim/.git
    fi
    backup ~/.config/nvim/lua/plugins/dotfiles.lua
    ln -s "$DOTFILES_DIR/nvim/settings.lua" ~/.config/nvim/lua/plugins/dotfiles.lua
    ok "Neovim 配置已链接"
}

# ============================================
# 主流程
# ============================================
main() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Dotfiles 一键安装${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    install_homebrew
    install_deps
    clone_dotfiles
    setup_links

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  安装完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "请执行以下命令重新加载配置："
    echo ""
    echo -e "  ${YELLOW}source ~/.zshrc${NC}"
    echo ""
    echo "或重新打开终端即可生效。"
    echo ""
}

main
