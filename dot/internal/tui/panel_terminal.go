package tui

import (
	"strings"
	"time"

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
	lines     []outputLine
	input     string
	inputMode bool
	focused   bool
	styles    Styles
	theme     Theme
	scrollOff int // lines scrolled up from bottom (0 = at bottom)
}

// NewTerminal constructs a Terminal panel.
func NewTerminal(styles Styles, theme Theme) *Terminal {
	return &Terminal{
		styles: styles,
		theme:  theme,
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
	// Auto-scroll to bottom when new output arrives.
	t.scrollOff = 0
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
		t.scrollOff = 0
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
	case m.Type == tea.KeyPgUp || m.String() == "k":
		t.scrollOff += 5
		if t.scrollOff > len(t.lines) {
			t.scrollOff = len(t.lines)
		}
	case m.Type == tea.KeyPgDown || m.String() == "j":
		t.scrollOff -= 5
		if t.scrollOff < 0 {
			t.scrollOff = 0
		}
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

	outputArea := t.renderOutput(inner, outputHeight)
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

// renderOutput renders the scrollable output log area.
func (t *Terminal) renderOutput(width, height int) string {
	if height <= 0 {
		return ""
	}

	dimmed := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Dimmed()))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Red()))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.Subtle()))

	// Determine which slice of lines to show.
	total := len(t.lines)

	// viewEnd is the index (exclusive) of the last visible line.
	viewEnd := total - t.scrollOff
	if viewEnd < 0 {
		viewEnd = 0
	}
	viewStart := viewEnd - height
	if viewStart < 0 {
		viewStart = 0
	}

	visible := t.lines[viewStart:viewEnd]

	// Build rendered lines, padding to fill the height.
	rendered := make([]string, height)
	for i := range rendered {
		rendered[i] = "" // blank line placeholder
	}

	for i, ol := range visible {
		ts := dimmed.Render(ol.timestamp.Format("15:04:05") + " ")
		var body string
		if ol.isError {
			body = errStyle.Render("✗ " + ol.text)
		} else {
			body = textStyle.Render(ol.text)
		}
		// Truncate to fit width (timestamp is 10 chars + space = 10 total).
		line := ts + body
		rendered[i] = line
	}

	return strings.Join(rendered, "\n")
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
