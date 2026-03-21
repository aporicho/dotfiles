-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here

-- ============================================
-- 文件浏览器快捷键配置
-- ============================================

-- 使用 Command-e 打开/关闭文件浏览器（所有模式）
vim.keymap.set({ "n", "i", "v", "t" }, "<D-e>", function()
  Snacks.explorer()
end, { desc = "Toggle Explorer" })

-- ============================================
-- 终端快捷键配置
-- ============================================

-- 安全删除默认的 Ctrl-/ 映射（如果存在）
pcall(vim.keymap.del, "n", "<C-/>")
pcall(vim.keymap.del, "t", "<C-/>")

-- 使用 Command-/ 切换浮动终端（所有模式）
vim.keymap.set({ "n", "i", "v", "t" }, "<D-/>", function()
  Snacks.terminal.toggle(nil, {
    cwd = vim.uv.cwd(),
    env = { TERMINAL_TYPE = "float" }, -- 标识为浮动终端
    win = {
      wo = {
        winbar = "", -- 隐藏标题栏
      },
    },
  })
end, { desc = "Toggle Terminal (float)" })

-- 使用 Command-t 切换右侧垂直终端（所有模式）
vim.keymap.set({ "n", "i", "v", "t" }, "<D-t>", function()
  Snacks.terminal.toggle(nil, {
    cwd = vim.uv.cwd(),
    env = { TERMINAL_TYPE = "right" }, -- 标识为右侧终端
    win = {
      position = "right",
      width = 0.4, -- 占屏幕宽度的40%
      wo = {
        winbar = "", -- 隐藏标题栏
      },
    },
  })
end, { desc = "Toggle Terminal (right split)" })
