package style

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
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
	StyleFocus = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	StyleBlur  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StyleSave  = lipgloss.NewStyle().Foreground(lipgloss.Color("050"))
	StyleClear = lipgloss.NewStyle().Foreground(lipgloss.Color("500"))
	StyleNone  = lipgloss.NewStyle()
	StyleHelp  = StyleBlur.Copy()
)

// General stuff for styling the view
var (
	term   = termenv.EnvColorProfile()
	Pink   = MakeFgStyle("211")
	Tides  = MakeFgStyle("#80bfff")
	Tides2 = MakeFgStyle("#007eff")
	Skuid  = MakeFgStyle("#4da6ff")
	Subtle = MakeFgStyle("241")
	// dark blue #00448a
	// light blue #ebf5ff
	Dot = ColorFg(" • ", "236")

	// Gradient colors we'll use for the progress bar
	// ramp = Blend("#B14FFF", "#00FFA3", progressBarWidth)
)

// Convert a colorful.Color to a hexadecimal format compatible with termenv.
func ColorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", ColorFloatToHex(c.R), ColorFloatToHex(c.G), ColorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and
// 1.
func ColorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}

// Color a string's foreground with the given value.
func ColorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Generate a blend of colors.
func Blend(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, ColorToHex(c))
	}
	return
}

// Return a function that will colorize the foreground of a given string.
func MakeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

// Color a string's foreground and background with the given value.
func MakeFgBgStyle(fg, bg string) func(string) string {
	return termenv.Style{}.
		Foreground(term.Color(fg)).
		Background(term.Color(bg)).
		Styled
}
