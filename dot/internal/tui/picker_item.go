package tui

// ModuleItem represents a module entry in the TUI picker.
type ModuleItem struct {
	Name      string
	Installed bool
	HasUpdate bool
	Selected  bool
}

// StatusText returns a human-readable status string for the module.
func (item ModuleItem) StatusText() string {
	if item.Installed && item.HasUpdate {
		return "\u2713 \u5df2\u5b89\u88c5 \u00b7 \u6709\u66f4\u65b0"
	}
	if item.Installed {
		return "\u2713 \u5df2\u5b89\u88c5 \u00b7 \u65e0\u53d8\u66f4"
	}
	return "\u2717 \u672a\u5b89\u88c5"
}

// IsHighlight returns true if the module deserves visual emphasis,
// i.e. it has a pending update or is not yet installed.
func (item ModuleItem) IsHighlight() bool {
	return item.HasUpdate || !item.Installed
}
