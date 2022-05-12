package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

type sel struct {
	SubCommandIndex int
}

func (m sel) Init() tea.Cmd {
	return nil
}
