package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/keys"
	"github.com/skuid/tides/ui/lit"
	"github.com/skuid/tides/ui/style"
	"github.com/spf13/cobra"
)

type main struct {
	quitting bool

	Command *cobra.Command
	index   int
}

func Main(command *cobra.Command) main {
	return main{
		Command: command,
	}
}

func (v main) Init() tea.Cmd {
	return nil
}

func (v main) SelectedCommand() *cobra.Command {
	return v.Command.Commands()[v.index]
}

func (v main) Update(msg tea.Msg) (m tea.Model, c tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case keys.ESC, keys.CTRL_C:
			v.quitting = true
			m = v
			c = tea.Quit
			return
		case keys.UP, keys.DOWN, keys.TAB, keys.SHIFT_TAB:
			cmdLength := len(v.Command.Commands()) - 1
			if s == keys.UP || s == keys.SHIFT_TAB {
				v.index--
			}
			if s == keys.DOWN || s == keys.TAB {
				v.index++
			}
			if v.index > cmdLength {
				v.index = cmdLength
			} else if v.index < 0 {
				v.index = 0
			}
		case keys.ENTER:
			m, c = Configure(&v, v.SelectedCommand())
			return
		}
	}

	m = v
	return
}

func (v main) View() string {
	if v.quitting {
		return indent.String(
			"\n"+
				fmt.Sprintf(style.Skuid("Thank you for using %v!"), style.Tides("Skuid Tides"))+
				style.Subtle(fmt.Sprintf(" (Version: %v) ", constants.VERSION_NAME))+
				"\n\n",
			2)
	}

	var sections []string

	sections = append(sections, lit.MainHeader)
	sections = append(sections, v.Body())
	sections = append(sections, help.SelectionHelp)

	return strings.Join(
		sections,
		"\n\n",
	)
}

func (v main) Body() string {
	var commands []string
	for i, command := range v.Command.Commands() {
		commands = append(commands, style.CommandString(command, v.index == i))
	}
	return strings.Join(commands, "\n")
}
