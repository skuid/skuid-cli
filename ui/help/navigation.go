package help

import (
	"fmt"
	"strings"

	"github.com/skuid/tides/ui/style"
)

var (
	SelectionControls = map[string][]string{
		"select":    {"up", "down"}, // "tab", "shift+tab"},
		"quit/back": {"esc", "ctrl+c"},
	}

	SelectionHelp = helpText(SelectionControls)

	EditingControls = map[string][]string{
		"select":      {"up", "down"}, // "tab", "shift+tab"},
		"run/save":    {"enter"},      // "ctrl+s"},
		"revert/back": {"esc"},
		"quit":        {"ctrl+c"},
	}

	EditingHelp = helpText(EditingControls)
)

func helpText(controls map[string][]string) string {
	var helps []string

	for k, v := range controls {
		var options []string
		for _, option := range v {
			options = append(options, fmt.Sprintf("[%v]", option))
		}

		helps = append(helps,
			style.Subtle(fmt.Sprintf("%v: %v", strings.Join(options, "/"), k)),
		)
	}

	return strings.Join(helps, style.Subtle(style.Dot))
}
