package style

import (
	"strings"

	"github.com/spf13/pflag"
)

func IsPassword(flag *pflag.Flag) bool {
	return strings.ToLower(flag.Name) == "password"
}
