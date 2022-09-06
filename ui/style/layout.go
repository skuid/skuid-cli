package style

import (
	"strings"

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
			Width(Width)
		// Border(lipgloss.RoundedBorder(), true, true, true, true)

	HeaderStyle = lipgloss.NewStyle().
			Height(HeaderHeight).
		// Border(lipgloss.DoubleBorder(), false, false, true).
		Width(Width - (PadLeft + PadRight))

	ASCII_WIDTH = 50

	BodyStyle = lipgloss.NewStyle().
			Height(Height - (HelpHeight + HeaderHeight)).
			Width((Width - 20) - 3).
			Padding(1)

	LeftStyle = lipgloss.NewStyle().
			Foreground(StyleSkuid.GetForeground()).
			Height(Height - (HelpHeight + HeaderHeight)).
			Width((20) - 3).
			Padding(1)

	HelpStyle = lipgloss.NewStyle().
			Height(HelpHeight).
		// Border(lipgloss.DoubleBorder(), true, false, false, false).
		Width(Width - (PadLeft + PadRight))
	// Align(lipgloss.Right)
)

// StandardView is what we'll use to create the logic that handles the frontend.
// TODO: auto-size the container we're going to use
func StandardView(header string, body string, help string) string {

	styledHeader := HeaderStyle.Render(header)

	styledBody := BodyStyle.Render(body)
	// lipgloss.JoinHorizontal(lipgloss.Left,
	// 	LeftStyle.Render(strings.Join(SKUID_ASCII, "\n")), BodyStyle.Render(body),
	// )

	styledFooter := HelpStyle.Render(help)

	content := strings.Join([]string{
		styledHeader,
		styledBody,
		styledFooter,
	}, "\n")

	return ViewStyle.Render(content)
}
