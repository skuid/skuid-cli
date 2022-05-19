package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// Main is going to be the entry that takes the cmd
// pointer and returns the model we want to use to initialize
// the user interface
func Main(command *cobra.Command) main {
	return main{
		Command: command,
	}
}

// Init is called when the model is created
func (v main) Init() tea.Cmd {
	return nil
}

// Update is the main function that handles all tea.Msgs.
// Think of this as the event listener for all inputs within
// the machine. So we need to handle input, that's all.
// NOTE: if you don't assign m = v, you'll get a nil pointer.
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
			cmdLength := len(v.getCommands()) - 1
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
				v.selectedCommand(),
			)
			return
		}
	}

	m = v
	return
}

// View is what's rendered in the command line. This string
// is styled and assembled every time that there's a change,
// so you want to check state
func (v main) View() string {

	if v.quitting {
		return "\n" +
			fmt.Sprintf(style.StyleSkuid.Render("Thank you for using %v!"), style.StyleTides.Render("Skuid Tides")) +
			style.StyleSubtle.Render(fmt.Sprintf(" (Version: %v) ", constants.VERSION_NAME)) +
			"\n\n"
	}

	return style.StandardView(
		lit.MainHeader,
		v.body(),
		help.SelectionHelp,
	)
}

// returns the selected command that we're going to
// pass to the next model
func (v main) selectedCommand() *cobra.Command {
	return v.getCommands()[v.index]
}

// GetSubcommands gets all the commands EXCEPT FOR help
func (v main) getCommands() (nonHelpCommands []*cobra.Command) {
	for _, sub := range v.Command.Commands() {
		if !strings.EqualFold(sub.Name(), "help") {
			nonHelpCommands = append(nonHelpCommands, sub)
		}
	}
	return
}

func (v main) body() string {
	var commands []string
	for i, command := range v.getCommands() {
		commands = append(commands, style.CommandString(command, v.index == i))
	}
	return strings.Join(commands, "\n\n")
}
