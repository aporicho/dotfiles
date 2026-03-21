# ============================================
# Zsh History Configuration
# ============================================

# History file location
HISTFILE=~/.zsh_history

# Number of commands to keep in memory
HISTSIZE=10000

# Number of commands to save to file
SAVEHIST=10000

# ============================================
# History Options
# ============================================

# Record timestamp in history file (format: : <beginning time>:<elapsed seconds>;<command>)
setopt EXTENDED_HISTORY

# Share history between all sessions
setopt SHARE_HISTORY

# Write to history file immediately, not when shell exits
setopt INC_APPEND_HISTORY

# Delete old recorded entry if new entry is a duplicate
setopt HIST_IGNORE_ALL_DUPS

# Don't display duplicates when searching
setopt HIST_FIND_NO_DUPS

# Don't save duplicates to history file
setopt HIST_SAVE_NO_DUPS

# Remove superfluous blanks before recording entry
setopt HIST_REDUCE_BLANKS

# Don't record commands starting with a space
setopt HIST_IGNORE_SPACE

# Expire duplicate entries first when trimming history
setopt HIST_EXPIRE_DUPS_FIRST
