package flags

import (
	"time"

	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	Since = &Flag[*time.Time]{
		Name:          "since",
		Shorthand:     "s",
		Usage:         "Timestamp or time span specifying only updated records to retrieve. Suggested timestamp format is: \"yyyy-MM-dd HH:mm AM\" or \"HH:mm AM\". Valid time spans look like various combination of \"1y2w3d8h30m\"",
		LegacyEnvVars: []string{constants.ENV_SKUID_RETRIEVE_SINCE},
		Parse:         ParseSince,
	}
)
