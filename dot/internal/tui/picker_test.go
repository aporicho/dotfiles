package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_ToggleSelection(t *testing.T) {
	m := Model{
		Items: []ModuleItem{
			{Name: "kitty", Installed: false},
			{Name: "zsh", Installed: true},
		},
		Cursor: 0,
	}

	// Toggle first item on
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)

	if !m.Items[0].Selected {
		t.Error("expected item 0 to be selected after space key")
	}
	if m.Items[1].Selected {
		t.Error("expected item 1 to remain unselected")
	}

	// Toggle first item off
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)

	if m.Items[0].Selected {
		t.Error("expected item 0 to be deselected after second space key")
	}
}

func TestModel_ToggleAll(t *testing.T) {
	m := Model{
		Items: []ModuleItem{
			{Name: "kitty"},
			{Name: "zsh"},
			{Name: "nvim"},
		},
	}

	// Select all
	m.ToggleAll()
	for i, item := range m.Items {
		if !item.Selected {
			t.Errorf("expected item %d (%s) to be selected after ToggleAll", i, item.Name)
		}
	}

	// Deselect all
	m.ToggleAll()
	for i, item := range m.Items {
		if item.Selected {
			t.Errorf("expected item %d (%s) to be deselected after second ToggleAll", i, item.Name)
		}
	}
}

func TestModel_SelectedNames(t *testing.T) {
	m := Model{
		Items: []ModuleItem{
			{Name: "kitty", Selected: true},
			{Name: "zsh", Selected: false},
			{Name: "nvim", Selected: true},
		},
	}

	names := m.SelectedNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 selected names, got %d", len(names))
	}
	if names[0] != "kitty" {
		t.Errorf("expected first selected name to be 'kitty', got %q", names[0])
	}
	if names[1] != "nvim" {
		t.Errorf("expected second selected name to be 'nvim', got %q", names[1])
	}
}

func TestModuleItem_StatusText(t *testing.T) {
	tests := []struct {
		name     string
		item     ModuleItem
		expected string
	}{
		{
			name:     "installed with update",
			item:     ModuleItem{Installed: true, HasUpdate: true},
			expected: "\u2713 \u5df2\u5b89\u88c5 \u00b7 \u6709\u66f4\u65b0",
		},
		{
			name:     "installed no update",
			item:     ModuleItem{Installed: true, HasUpdate: false},
			expected: "\u2713 \u5df2\u5b89\u88c5 \u00b7 \u65e0\u53d8\u66f4",
		},
		{
			name:     "not installed",
			item:     ModuleItem{Installed: false},
			expected: "\u2717 \u672a\u5b89\u88c5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.StatusText()
			if got != tt.expected {
				t.Errorf("StatusText() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestModuleItem_IsHighlight(t *testing.T) {
	tests := []struct {
		name     string
		item     ModuleItem
		expected bool
	}{
		{
			name:     "not installed highlights",
			item:     ModuleItem{Installed: false},
			expected: true,
		},
		{
			name:     "has update highlights",
			item:     ModuleItem{Installed: true, HasUpdate: true},
			expected: true,
		},
		{
			name:     "installed no update no highlight",
			item:     ModuleItem{Installed: true, HasUpdate: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.IsHighlight()
			if got != tt.expected {
				t.Errorf("IsHighlight() = %v, want %v", got, tt.expected)
			}
		})
	}
}
