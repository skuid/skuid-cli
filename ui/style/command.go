package style

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/indent"
	"github.com/skuid/domain/constants"
	"github.com/skuid/domain/util"
	"github.com/spf13/cobra"
)

func CommandText(cmd *cobra.Command) string {
	var text []string
	text = append(text, constants.PROJECT_NAME)
	text = append(text, indent.String(cmd.Name(), 1))
	for _, flag := range util.AllFlags(cmd) {
		s := flag.Value.String()
		if s != "" && s != "false" && s != "[]" {
			if IsPassword(flag) {
				s = constants.PasswordPlaceholder
			}
			text = append(text, indent.String(
				fmt.Sprintf(
					"%v %v",
					fmt.Sprintf("--%v", flag.Name),
					RemoveBrackets(s),
				), 3),
			)
		}
	}
	return strings.Join(text, "\n")
}
