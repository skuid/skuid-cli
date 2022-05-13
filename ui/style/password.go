package style

import (
	"strings"

	"github.com/spf13/pflag"
)

var (
	PasswordPlaceholder = strings.Repeat("•", 8)
)

func IsPassword(flag *pflag.Flag) bool {
	return strings.ToLower(flag.Name) == "password"
}
