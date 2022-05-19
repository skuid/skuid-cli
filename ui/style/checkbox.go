package style

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func CommandString(cmd *cobra.Command, selected bool) string {
	commandName := cmd.Name()
	description := cmd.Short

	var text string
	if selected {
		text = fmt.Sprintf(" %s\n%s", commandName, StyleTides.Render(description))
	} else {
		text = fmt.Sprintf(" %s\n%s", commandName, StyleBlur.Render(description))
	}

	return Checkbox(text, selected, true)
}

func Flag(flag *pflag.Flag, selected bool) string {

	var text string
	if selected {
		text = StyleTides.Render(fmt.Sprintf(" %v %v", Pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	} else {
		text = StyleSubtle.Render(fmt.Sprintf(" %v %v", Pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	}

	return Checkbox(text, selected, true)
}

var (
	selector = "•"
)

func Checkbox(text string, selected bool, left bool) (ret string) {
	if selected {
		if left {
			ret = fmt.Sprintf("[%s]%s", selector, text)
		} else {
			ret = fmt.Sprintf("%s[%s]", text, selector)
		}
		ret = StyleTides.Render(ret)
	} else {
		if left {
			ret = fmt.Sprintf("[ ]%s", text)
		} else {
			ret = fmt.Sprintf("%s[ ]", text)
		}
	}

	return
}

func HighlightIf(text string, selected bool) (ret string) {
	if selected {
		ret = StyleSkuid.Render(text)
	} else {
		ret = text
	}
	return
}
