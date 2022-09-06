package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// Run is the public function that we use to export the model
// to the other view
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

// TODO:
// replace View() logic with styled options
func (v run) View() string {
	return style.StandardView(lit.MainHeader, v.body(), help.SelectionHelp)
}

func (v run) body() string {
	return strings.Join(
		[]string{
			style.CommandText(v.cmd),
			v.output(),
		},
		"\n",
	)
}

func (v run) output() string {
	return fmt.Sprintf("Status: %v", v.status)
}
