package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/skuid/tides/pkg/constants"
	"github.com/spf13/cobra"
)

// ViewModel is the public method that will be used
// in the initialization call of the interface, injecting
// whatever variables we want to so they're added to the command
//
// with the command pointer, we are also able to update flags
// using the user interface
func ViewModel(cmd *cobra.Command) viewModel {
	return viewModel{
		State:           MAIN_MENU,
		Command:         cmd,
		CommandIndex:    0,
		Quitting:        false,
		SelectedCommand: nil,
	}
}

type state string

const (
	MAIN_MENU = "main"
	PREPARE   = "prep"
	EDIT      = "edit"
	RUN       = "run"
)

type viewModel struct {
	State   state
	Command *cobra.Command

	// this is the index of the thing we're
	// selecting
	CommandIndex int
	FlagIndex    int

	// stateful indicators
	Quitting bool

	SelectedCommand *cobra.Command
}

// this function is called when the
// viewmodel is created
// part of the tea.Model interface
func (vm viewModel) Init() tea.Cmd {
	return nil // this means "no input right now, thanks"
}

// Main update function. This is called every time a message
// goes down the channel. User input (keys) are typically the base
// case for input messages.
// part of the tea.Model interface
func (vm viewModel) Update(msg tea.Msg) (m tea.Model, c tea.Cmd) {
	// Make sure these keys always quit

	quit := func() {
		vm.Quitting = true
		m = vm
		c = tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			quit()
			return
		case "esc":
			switch vm.State {
			case MAIN_MENU:
				quit()
				return
			case PREPARE:
				vm.SelectedCommand = nil
				vm.State = MAIN_MENU
			}
		}
	}

	switch vm.State {
	case MAIN_MENU:
		m, c = updateSelect(msg, vm)
	case PREPARE:
		m, c = updatePrepare(msg, vm)
	}

	return

}

// The main view, which just calls the appropriate sub-view
func (vm viewModel) View() string {
	var s string

	if vm.Quitting {
		return indent.String(
			"\n"+
				fmt.Sprintf(skuid("Thank you for using %v!"), tides("Skuid Tides"))+
				subtle(fmt.Sprintf(" (Version: %v) ", constants.VERSION_NAME))+
				"\n\n",
			2)
	}

	switch vm.State {
	case MAIN_MENU:
		s = viewSelect(vm)
	case PREPARE:
		s = viewPrepare(vm)
	}

	return indent.String("\n"+s+"\n\n", 2)
}
