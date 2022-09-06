package style

import (
	"fmt"
	"strings"
)

func Pad(text string) string {
	return fmt.Sprintf("%-20s", text)
}

func RemoveBrackets(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			s, "[", "",
		), "]", "",
	)
}
