package style

import (
	"github.com/charmbracelet/lipgloss"
)

// lipgloss style layouts

var (
	Width        = 100
	Height       = 30
	HeaderHeight = 2
	HelpHeight   = 1
	PadLeft      = 1
	PadRight     = 1

	ViewStyle = lipgloss.NewStyle().
			Height(Height).
			PaddingLeft(PadLeft).
			PaddingRight(PadRight).
			Width(Width).
			Border(lipgloss.RoundedBorder(), true, true, true, true)

	HeaderStyle = lipgloss.NewStyle().
			Height(HeaderHeight).
			Border(lipgloss.DoubleBorder(), false, false, true).
			Width(Width - (PadLeft + PadRight)).
			Align(lipgloss.Center)

	BodyStyle = lipgloss.NewStyle().
			Height(Height - (HelpHeight + HeaderHeight)).
			Padding(1)

	HelpStyle = lipgloss.NewStyle().
			Height(HelpHeight).
			Border(lipgloss.DoubleBorder(), true, false, false, false).
			Width(Width - (PadLeft + PadRight)).
			Align(lipgloss.Right)
)

// lipgloss styles colors
var (
	StyleFocus  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	StyleBlur   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StyleSave   = lipgloss.NewStyle().Foreground(lipgloss.Color("050"))
	StyleClear  = lipgloss.NewStyle().Foreground(lipgloss.Color("500"))
	StylePink   = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	StyleTides  = lipgloss.NewStyle().Foreground(lipgloss.Color("#80bfff"))
	StyleTides2 = lipgloss.NewStyle().Foreground(lipgloss.Color("#007eff"))
	StyleSkuid  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4da6ff"))
	StyleSubtle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	StyleNone   = lipgloss.NewStyle()
	StyleHelp   = StyleBlur.Copy()
)

// General stuff for styling the view
var (
	Dot = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(" • ")
)
