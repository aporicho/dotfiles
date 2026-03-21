-- ==========================================
-- Dotfiles 统一加载器
-- ==========================================
-- 这个文件会被符号链接到 ~/.config/nvim/lua/plugins/dotfiles.lua
-- 它会自动加载 ~/dotfiles/nvim 中的所有配置和插件
--
-- 使用方式：
--   ln -s ~/dotfiles/nvim/settings.lua ~/.config/nvim/lua/plugins/dotfiles.lua

local dotfiles_root = vim.fn.expand("~/dotfiles/nvim")

-- ==========================================
-- 1. 加载 config/ 目录下的配置文件
-- ==========================================
local function load_configs()
  local config_dir = dotfiles_root .. "/lua/config"

  -- 检查目录是否存在
  if vim.fn.isdirectory(config_dir) == 0 then
    vim.notify("Dotfiles config 目录不存在: " .. config_dir, vim.log.levels.WARN)
    return
  end

  -- 获取所有 .lua 文件
  local config_files = vim.fn.glob(config_dir .. "/*.lua", false, true)

  for _, file in ipairs(config_files) do
    local ok, err = pcall(dofile, file)
    if not ok then
      vim.notify("加载配置失败: " .. vim.fn.fnamemodify(file, ":t") .. "\n" .. tostring(err), vim.log.levels.ERROR)
    end
  end
end

-- ==========================================
-- 2. 加载 plugins/ 目录下的插件规范
-- ==========================================
local function load_plugins()
  local plugins = {}
  local plugins_dir = dotfiles_root .. "/lua/plugins"

  -- 检查目录是否存在
  if vim.fn.isdirectory(plugins_dir) == 0 then
    vim.notify("Dotfiles plugins 目录不存在: " .. plugins_dir, vim.log.levels.WARN)
    return plugins
  end

  -- 获取所有 .lua 文件
  local plugin_files = vim.fn.glob(plugins_dir .. "/*.lua", false, true)

  for _, file in ipairs(plugin_files) do
    local ok, spec = pcall(dofile, file)
    if ok and spec then
      if type(spec) == "table" then
        -- 检查是否是插件数组格式
        if vim.tbl_islist(spec) then
          -- 数组格式：{ {...}, {...} }
          vim.list_extend(plugins, spec)
        else
          -- 单个插件：{ "plugin/name", ... }
          table.insert(plugins, spec)
        end
      end
    else
      local filename = vim.fn.fnamemodify(file, ":t")
      vim.notify("加载插件失败: " .. filename .. "\n" .. tostring(spec), vim.log.levels.WARN)
    end
  end

  return plugins
end

-- ==========================================
-- 3. 返回加载器插件 + 所有动态加载的插件
-- ==========================================
local result = {
  -- 配置加载器（高优先级，早期加载）
  {
    "dotfiles-config-loader",
    dir = dotfiles_root,
    priority = 1000,
    lazy = false, -- 立即加载
    config = function()
      load_configs()
    end,
  },
}

-- 添加所有插件规范
vim.list_extend(result, load_plugins())

return result
