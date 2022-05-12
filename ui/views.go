package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/skuid/tides/pkg/util"
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
			subtle("[up]/[down]: select"),
			subtle("[enter]: choose"),
			subtle("[esc]: quit/back"),
		},
		dot,
	)

	welcomeHeader = fmt.Sprintf(`Welcome to Skuid's Command Line Interface (CLI): %v`, tides("Tides"))
)

func viewSelect(vm viewModel) string {
	var commands []string
	for i, command := range vm.Command.Commands() {
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
		return skuid(fmt.Sprintf("[x] %s\n%s", name, tides(description)))
	}
	return fmt.Sprintf("[ ] %s\n%s", name, subtle(description))
}

// Update loop for the first view where you're choosing a task.
func updateSelect(msg tea.Msg, vm viewModel) (m tea.Model, c tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			vm.CommandIndex += 1
			if vm.CommandIndex > len(vm.Command.Commands())-1 {
				vm.CommandIndex = len(vm.Command.Commands()) - 1
			}
		case "up":
			vm.CommandIndex -= 1
			if vm.CommandIndex < 0 {
				vm.CommandIndex = 0
			}
		case "enter":
			vm.SelectedCommand = vm.Command.Commands()[vm.CommandIndex]
			vm.State = PREPARE
		}
	}

	m = vm

	return
}

// ------------------------------------------------------------------------------

func updatePrepare(msg tea.Msg, vm viewModel) (m tea.Model, c tea.Cmd) {

	flagLength := len(util.AllFlags(vm.SelectedCommand))
	// flagLength +1
	// when we get to index = flagLength,
	// we want to show the option for "execute"

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			vm.FlagIndex += 1
			// this is different than the others
			// since we have the ability to EXECUTE
			// this command, so the final option
			// will be "execute" and we will use that
			if vm.FlagIndex > flagLength {
				vm.FlagIndex = flagLength
			}
		case "up":
			vm.FlagIndex -= 1
			if vm.FlagIndex < 0 {
				vm.FlagIndex = 0
			}
		case "enter":
			if vm.FlagIndex == flagLength {
				vm.State = RUN
			} else {
				vm.State = EDIT
			}
		}
	}

	m = vm

	return
}

func viewPrepare(vm viewModel) string {
	executionHeader := fmt.Sprintf(`Configure Command: %v`, tides(vm.SelectedCommand.Name()))

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
		executionHeader,
		executeText,
		flagsText,
		helpSelect,
	}, "\n\n")
}

func flagString(flag *pflag.Flag, selected bool) string {
	var selectString string
	var selectHelp string

	if selected {
		selectString = tides(fmt.Sprintf("[x] %v %v", flag.Name, flag.Value.String()))
		selectHelp = subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 2))
	} else {
		selectString = subtle(fmt.Sprintf("[ ] %v %v", flag.Name, flag.Value.String()))
		selectHelp = subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 2))
	}

	return fmt.Sprintf("%v\n%v", selectString, selectHelp)
}

func executeString(selected bool) string {
	if selected {
		return pink(fmt.Sprintf("[x] EXECUTE"))
	}
	return subtle(fmt.Sprintf("[ ] EXECUTE"))
}
