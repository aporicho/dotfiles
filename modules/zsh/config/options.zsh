# ============================================
# Zsh Shell Options
# ============================================

# ============================================
# Directory Navigation
# ============================================

# Change directory by typing directory name (no need for cd)
setopt AUTO_CD

# Automatically push directories onto the directory stack
setopt AUTO_PUSHD

# Don't push duplicate directories onto the stack
setopt PUSHD_IGNORE_DUPS

# Don't print directory stack after pushd/popd
setopt PUSHD_SILENT

# ============================================
# Globbing (Pattern Matching)
# ============================================

# Enable extended globbing (e.g., ^file matches everything except 'file')
setopt EXTENDED_GLOB

# If a pattern has no matches, leave it unchanged rather than error
setopt NO_NOMATCH

# ============================================
# Command Correction
# ============================================

# Suggest corrections for mistyped commands
setopt CORRECT

# Don't suggest corrections for arguments (only commands)
setopt NO_CORRECT_ALL

# ============================================
# Performance
# ============================================

# Disable magic functions for better performance
DISABLE_MAGIC_FUNCTIONS="true"

# ============================================
# Job Control
# ============================================

# Report status of background jobs immediately
setopt NOTIFY

# Don't send HUP signal to background jobs when shell exits
setopt NO_HUP

# ============================================
# Other Useful Options
# ============================================

# Allow comments in interactive shell
setopt INTERACTIVE_COMMENTS

# Beep on error
setopt BEEP
