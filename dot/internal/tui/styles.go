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
			Border(lipgloss.RoundedBorder()),
		PanelFocused: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(t.PanelFocusBorder())).
			Border(lipgloss.RoundedBorder()),
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
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.ChipBorder())).
			Padding(0, 1),
		ChipSelected: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.ChipSelectedBorder())).
			Padding(0, 1),
		Button: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Align(lipgloss.Center),
	}
}
