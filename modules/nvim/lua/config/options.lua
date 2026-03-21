-- Options are automatically loaded before lazy.nvim startup
-- Default options that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/options.lua
-- Add any additional options here

-- ============================================
-- 通用选项
-- ============================================
vim.opt.spell = false -- 关闭拼写检查
vim.opt.number = false -- 关闭行号
vim.opt.relativenumber = false -- 关闭相对行号
vim.opt.signcolumn = "auto" -- 有符号时才显示符号列
vim.opt.statuscolumn = "" -- 禁用自定义状态列

-- ============================================
-- Neovide GUI 配置
-- ============================================
if vim.g.neovide then
  -- 字体配置（与 Kitty 一致）
  vim.o.guifont = "JetBrainsMono Nerd Font Mono:h16"

  -- Neovide 特定设置
  vim.g.neovide_scale_factor = 1.0 -- 缩放因子

  -- 动画效果
  vim.g.neovide_cursor_animation_length = 0.05 -- 光标动画时长
  vim.g.neovide_cursor_trail_size = 0.3 -- 光标轨迹大小
  vim.g.neovide_cursor_antialiasing = true -- 光标抗锯齿

  -- 光标特效（可选）
  vim.g.neovide_cursor_vfx_mode = "ripple" -- 光标特效：railgun, torpedo, pixiedust, sonicboom, ripple, wireframe
  vim.g.neovide_cursor_vfx_opacity = 200.0 -- 特效透明度
  vim.g.neovide_cursor_vfx_particle_density = 7.0 -- 粒子密度

  -- 性能设置
  vim.g.neovide_refresh_rate = 60 -- 刷新率
  vim.g.neovide_refresh_rate_idle = 5 -- 空闲时刷新率
  vim.g.neovide_no_idle = false -- 允许空闲优化

  -- 输入设置
  vim.g.neovide_input_macos_option_key_is_meta = "both" -- macOS Option 键作为 Meta（both/only_left/only_right）
  vim.g.neovide_touch_deadzone = 6.0 -- 触摸死区
  vim.g.neovide_touch_drag_timeout = 0.17 -- 触摸拖拽超时

  -- 窗口设置
  -- 注意：macOS 的窗口框架设置需要在 ~/.config/neovide/config.toml 中配置
  vim.g.neovide_remember_window_size = true -- 记住窗口大小
  vim.g.neovide_fullscreen = false -- 启动时不全屏
  vim.g.neovide_hide_mouse_when_typing = true -- 输入时隐藏鼠标
  vim.g.neovide_confirm_quit = true -- 退出时确认（有未保存更改）

  -- macOS 全屏模式（隐藏 Dock 和菜单栏）
  -- vim.g.neovide_macos_simple_fullscreen = false

  -- 输入法设置
  vim.g.neovide_input_ime = true -- 启用输入法支持
end
