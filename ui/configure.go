package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/skuid/tides/pkg/util"
	"github.com/skuid/tides/ui/help"
	"github.com/skuid/tides/ui/keys"
	"github.com/skuid/tides/ui/lit"
	"github.com/skuid/tides/ui/style"
)

var (
	StyleFocus = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	StyleBlur  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	StyleSave  = lipgloss.NewStyle().Foreground(lipgloss.Color("050"))
	StyleClear = lipgloss.NewStyle().Foreground(lipgloss.Color("500"))
	StyleNone  = lipgloss.NewStyle()
	StyleHelp  = StyleBlur.Copy()
)

const (
	FLAG_STRING       = "string"
	FLAG_STRING_ARRAY = "stringArray"
	FLAG_BOOL         = "bool"
	FLAG_INT          = "int"

	RUN_INDEX = 0 // first thing
)

type configure struct {
	// how we get back to the main view if we hit [esc]
	back *main

	// subcommand we're in on
	sub *cobra.Command

	// selector index
	index int

	// these are all the inputs for whatever command we're using
	inputs []textinput.Model

	// hiddenIndices are basically for password fields
	// and other keys we don't want visible when we hover
	// over something
	hiddenIndices []int
}

func Configure(back *main, cmd *cobra.Command) (v configure, c tea.Cmd) {
	v = configure{
		back: back,
		sub:  cmd,
	}
	v.inputs = make([]textinput.Model, len(v.getFlags()))
	v.createFlags()
	v.reset()
	c = v.focus(0)
	return
}

func (v configure) createFlags() {
	for i, flag := range v.getFlags() {
		if strings.ToLower(flag.Name) == "password" {
			v.hiddenIndices = append(v.hiddenIndices, i)
		}
		v.inputs[i] = textinput.New()
		v.inputs[i].CursorStyle = StyleFocus
		v.inputs[i].Prompt = style.Pad(fmt.Sprintf("%v:", flag.Name))
	}
}

func (v configure) Init() tea.Cmd {
	return textinput.Blink
}

func (v configure) View() string {
	var sections []string
	sections = append(sections, lit.MainHeader)
	sections = append(sections, v.body())
	sections = append(sections, help.EditingHelp)
	return strings.Join(sections, "\n\n")
}

func (v configure) Update(msg tea.Msg) (m tea.Model, c tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case keys.CTRL_C:
			v.back.quitting = true
			m = v.back
			return m, tea.Quit
		case keys.ESC:
			m = v.back
			return
		case keys.UP, keys.DOWN, keys.SHIFT_TAB, keys.TAB:
			bounds := len(v.getFlags())
			if s == keys.UP || s == keys.SHIFT_TAB {
				v.index--
			} else if s == keys.DOWN || s == keys.TAB {
				v.index++
			}
			if v.index < 0 {
				v.index = 0
			} else if v.index > bounds {
				v.index = bounds
			}
			v.focus(v.index)
			v.reset()
		case keys.ENTER:
			if v.index == RUN_INDEX {
				m, c = v.run()
			} else {
				m, c = v.save(v.index)
			}
			return
		case keys.CTRL_S:
			m, c = v.save(v.index)
			return
		default:
			v.focus(v.index)
		}

	}

	var cmds = make([]tea.Cmd, len(v.inputs))
	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}

	m = v
	c = tea.Batch(cmds...)

	return
}

func (v configure) focus(index int) (cmd tea.Cmd) {
	focusIndex := index - 1 // because EXECUTE is first

	for i := range v.inputs {
		if i == focusIndex {
			cmd = v.inputs[i].Focus()
			v.inputs[i].PromptStyle = StyleFocus
			v.inputs[i].TextStyle = StyleFocus
		} else {
			v.inputs[i].Blur()
			v.inputs[i].PromptStyle = StyleNone
			v.inputs[i].TextStyle = StyleNone
		}
	}
	return
}

func (v configure) save(index int) (m tea.Model, cmd tea.Cmd) {
	saveIndex := index - 1 // because EXECUTE is first
	input := v.inputs[saveIndex]
	flag := v.getFlags()[saveIndex]
	switch f := flag.Value.(type) {
	// need to handle the slice value separately
	// the default behavior when calling "flag.Value.Set()"
	// on a slice is to *FREAKING APPEND IT*.
	// this doesn't allow us to update the value at all.
	case pflag.SliceValue:
		array := strings.Split(input.Value(), ",")
		f.Replace(array)
	default:
		flag.Value.Set(input.Value())
	}

	v.inputs[saveIndex].TextStyle = StyleSave
	v.inputs[saveIndex].PromptStyle = StyleSave

	m = v
	return
}

func (v configure) run() (m tea.Model, cmd tea.Cmd) {
	r := run{
		v.sub,
		&v,
	}

	m = r

	return
}

var (
	passwordPlaceholder = strings.Repeat("•", 8)
)

func isPassword(flag *pflag.Flag) bool {
	return strings.EqualFold(flag.Name, "password")
}

func (v configure) reset() (m tea.Model, cmd tea.Cmd) {
	for i, flag := range v.getFlags() {
		if isPassword(flag) && v.index != i+1 {
			v.inputs[i].SetValue(passwordPlaceholder)
		} else {
			v.inputs[i].SetValue(style.RemoveBrackets(flag.Value.String()))
			v.inputs[i].CursorEnd() // otherwise when you hit a password you are in the middle of it
		}
	}
	return
}

func (v configure) getFlags() []*pflag.Flag {
	return util.AllFlags(v.sub)
}

func commandText(cmd *cobra.Command) string {
	var command []string
	command = append(command, cmd.Name())
	for _, flag := range util.AllFlags(cmd) {
		s := flag.Value.String()
		if s != "" &&
			s != "false" && /* bool */
			s != "[]" /* array */ {
			if isPassword(flag) {
				s = passwordPlaceholder
			}
			command = append(command, fmt.Sprintf("\t--%v %v", flag.Name, s))
		}
	}
	return strings.Join(command, "\n")
}

// body is going to show us all of the inputs possible
// for this
func (v configure) body() string {
	var lines []string
	// add execute checkbox
	lines = append(lines, style.Checkbox("run the command ", v.index == 0, false))
	lines = append(lines, style.HighlightIf(indent.String("tides "+commandText(v.sub), 2), v.index == 0))
	lines = append(lines, style.Subtle(strings.Repeat("-", 60)))
	lines = append(lines, style.HighlightIf("flags:", v.index != 0))
	// add text inputs
	for i := range v.inputs {
		lines = append(lines, v.inputs[i].View())
	}
	return indent.String(strings.Join(lines, "\n"), 4)
}
