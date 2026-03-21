package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the bubbletea model for the module picker.
type Model struct {
	Items   []ModuleItem
	Cursor  int
	Done    bool
	Quitted bool
}

// Init returns nil; no initial command is needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages for navigation and selection.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.Quitted = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
		case " ":
			if len(m.Items) > 0 {
				m.Items[m.Cursor].Selected = !m.Items[m.Cursor].Selected
			}
		case "enter":
			m.Done = true
			return m, tea.Quit
		case "a":
			m.ToggleAll()
		}
	}
	return m, nil
}

// ToggleAll selects all items if any are unselected, otherwise deselects all.
func (m *Model) ToggleAll() {
	allSelected := true
	for _, item := range m.Items {
		if !item.Selected {
			allSelected = false
			break
		}
	}
	for i := range m.Items {
		m.Items[i].Selected = !allSelected
	}
}

// SelectedNames returns the names of all selected modules.
func (m Model) SelectedNames() []string {
	var names []string
	for _, item := range m.Items {
		if item.Selected {
			names = append(names, item.Name)
		}
	}
	return names
}

// RunPicker launches the interactive TUI picker and returns the selected
// module names. It returns nil if the user cancelled (q/esc).
func RunPicker(items []ModuleItem) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no modules available")
	}

	m := Model{Items: items}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	result := finalModel.(Model)
	if result.Quitted {
		return nil, nil
	}
	return result.SelectedNames(), nil
}
