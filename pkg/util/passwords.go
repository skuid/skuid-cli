package util

import (
	"strings"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
)

// RemovePasswordBytes is a pass-through to RemovePassword.
func RemovePasswordBytes(bytes []byte) []byte {
	bytes = []byte(RemovePassword(string(bytes)))
	return bytes
}

// RemovePassword is the reason I wrote the
// "GetValue()" function at all. We want to be
// able to safely remove the password from the
// strings we output to logs, so we should do this.
func RemovePassword(str string) string {
	str = strings.ReplaceAll(
		str,
		flags.Password.GetValue(),
		constants.PasswordPlaceholder,
	)
	return str
}
