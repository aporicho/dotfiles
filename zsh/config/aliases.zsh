# ============================================
# Zsh Aliases Configuration
# ============================================

# ============================================
# ls Enhancements
# ============================================

# Use colors with ls
alias ls='ls -G'
alias ll='ls -lah'
alias la='ls -A'
alias l='ls -CF'

# ============================================
# Directory Navigation
# ============================================

# Quick directory navigation
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'

# Directory shortcuts
alias ~='cd ~'
alias -- -='cd -'

# ============================================
# Git Shortcuts
# ============================================

alias g='git'
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias gd='git diff'
alias gco='git checkout'
alias gb='git branch'
alias glog='git log --oneline --graph --decorate'

# ============================================
# Safety Nets
# ============================================

# Confirm before overwriting
alias cp='cp -i'
alias mv='mv -i'
alias rm='rm -i'

# ============================================
# Utilities
# ============================================

# Show hidden files
alias show='defaults write com.apple.finder AppleShowAllFiles -bool true && killall Finder'
alias hide='defaults write com.apple.finder AppleShowAllFiles -bool false && killall Finder'

# Reload zsh configuration
alias reload='source ~/.zshrc'

# Clear screen
alias c='clear'

# ============================================
# macOS Specific
# ============================================

# Flush DNS cache
alias flushdns='sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder'

# Show/hide desktop icons
alias hidedesktop='defaults write com.apple.finder CreateDesktop -bool false && killall Finder'
alias showdesktop='defaults write com.apple.finder CreateDesktop -bool true && killall Finder'
