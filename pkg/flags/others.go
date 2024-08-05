package flags

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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

	LogLevel = &Flag[logrus.Level]{
		Name:    "log-level",
		Usage:   "The `level` of logging: {trace|debug|info|warn|error|fatal|panic}",
		Default: logrus.InfoLevel,
		Global:  true,
		Parse: func(val string) (logrus.Level, error) {
			// simple wrapper to customize error message, otherwise it will contain the word logrus
			if l, err := logrus.ParseLevel(val); err != nil {
				return l, fmt.Errorf("not a valid log level: %q", val)
			} else {
				return l, nil
			}
		},
	}
)
