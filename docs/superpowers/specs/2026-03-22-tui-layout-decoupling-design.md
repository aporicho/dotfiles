# TUI 布局解耦设计

将布局计算从集中式改为模块化：每个面板自己决定列结构，Layout 只管垂直预算。

---

## 问题

改一个面板的列结构（如去掉 ADD 芯片）需要同时修改 layout.go + panel_channel.go + dashboard_view.go。根因是"有多少列、每列多宽"这个知识散布在三个文件中。

---

## 设计

### Layout 只管垂直预算

```go
type Layout struct {
    TotalW, TotalH int
    PanelH         int   // 中间面板可用高度
    ShowOverview   bool
    ShowControls   bool
    ShowFooter     bool
}
```

不再包含 ChipCols、PanelCols、CtrlCols、FooterW。

### 标准接口

每个面板通过 `BuildRow` 产出 `Row`（列宽 + 内容）。

```go
type Row struct {
    Cols     []int
    Contents []string
    Height   int
}
```

### 各面板职责

| 面板 | BuildRow 签名 | 自行决定 |
|------|--------------|---------|
| ChannelStrip | `BuildRow(totalW int) Row` | 芯片数、芯片宽度、填充列 |
| PanelRow | `buildPanelRow(totalW, panelH int, showOverview bool, ...) Row` | 列数（2/3）、比例 |
| Controls | `BuildRow(totalW int, panelRow Row) Row` | 按钮数，但对齐面板分隔符 |
| Footer | `buildFooterRow(totalW int) Row` | 单列宽度 |

### Controls 对齐策略

Controls 接收面板的 `Row` 来对齐分隔符：

- 3 面板 `[ow, sw, tw]` → 按钮 `[ow, sw/2, sw/2, tw]`（Scope 一分为二）
- 2 面板 `[sw, tw]` → 按钮 `[sw/2, sw/2, tw/2, tw/2]`

分隔符扣减（`sw - 1` 对半分）由 Controls 自己计算。

### dashboard_view.go 纯组装

```go
func (d *Dashboard) View() string {
    lay := ComputeLayout(d.width, d.height)

    chipRow   := d.channel.BuildRow(lay.TotalW)
    panelRow  := buildPanelRow(lay.TotalW, lay.PanelH, lay.ShowOverview, d.overview, d.scope, d.terminal)

    rows := []Row{chipRow, panelRow}
    if lay.ShowControls {
        rows = append(rows, d.controls.BuildRow(lay.TotalW, panelRow))
    }
    if lay.ShowFooter {
        rows = append(rows, buildFooterRow(lay.TotalW, d.theme))
    }

    return RenderUnifiedFrame(borderSty, title, lay.TotalW, rows)
}
```

---

## 改动文件清单

| 文件 | 改动 |
|------|------|
| layout.go | 简化为只算 TotalW/TotalH/PanelH/降级标志 |
| panel_channel.go | 新增 `BuildRow(totalW) Row`，内部自行计算列宽、填充、滚动 |
| panel_controls.go | 改 `BuildRow(totalW, panelRow Row) Row`，内部对齐面板分隔符 |
| dashboard_view.go | 简化为收集 Row + 调 RenderUnifiedFrame |
| frame.go | 不变（已有 RenderUnifiedFrame） |
| panel_overview.go | 不变（View(w,h) 接口不变） |
| panel_scope.go | 不变 |
| panel_terminal.go | 不变 |

---

## 验证标准

- 去掉/增加芯片只改 panel_channel.go
- 改芯片宽度只改 panel_channel.go
- 改按钮数只改 panel_controls.go
- 改面板比例只改 buildPanelRow 所在文件
- `go build ./...` 通过
