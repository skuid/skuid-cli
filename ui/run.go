package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/keys"
	"github.com/skuid/tides/ui/lit"
)

type run struct {
	cmd  *cobra.Command
	back *configure
}

func (v run) Init() tea.Cmd {
	return nil
}

func (v run) Update(msg tea.Msg) (m tea.Model, c tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case keys.CTRL_C:
			v.back.back.quitting = true
			m = v.back.back
			return m, tea.Quit
		case keys.ESC:
			m = v.back
			return
		}
	}

	m = v

	return
}

func (v run) Body() string {
	return "we runnin."
}

func (v run) View() string {
	var sections []string
	sections = append(sections, lit.MainHeader)
	sections = append(sections, v.Body())
	sections = append(sections, help.SelectionHelp)
	return strings.Join(sections, "\n\n")
}
