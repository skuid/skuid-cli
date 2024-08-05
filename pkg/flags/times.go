package flags

import (
	"time"

	"github.com/skuid/skuid-cli/pkg/constants"
)

var (
	Since = &Flag[*time.Time]{
		Name:          "since",
		Shorthand:     "s",
		Usage:         "Only retrieve objects modified since the specified `time`, e.g., \"yyyy-MM-dd HH:mm AM\", \"HH:mm AM\", \"1y2w3d8h30m\"",
		LegacyEnvVars: []string{constants.ENV_SKUID_RETRIEVE_SINCE},
		Parse:         ParseSince,
	}
)
