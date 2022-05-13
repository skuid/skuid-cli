package style

import (
	"fmt"

	"github.com/muesli/reflow/indent"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func CommandString(cmd *cobra.Command, selected bool) string {
	commandName := cmd.Name()
	description := indent.String(cmd.Short, 4)

	var text string
	if selected {
		text = fmt.Sprintf(" %s\n%s", commandName, Tides(description))
	} else {
		text = fmt.Sprintf(" %s\n%s", commandName, Subtle(description))
	}

	return Checkbox(text, selected, true)
}

func Flag(flag *pflag.Flag, selected bool) string {

	var text string
	if selected {
		text = Tides(fmt.Sprintf(" %v %v", Pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	} else {
		text = Subtle(fmt.Sprintf(" %v %v", Pad(flag.Name), flag.Value.String()))
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
		ret = Skuid(ret)
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
		ret = Tides(text)
	} else {
		ret = text
	}
	return
}
