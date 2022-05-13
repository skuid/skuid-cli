package style

import (
	"fmt"

	"github.com/muesli/reflow/indent"
	"github.com/spf13/cobra"
)

func CommandString(cmd *cobra.Command, selected bool) string {
	name := cmd.Name()
	description := indent.String(cmd.Short, 4)
	if selected {
		return Skuid(fmt.Sprintf("[x] %s\n%s", name, Tides(description)))
	}
	return fmt.Sprintf("[ ] %s\n%s", name, Subtle(description))
}
