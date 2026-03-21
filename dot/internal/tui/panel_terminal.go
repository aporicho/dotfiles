package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TerminalExecMsg is sent when the user submits a command in the terminal panel.
type TerminalExecMsg struct {
	Input string
}

// outputLine holds a single line of terminal output with a timestamp.
type outputLine struct {
	text      string
	timestamp time.Time
	isError   bool
}

const maxLines = 1000

// Terminal is the output log + command input panel (right side).
type Terminal struct {
	lines    []outputLine
	input    string
	inputMode bool
	focused  bool
	styles   Styles
	theme    Theme
	viewport viewport.Model
}

// NewTerminal constructs a Terminal panel.
func NewTerminal(styles Styles, theme Theme) *Terminal {
	vp := viewport.New(0, 0)
	return &Terminal{
		styles:   styles,
		theme:    theme,
		viewport: vp,
	}
}

// Focused implements Panel.
func (t *Terminal) Focused() bool { return t.focused }

// SetFocus implements Panel.
func (t *Terminal) SetFocus(f bool) { t.focused = f }

// InputMode reports whether the terminal is currently in command-input mode.
func (t *Terminal) InputMode() bool { return t.inputMode }

// AppendOutput adds lines to the terminal output history programmatically.
func (t *Terminal) AppendOutput(text string) {
	for _, line := range strings.Split(text, "\n") {
		t.lines = append(t.lines, outputLine{text: line, timestamp: time.Now()})
	}
	// Trim to max lines.
	if len(t.lines) > maxLines {
		t.lines = t.lines[len(t.lines)-maxLines:]
	}
	// Update viewport content and auto-scroll to bottom.
	t.syncViewport()
	t.viewport.GotoBottom()
}

// syncViewport rebuilds the rendered content string and sets it on the viewport.
func (t *Terminal) syncViewport() {
	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Dimmed()))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Red()))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Subtle()))

	rendered := make([]string, len(t.lines))
	for i, ol := range t.lines {
		ts := dimmed.Render(ol.timestamp.Format("15:04:05") + " ")
		var body string
		if ol.isError {
			body = errStyle.Render(fmt.Sprintf("%s %s", "✗", ol.text))
		} else {
			body = textStyle.Render(ol.text)
		}
		rendered[i] = ts + body
	}
	t.viewport.SetContent(strings.Join(rendered, "\n"))
}

// Update implements Panel.
func (t *Terminal) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch m := msg.(type) {
	case CmdOutputMsg:
		isErr := m.Err != nil
		for _, line := range strings.Split(m.Output, "\n") {
			t.lines = append(t.lines, outputLine{
				text:      line,
				timestamp: time.Now(),
				isError:   isErr,
			})
		}
		if len(t.lines) > maxLines {
			t.lines = t.lines[len(t.lines)-maxLines:]
		}
		t.syncViewport()
		t.viewport.GotoBottom()
		return t, nil

	case tea.KeyMsg:
		if !t.focused {
			return t, nil
		}
		if t.inputMode {
			return t.handleInputMode(m)
		}
		return t.handleNavMode(m)
	}
	return t, nil
}

// handleInputMode handles keys when the user is typing a command.
func (t *Terminal) handleInputMode(m tea.KeyMsg) (Panel, tea.Cmd) {
	switch m.Type {
	case tea.KeyEsc:
		t.inputMode = false
		t.input = ""
	case tea.KeyEnter:
		if t.input != "" {
			cmd := t.input
			t.input = ""
			t.inputMode = false
			return t, func() tea.Msg {
				return TerminalExecMsg{Input: cmd}
			}
		}
	case tea.KeyBackspace, tea.KeyDelete:
		if len(t.input) > 0 {
			// Remove last rune safely.
			runes := []rune(t.input)
			t.input = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes:
		t.input += m.String()
	}
	return t, nil
}

// handleNavMode handles keys when the user is not typing (navigation mode).
func (t *Terminal) handleNavMode(m tea.KeyMsg) (Panel, tea.Cmd) {
	switch {
	case m.String() == ":":
		t.inputMode = true
		return t, nil
	case m.Type == tea.KeyPgUp || m.String() == "k",
		m.Type == tea.KeyPgDown || m.String() == "j":
		var cmd tea.Cmd
		t.viewport, cmd = t.viewport.Update(m)
		return t, cmd
	}
	return t, nil
}

// View implements Panel.
func (t *Terminal) View(width, height int) string {
	// inner dimensions accounting for border (1 char each side).
	inner := width - 2
	if inner < 1 {
		inner = 1
	}
	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Reserve 1 line for the input/hint bar at the bottom.
	outputHeight := innerHeight - 1
	if outputHeight < 0 {
		outputHeight = 0
	}

	// Resize viewport to match current dimensions.
	t.viewport.Width = inner
	t.viewport.Height = outputHeight

	outputArea := t.viewport.View()
	inputBar := t.renderInputBar(inner)

	content := lipgloss.JoinVertical(lipgloss.Left, outputArea, inputBar)

	var borderStyle lipgloss.Style
	if t.focused {
		borderStyle = t.styles.PanelFocused
	} else {
		borderStyle = t.styles.Panel
	}

	return borderStyle.Width(inner).Height(innerHeight).Render(content)
}

// renderInputBar renders the bottom prompt/hint line.
func (t *Terminal) renderInputBar(width int) string {
	if t.inputMode {
		prompt := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Blue())).Render("dot ▸ ")
		cursor := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Subtle())).Render("▋")
		return prompt + t.input + cursor
	}
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Dimmed())).Render("press : to type")
	return hint
}
