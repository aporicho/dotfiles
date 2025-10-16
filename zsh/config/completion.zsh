# ============================================
# Zsh Completion Configuration
# ============================================

# Load and initialize the completion system
autoload -Uz compinit

# Only regenerate compdump once a day for better performance
if [ "$(date +'%j')" != "$(stat -f '%Sm' -t '%j' ~/.zcompdump 2>/dev/null)" ]; then
  compinit
else
  compinit -C
fi

# ============================================
# Completion Styles
# ============================================

# Case-insensitive completion
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'

# Enable menu selection for completion
zstyle ':completion:*' menu select

# Use colors in completion lists (same as ls)
zstyle ':completion:*:default' list-colors ${(s.:.)LS_COLORS}

# Group results by category
zstyle ':completion:*' group-name ''

# Disable descriptions (no "-- local directory --" messages)
# zstyle ':completion:*:descriptions' format '%F{yellow}-- %d --%f'

# Add warnings when no matches found
zstyle ':completion:*:warnings' format '%F{red}-- no matches found --%f'

# Enable caching to speed up completions
zstyle ':completion:*' use-cache on
zstyle ':completion:*' cache-path ~/.zsh/cache
