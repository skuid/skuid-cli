package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type startLoading struct{}

func load() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return startLoading{}
	})
}
