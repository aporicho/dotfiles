-- Snacks explorer 文件浏览器配置
return {
  "folke/snacks.nvim",
  opts = {
    statuscolumn = { enabled = false }, -- 禁用 Snacks 状态列，减少左边距
    picker = {
      sources = {
        explorer = {
          layout = {
            layout = {
              width = 20,
            },
          },
        },
      },
    },
  },
}
