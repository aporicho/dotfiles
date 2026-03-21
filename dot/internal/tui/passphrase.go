package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type PassphraseModel struct {
	prompt  string
	value   string
	confirm bool
	phase   int // 0=first input, 1=confirm input
	first   string
	err     string
	done    bool
	aborted bool
}

func NewPassphraseModel(prompt string, confirm bool) PassphraseModel {
	return PassphraseModel{prompt: prompt, confirm: confirm}
}

func (m PassphraseModel) Init() tea.Cmd { return nil }

// Update uses msg.String() pattern (bubbletea v1.x API), matching picker.go style
func (m PassphraseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			m.done = true
			return m, tea.Quit
		case "enter":
			if len(m.value) == 0 {
				m.err = "passphrase 不能为空"
				return m, nil
			}
			if m.confirm && m.phase == 0 {
				m.first = m.value
				m.value = ""
				m.phase = 1
				m.err = ""
				return m, nil
			}
			if m.confirm && m.phase == 1 && m.value != m.first {
				m.err = "两次输入不一致，请重��输入"
				m.value = ""
				m.phase = 0
				m.first = ""
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.value += msg.String()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m PassphraseModel) View() string {
	var b strings.Builder
	if m.confirm && m.phase == 1 {
		b.WriteString("  确认 Passphrase: ")
	} else {
		b.WriteString(fmt.Sprintf("  %s ", m.prompt))
	}
	b.WriteString(strings.Repeat("•", len(m.value)))
	if m.err != "" {
		b.WriteString(fmt.Sprintf("\n  ✗ %s", m.err))
	}
	b.WriteString("\n")
	return b.String()
}

func (m PassphraseModel) Value() string { return m.value }
func (m PassphraseModel) Aborted() bool { return m.aborted }

func RunPassphraseInput(prompt string, confirm bool) (string, error) {
	m := NewPassphraseModel(prompt, confirm)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	final := result.(PassphraseModel)
	if final.Aborted() {
		return "", fmt.Errorf("已取消")
	}
	return final.Value(), nil
}
