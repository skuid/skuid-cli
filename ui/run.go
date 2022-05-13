package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/keys"
	"github.com/skuid/tides/ui/lit"
	"github.com/skuid/tides/ui/style"
)

type run struct {
	cmd  *cobra.Command
	back *configure

	status string
}

func Run(back *configure, sub *cobra.Command) run {
	v := run{
		cmd:  sub,
		back: back,
	}

	err := v.cmd.RunE(v.cmd, []string{})
	if err != nil {
		v.status = "err encountered: " + err.Error()
	} else {
		v.status = "success"
	}

	return v
}

func (v run) Init() tea.Cmd {
	return nil /*func() tea.Msg {
		return v.cmd.RunE(v.cmd, []string{})
	}*/
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
		case keys.ENTER:
			v.status = "enter"
		}
	case error:
		if msg != nil {
			v.status = fmt.Sprintf("error encountered: %v", msg)
		} else {
			v.status = "success"
		}
	default:
		v.status = "miss"
	}

	m = v

	return
}

func (v run) body() string {
	var s = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#005CB9")).
		PaddingLeft(2).
		PaddingRight(2)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		s.Render(style.SKUID_ASCII),
		style.CommandText(v.cmd),
	)
}

func (v run) output() string {
	return fmt.Sprintf("Status: %v", v.status)
}

func (v run) View() string {
	var sections []string
	sections = append(sections, lit.MainHeader)
	sections = append(sections, v.body())
	sections = append(sections, v.output())
	sections = append(sections, help.SelectionHelp)
	return strings.Join(sections, "\n\n")
}
