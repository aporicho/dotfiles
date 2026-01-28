-- 删除 LazyVim 默认的拼写检查 autocmd
-- LazyVim 会在 markdown/text 等文件类型中自动启用拼写检查
vim.api.nvim_create_autocmd("User", {
  pattern = "VeryLazy",
  once = true,
  callback = function()
    pcall(vim.api.nvim_del_augroup_by_name, "lazyvim_wrap_spell")
  end,
})

-- markdown 文件禁用缩进指引线
vim.api.nvim_create_autocmd("FileType", {
  pattern = "markdown",
  callback = function()
    vim.b.snacks_indent = false
  end,
})
