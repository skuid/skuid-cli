package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
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
		case "shift+enter":
			vm.State = EXECUTE
		}
	}

	m = vm

	return
}

func updatePrepare(msg tea.Msg, vm viewModel) (tea.Model, tea.Cmd) {
	return vm, nil
}

func viewPrepare(vm viewModel) string {
	executionHeader := fmt.Sprintf(`Configure Command: %v`, tides(vm.SelectedCommand.Name()))

	var globalFlags []*pflag.Flag
	vm.Command.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		globalFlags = append(globalFlags, f)
	})

	var globalFlagText []string
	for _, flag := range globalFlags {
		globalFlagText = append(
			globalFlagText,
			globalFlagString(flag, false),
		)
	}

	globalFlagString := indent.String(strings.Join(globalFlagText, "\n"), 2)

	var flags []*pflag.Flag
	vm.SelectedCommand.Flags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})

	var flagNames []string
	for i, flag := range flags {
		flagNames = append(
			flagNames,
			commandFlagString(flag, vm.FlagIndex == i),
		)
	}

	flagString := indent.String(strings.Join(flagNames, "\n"), 2)

	return strings.Join([]string{
		welcomeHeader,
		executionHeader,
		globalFlagString,
		flagString,
		helpSelect,
	}, "\n\n")
}

func globalFlagString(flag *pflag.Flag, selected bool) string {
	return fmt.Sprintf("%v", flag.Name)
}

func commandFlagString(flag *pflag.Flag, selected bool) string {
	return fmt.Sprintf("%v", flag.Name)
}
