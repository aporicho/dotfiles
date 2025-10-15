-- 文件自动刷新插件 - 激进模式
-- 只要文件被修改就立即刷新，无论在做什么
return {
  "autoread-config",
  dir = vim.fn.stdpath("config") .. "/lua/plugins",
  name = "autoread",
  priority = 1000, -- 高优先级，确保早期加载
  config = function()
    -- 启用自动读取外部修改的文件
    vim.o.autoread = true
    -- 降低更新时间，加快检测（毫秒）
    vim.o.updatetime = 200

    -- 创建 autoread 组
    vim.api.nvim_create_augroup("AutoReadAggressive", { clear = true })

    -- 在关键事件时检查文件是否改变（避免命令行模式冲突）
    vim.api.nvim_create_autocmd({
      "FocusGained",
      "BufEnter",
      "CursorHold",
      "CursorHoldI",
    }, {
      group = "AutoReadAggressive",
      pattern = "*",
      callback = function()
        -- 避免在命令行模式下执行 checktime（会破坏命令行窗口）
        if vim.fn.mode() ~= "c" then
          vim.cmd("checktime")
        end
      end,
      desc = "检查文件是否被外部修改并自动重新加载",
    })

    -- 使用定时器持续检查（每500ms检查一次，避免过于频繁）
    local timer = vim.uv.new_timer()
    timer:start(
      500,
      500,
      vim.schedule_wrap(function()
        -- 同样避免命令行模式
        if vim.fn.mode() ~= "c" then
          vim.cmd("checktime")
        end
      end)
    )

    -- 文件被外部修改时静默重新加载，不显示警告
    vim.api.nvim_create_autocmd("FileChangedShell", {
      group = "AutoReadAggressive",
      pattern = "*",
      callback = function()
        vim.v.fcs_choice = "" -- 自动选择重新加载
      end,
      desc = "静默重新加载外部修改的文件",
    })

    -- 文件重新加载后的通知（可选，如果觉得烦可以注释掉）
    vim.api.nvim_create_autocmd("FileChangedShellPost", {
      group = "AutoReadAggressive",
      pattern = "*",
      callback = function()
        vim.notify("文件已自动刷新", vim.log.levels.INFO)
      end,
      desc = "文件重新加载后显示通知",
    })
  end,
}
