package style

import "fmt"

func Pad(text string) string {
	return fmt.Sprintf("%-20s", text)
}
