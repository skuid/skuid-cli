package ui

import (
	"fmt"
	"strconv"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

// General stuff for styling the view
var (
	term   = termenv.EnvColorProfile()
	pink   = makeFgStyle("211")
	tides  = makeFgStyle("#80bfff")
	tides2 = makeFgStyle("#007eff")
	skuid  = makeFgStyle("#4da6ff")
	subtle = makeFgStyle("241")
	// dark blue #00448a
	// light blue #ebf5ff
	progressEmpty = subtle(progressEmptyChar)
	dot           = colorFg(" • ", "236")

	// Gradient colors we'll use for the progress bar
	ramp = blend("#B14FFF", "#00FFA3", progressBarWidth)
)

// Convert a colorful.Color to a hexadecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and
// 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Generate a blend of colors.
func blend(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}
