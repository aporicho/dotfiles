# ============================================
# Kitty 终端增强配置
# ============================================

# 检测是否在 Kitty 终端中运行
if [[ "$TERM" == "xterm-kitty" ]]; then
    # Kitty shell 集成
    # (Kitty 会自动注入，这里仅作说明)

    # Kitty 补全支持
    if command -v kitty &> /dev/null; then
        # 加载 Kitty 补全
        autoload -Uz compinit
        if [[ ! -f ~/.zcompdump || ~/.zshrc -nt ~/.zcompdump ]]; then
            compinit
        else
            compinit -C
        fi
        source <(kitty + complete setup zsh 2>/dev/null)
    fi

    # 别名: 在当前目录打开新 Kitty 窗口
    alias kcd='kitty @ launch --type=os-window --cwd=current'

    # 别名: 在新标签打开
    alias kt='kitty @ launch --type=tab --cwd=current'

    # 别名: SSH 时使用 Kitty kitten
    alias kssh='kitty +kitten ssh'

    # 函数: 在 Kitty 窗口中打开文件
    kopen() {
        if [[ -z "$1" ]]; then
            echo "用法: kopen <文件路径>"
            return 1
        fi
        kitty @ launch --type=overlay ${EDITOR:-nvim} "$@"
    }

    # 函数: 快速切换会话
    ksession() {
        local session="$1"
        local session_dir="$HOME/.config/kitty/sessions"

        if [[ -z "$session" ]]; then
            echo "可用会话:"
            if [[ -d "$session_dir" ]]; then
                ls "$session_dir"/*.session 2>/dev/null | xargs -n1 basename | sed 's/.session//' | sed 's/^/  - /'
            else
                echo "  (无会话文件)"
            fi
            return
        fi

        local session_file="$session_dir/$session.session"
        if [[ -f "$session_file" ]]; then
            kitty @ load-config "$session_file"
            echo "✅ 已加载会话: $session"
        else
            echo "❌ 会话文件不存在: $session_file"
            return 1
        fi
    }

    # 函数: 快速重载 Kitty 配置
    kreload() {
        kitty @ load-config
        echo "✅ Kitty 配置已重新加载"
    }

    # 自定义提示符，显示 Kitty 窗口 ID (可选，默认注释)
    # export PS1="%F{blue}[Kitty:%F{yellow}$KITTY_WINDOW_ID%F{blue}]%f $PS1"
fi
