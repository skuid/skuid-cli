package style

import (
	"github.com/charmbracelet/lipgloss"
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
