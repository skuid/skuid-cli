package style

import (
	"fmt"

	"github.com/spf13/pflag"
)

func Flag(flag *pflag.Flag, selected bool) string {
	var selectString string

	pad := func(flag string) string {
		return fmt.Sprintf("%-20s", flag)
	}

	if selected {
		selectString = Tides(fmt.Sprintf("[x] %v %v", pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	} else {
		selectString = Subtle(fmt.Sprintf("[ ] %v %v", pad(flag.Name), flag.Value.String()))
		// selectHelp = style.Subtle(indent.String(fmt.Sprintf("%v (%v)", flag.Usage, flag.NoOptDefVal), 3))
	}

	return fmt.Sprintf("%v", selectString)
}
