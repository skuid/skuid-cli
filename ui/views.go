package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/util"
	"github.com/skuid/tides/ui/style"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Choice struct {
	Label    string
	Selected string
}

var (
	helpSelect = strings.Join(
		[]string{
			style.Subtle("[up]/[down]: select"),
			style.Subtle("[enter]: choose"),
			style.Subtle("[esc]: quit/back"),
		},
		style.Dot,
	)

	helpEdit = strings.Join(
		[]string{
			style.Subtle("[enter]: save"),
			style.Subtle("[esc]: revert/back"),
		},
		style.Dot,
	)

	welcomeHeader = fmt.Sprintf(`Welcome to Skuid's Command Line Interface (CLI): %v`, style.Tides("Tides"))
)

func viewSelect(vm viewModel) string {
	var commands []string
	for i, command := range vm.TidesCommand.Commands() {
		commands = append(commands, selectCommandString(command, vm.CommandIndex == i))
	}

	comandsSelection := strings.Join(commands, "\n")

	return strings.Join([]string{
		welcomeHeader,
		comandsSelection,
		helpSelect,
	}, "\n\n")
}

func selectCommandString(cmd *cobra.Command, selected bool) string {
	name := cmd.Name()
	description := indent.String(cmd.Short, 4)
	if selected {
		return style.Skuid(fmt.Sprintf("[x] %s\n%s", name, style.Tides(description)))
	}
	return fmt.Sprintf("[ ] %s\n%s", name, style.Subtle(description))
}

// Update loop for the first view where you're choosing a task.
func updateSelect(msg tea.Msg, vm viewModel) (m tea.Model, c tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			vm.CommandIndex += 1
			if vm.CommandIndex > len(vm.TidesCommand.Commands())-1 {
				vm.CommandIndex = len(vm.TidesCommand.Commands()) - 1
			}
		case "up":
			vm.CommandIndex -= 1
			if vm.CommandIndex < 0 {
				vm.CommandIndex = 0
			}
		case "enter":
			vm.SelectedCommand = vm.TidesCommand.Commands()[vm.CommandIndex]
			vm.State = CONFIGURE
		}
	}

	m = vm

	return
}

// ------------------------------------------------------------------------------

func updateConfigure(msg tea.Msg, vm viewModel) (m tea.Model, c tea.Cmd) {

	flagLength := len(util.AllFlags(vm.SelectedCommand))
	// flagLength +1
	// when we get to index = flagLength,
	// we want to show the option for "execute"

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "down", "tab", "shift+tab":
			// movement
			switch msg.String() {
			case "down", "tab":
				vm.FlagIndex++
			case "up", "shift+tab":
				vm.FlagIndex--
			}

			// protection
			// this is different than the others
			// since we have the ability to EXECUTE
			// this command, so the final option
			// will be "execute" and we will use that
			if vm.FlagIndex > flagLength {
				vm.FlagIndex = flagLength
			} else if vm.FlagIndex < 0 {
				vm.FlagIndex = 0
			}

		case "enter":
			if vm.State == CONFIGURE {
				if vm.FlagIndex == 0 {
					vm.State = RUN
				} else {
					vm.SelectedFlag = util.AllFlags(vm.SelectedCommand)[vm.FlagIndex-1]
					vm.State = EDIT
				}
			}

			if vm.State == EDIT {
				// save the current flag
			}
		}
	}

	m = vm

	return
}

func viewConfigure(vm viewModel) string {
	configureHeader := fmt.Sprintf(`Configure Command: %v`, style.Tides(vm.SelectedCommand.Name()))

	var flagsStrings []string
	for i, flag := range util.AllFlags(vm.SelectedCommand) {
		flagsStrings = append(
			flagsStrings,
			flagString(flag, vm.FlagIndex == i+1),
		)
	}

	executeText := indent.String(executeString(vm.FlagIndex == 0), 2)

	flagsText := indent.String(strings.Join(flagsStrings, "\n"), 2)

	return strings.Join([]string{
		welcomeHeader,
		configureHeader,
		executeText,
		flagsText,
		helpSelect,
	}, "\n\n")
}

func flagString(flag *pflag.Flag, selected bool) string {
	var selectString string

	pad := func(flag string) string {
		return fmt.Sprintf("%-20s", flag)
	}

	if selected {
		selectString = style.Tides(fmt.Sprintf("[x] %v %v", pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	} else {
		selectString = style.Subtle(fmt.Sprintf("[ ] %v %v", pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	}

	return fmt.Sprintf("%v", selectString)
}

func executeString(selected bool) string {
	if selected {
		return style.Pink(fmt.Sprintf("[%v] EXECUTE", constants.CHECKMARK))
	}
	return style.Subtle(fmt.Sprintf("[ ] EXECUTE"))
}
