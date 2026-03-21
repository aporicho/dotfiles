package tui

import "github.com/charmbracelet/lipgloss"

// Theme uses ANSI 0-15 colors so the terminal's palette drives all styling.
type Theme struct {
	Dark bool
}

func DetectTheme() Theme {
	return Theme{Dark: lipgloss.HasDarkBackground()}
}

func (t Theme) contrastFg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

// Semantic colors
func (t Theme) Green() string  { return "2" }
func (t Theme) Yellow() string { return "3" }
func (t Theme) Blue() string   { return "4" }
func (t Theme) Purple() string { return "5" }
func (t Theme) Cyan() string   { return "6" }
func (t Theme) Red() string    { return "1" }
func (t Theme) Dimmed() string { return "8" }
func (t Theme) Subtle() string { return "7" }

// Panel
func (t Theme) PanelBorder() string      { return "8" }
func (t Theme) PanelFocusBorder() string { return "4" }
func (t Theme) PanelBg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

// Channel Strip
func (t Theme) ChipBorder() string         { return "8" }
func (t Theme) ChipSelectedBorder() string { return "4" }

// Status LEDs
func (t Theme) LedHealthy() string { return "2" }
func (t Theme) LedWarning() string { return "3" }
func (t Theme) LedError() string   { return "1" }

// Controls
func (t Theme) BtnPull() string   { return "4" }
func (t Theme) BtnPush() string   { return "5" }
func (t Theme) BtnDoctor() string { return "2" }
func (t Theme) BtnRemove() string { return "1" }

// Footer
func (t Theme) FooterFg() string { return "8" }
func (t Theme) FooterBg() string { return t.PanelBg() }
