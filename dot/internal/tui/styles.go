package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles derived from a Theme. Created once at dashboard startup.
type Styles struct {
	Panel        lipgloss.Style
	PanelFocused lipgloss.Style
	Header       lipgloss.Style
	Value        lipgloss.Style
	Green        lipgloss.Style
	Yellow       lipgloss.Style
	Red          lipgloss.Style
	Blue         lipgloss.Style
	Purple       lipgloss.Style
	Dimmed       lipgloss.Style
	Chip         lipgloss.Style
	ChipSelected lipgloss.Style
	Button       lipgloss.Style
}

func NewStyles(t Theme) Styles {
	return Styles{
		Panel: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(t.PanelBorder())).
			Border(lipgloss.NormalBorder()),
		PanelFocused: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(t.PanelFocusBorder())).
			Border(lipgloss.NormalBorder()),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Dimmed())).
			Transform(strings.ToUpper),
		Value:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Subtle())),
		Green:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Green())),
		Yellow: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Yellow())),
		Red:    lipgloss.NewStyle().Foreground(lipgloss.Color(t.Red())),
		Blue:   lipgloss.NewStyle().Foreground(lipgloss.Color(t.Blue())),
		Purple: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Purple())),
		Dimmed: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dimmed())),
		Chip: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(t.Dimmed())),
		ChipSelected: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(lipgloss.Color(t.Blue())).
			Bold(true),
		Button: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Align(lipgloss.Center),
	}
}
