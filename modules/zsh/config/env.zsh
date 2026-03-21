# ============================================
# Zsh Environment Variables
# ============================================

# ============================================
# Editor Settings
# ============================================

# Set default editor
export EDITOR='vim'
export VISUAL='vim'

# ============================================
# Language and Locale
# ============================================

# Set UTF-8 encoding
export LANG='en_US.UTF-8'
export LC_ALL='en_US.UTF-8'

# ============================================
# Colors
# ============================================

# Enable colors in terminal
export CLICOLOR=1

# Customize ls colors (BSD ls)
export LSCOLORS='ExGxFxdaCxDaDahbadeche'

# For GNU ls (if installed via homebrew)
# export LS_COLORS='di=1;34:ln=1;36:so=1;35:pi=33:ex=1;32:bd=34;46:cd=34;43:su=30;41:sg=30;46:tw=30;42:ow=34;43'

# ============================================
# Less Configuration
# ============================================

# Make less more friendly for non-text input files
export LESS='-R -F -X'

# Syntax highlighting in less (if source-highlight is installed)
# export LESSOPEN='| /usr/local/bin/src-hilite-lesspipe.sh %s'

# ============================================
# FZF Configuration
# ============================================

# FZF default options (if you want to customize)
export FZF_DEFAULT_OPTS='--height 40% --layout=reverse --border'

# FZF color scheme (matches your terminal theme)
# export FZF_DEFAULT_OPTS='--color=fg:#f8f8f2,bg:#282a36,hl:#bd93f9 --color=fg+:#f8f8f2,bg+:#44475a,hl+:#bd93f9 --color=info:#ffb86c,prompt:#50fa7b,pointer:#ff79c6 --color=marker:#ff79c6,spinner:#ffb86c,header:#6272a4'

# ============================================
# Custom PATH additions
# ============================================

# Add your custom paths here if needed
# export PATH="$HOME/bin:$PATH"
# export PATH="/usr/local/bin:$PATH"

# Dotfiles bin directory (for custom scripts)
export PATH="$HOME/dotfiles/bin:$PATH"

# Secrets (decrypted by dot pull)
[ -f ~/.zsh/secrets.env ] && source ~/.zsh/secrets.env
