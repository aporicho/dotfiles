# ============================================
# Zsh Configuration (Modular Setup)
# ============================================

# ============================================
# Environment Variables
# ============================================
# Load environment variables first
source ~/.zsh/env.zsh

# ============================================
# Shell Options
# ============================================
# Load shell options and settings
source ~/.zsh/options.zsh

# ============================================
# History Configuration
# ============================================
# Load history settings
source ~/.zsh/history.zsh

# ============================================
# Completion System
# ============================================
# Add zsh-completions to fpath (before loading completion system)
fpath=(/opt/homebrew/share/zsh-completions $fpath)

# Load completion configuration (includes compinit with caching)
source ~/.zsh/completion.zsh

# ============================================
# Aliases
# ============================================
# Load custom aliases
source ~/.zsh/aliases.zsh

# ============================================
# Third-party Tools
# ============================================

# FZF - Fuzzy finder
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# Zsh autosuggestions
source /opt/homebrew/share/zsh-autosuggestions/zsh-autosuggestions.zsh

# Starship prompt (fast, customizable prompt)
eval "$(starship init zsh)"

# ============================================
# Syntax Highlighting (MUST BE LAST!)
# ============================================
# Load zsh-syntax-highlighting at the end for proper functionality
source /opt/homebrew/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh

# Kitty 终端集成
if [[ -f ~/.config/kitty/kitty.zsh ]]; then
    source ~/.config/kitty/kitty.zsh
fi
export PATH="$HOME/.local/bin:$PATH"

# bun completions
[ -s "/Users/aporicho/.bun/_bun" ] && source "/Users/aporicho/.bun/_bun"

# bun
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
# >>> gt >>>
export PATH="$HOME/.local/bin:$PATH"
__gt_cd() {
    local tmp="/tmp/gt_lastdir"
    [ -f "$tmp" ] || return
    local dir="$(cat "$tmp")"
    rm -f "$tmp"
    cd "$dir"
}
gt() {
    if [ $# -eq 0 ]; then command gt && __gt_cd; else command gt "$@"; fi
}
gtc() {
    if [ $# -eq 0 ]; then command gtc && __gt_cd; else command gtc "$@"; fi
}
# <<< gt <<<
eval "$(mark completion zsh)"

# OpenClaw Completion
source "/Users/aporicho/.openclaw/completions/openclaw.zsh"
