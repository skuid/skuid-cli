package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type configure struct {
	FlagIndex int
	inputs    []textinput.Model
}

func (m configure) Init() tea.Cmd {
	return textinput.Blink
}
