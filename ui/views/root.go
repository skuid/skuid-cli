package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/lit"
	"github.com/skuid/tides/ui/style"
	"github.com/spf13/cobra"
)

type st string

const (
	ROOT st = "root"
)

type main struct {
	state    st
	quitting bool

	command *cobra.Command
	s       *sel
	c       *configure
}

func Main(command *cobra.Command) main {
	return main{
		state:   ROOT,
		command: command,
	}
}

func (r main) Init() tea.Cmd {
	return nil
}

func (r main) Update(msg tea.Msg) (m tea.Model, c tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			r.quitting = true
			return r, tea.Quit
		}
	}

	m = r

	return
}

func (r main) View() string {
	if r.quitting {
		return indent.String(
			"\n"+
				fmt.Sprintf(style.Skuid("Thank you for using %v!"), style.Tides("Skuid Tides"))+
				style.Subtle(fmt.Sprintf(" (Version: %v) ", constants.VERSION_NAME))+
				"\n\n",
			2)
	}

	var sections []string

	sections = append(sections, lit.MainHeader)
	sections = append(sections, help.SelectionHelp)

	return strings.Join(
		sections,
		"\n\n",
	)
}
