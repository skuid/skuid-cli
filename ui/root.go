package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/keys"
	"github.com/skuid/tides/ui/lit"
	"github.com/skuid/tides/ui/style"
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
	return GetSubcommands(v.Command)[v.index]
}

// GetSubcommands gets all the commands EXCEPT FOR help
func GetSubcommands(cmd *cobra.Command) (nonHelpCommands []*cobra.Command) {
	for _, sub := range cmd.Commands() {
		if !strings.EqualFold(sub.Name(), "help") {
			nonHelpCommands = append(nonHelpCommands, sub)
		}
	}
	return
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
			cmdLength := len(GetSubcommands(v.Command)) - 1
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
			m, c = Configure(
				&v,
				v.SelectedCommand(),
			)
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
	sections = append(sections, v.body())
	sections = append(sections, help.SelectionHelp)

	return strings.Join(
		sections,
		"\n\n",
	)
}

func (v main) body() string {
	var commands []string
	for i, command := range GetSubcommands(v.Command) {
		commands = append(commands, style.CommandString(command, v.index == i))
	}
	return strings.Join(commands, "\n")
}
