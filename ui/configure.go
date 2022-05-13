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
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	savedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("050"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

const (
	FLAG_STRING       = "string"
	FLAG_STRING_ARRAY = "stringArray"
	FLAG_BOOL         = "bool"
	FLAG_INT          = "int"
)

type configure struct {
	back *main

	cmd   *cobra.Command
	index int
	saved int

	inputs []textinput.Model
}

func Configure(back *main, cmd *cobra.Command) (v configure, c tea.Cmd) {
	v = configure{
		back: back,
		cmd:  cmd,
	}
	v.inputs = make([]textinput.Model, len(v.GetFlags()))
	v.initFlags()
	v.Reset()
	c = v.Focus(0)
	return
}

func (v configure) initFlags() {
	for i, flag := range v.GetFlags() {
		v.inputs[i] = textinput.New()
		v.inputs[i].CursorStyle = cursorStyle
		v.inputs[i].Prompt = style.Pad(fmt.Sprintf("%v:", flag.Name))
	}
}

func (v configure) Init() tea.Cmd {
	return textinput.Blink
}

func removeBrackets(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			s, "[", "",
		), "]", "",
	)
}

func (v configure) Focus(index int) (cmd tea.Cmd) {
	for i := range v.inputs {
		if i == index {
			cmd = v.inputs[i].Focus()
			v.inputs[i].PromptStyle = focusedStyle
			v.inputs[i].TextStyle = focusedStyle
		} else {
			v.inputs[i].Blur()
			v.inputs[i].PromptStyle = noStyle
			v.inputs[i].TextStyle = noStyle
		}
	}
	return
}

func (v configure) Save(index int) (m tea.Model, cmd tea.Cmd) {
	input := v.inputs[index]
	flag := v.GetFlags()[index]
	flag.Value.Set(input.Value())

	v.inputs[index].TextStyle = savedStyle
	v.inputs[index].PromptStyle = savedStyle
	v.saved = index

	m = v
	return
}

func (v configure) Run() (m tea.Model, cmd tea.Cmd) {

	return
}

func (v configure) Reset() (m tea.Model, cmd tea.Cmd) {
	for i, flag := range v.GetFlags() {
		v.inputs[i].SetValue(removeBrackets(flag.Value.String()))
	}
	return
}

func (v configure) GetFlags() []*pflag.Flag {
	return util.AllFlags(v.cmd)
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
			bounds := len(v.GetFlags())
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
			v.Focus(v.index)
			v.Reset()
		case keys.ENTER:
			if v.index == 0 {
				m, c = v.Run()
			} else {
				m, c = v.Save(v.index)
				return
			}

		default:
			v.Focus(v.index)
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

func (v configure) Body() string {
	var inputs []string
	for i := range v.inputs {
		inputs = append(inputs, v.inputs[i].View())
	}
	return indent.String(strings.Join(inputs, "\n"), 4)
}

func (v configure) View() string {
	var sections []string
	sections = append(sections, lit.MainHeader)
	sections = append(sections, v.Body())
	sections = append(sections, help.SelectionHelp)
	return strings.Join(sections, "\n\n")
}
