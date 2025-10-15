-- ============================================
-- Tokyo Night 主题配置
-- 与 Kitty 终端主题保持一致
-- ============================================

return {
  -- 安装 Tokyo Night 主题
  {
    "folke/tokyonight.nvim",
    lazy = false, -- 启动时立即加载
    priority = 1000, -- 确保在其他插件之前加载
    opts = {
      -- 主题变体：night, storm, day, moon
      style = "night", -- 使用 night 变体（深蓝黑背景 #1a1b26，与 Kitty 一致）

      -- 背景设置
      transparent = false, -- 不透明背景
      terminal_colors = true, -- 配置终端颜色以匹配主题

      -- 代码样式
      styles = {
        comments = { italic = true }, -- 斜体注释
        keywords = { italic = true }, -- 斜体关键字
        functions = {}, -- 函数默认样式
        variables = {}, -- 变量默认样式
        -- 其他选项：sidebars, floats 可设置为 "dark", "transparent", "normal"
      },

      -- 侧边栏应用深色背景
      sidebars = { "qf", "help", "vista_kind", "terminal", "packer", "NvimTree" },

      -- 日间亮度（仅 day 主题有效）
      day_brightness = 0.3,

      -- 其他设置
      hide_inactive_statusline = false, -- 不隐藏非活动窗口的状态栏
      dim_inactive = false, -- 不使非活动窗口变暗
      lualine_bold = true, -- 状态栏粗体

      -- 高级配置：自定义颜色覆盖（可选）
      ---@param colors ColorScheme
      on_colors = function(colors)
        -- 可以在这里自定义颜色
        -- 例如：colors.hint = colors.orange
      end,

      ---@param highlights Highlights
      ---@param colors ColorScheme
      on_highlights = function(highlights, colors)
        -- 可以在这里自定义高亮组
        -- 例如：highlights.Comment = { fg = colors.green }
      end,
    },
  },

  -- 配置 LazyVim 使用 Tokyo Night 主题
  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = "tokyonight-night",
    },
  },
}
