package flags

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

var (
	Since = &Flag[*time.Time, *time.Time]{
		Name:          "since",
		Shorthand:     "s",
		Usage:         "Only retrieve objects modified since the specified `time`, e.g., \"yyyy-MM-dd HH:mm AM\", \"HH:mm AM\", \"1y2w3d8h30m\"",
		LegacyEnvVars: []string{constants.EnvSkuidRetrieveSince},
		Parse:         ParseSince,
	}

	LogLevel = &Flag[logrus.Level, logrus.Level]{
		Name:    "log-level",
		Usage:   "The `level` of logging: {trace|debug|info|warn|error|fatal|panic}",
		Default: logrus.InfoLevel,
		Global:  true,
		Parse:   ParseLogLevel,
	}

	// Skuid Review Required - Not sure what the best name for this flag is. The docs (https://docs.skuid.com/latest/en/skuid/cli/)
	// indicate that the CLI is used for deploying Skuid Objects (data and metadata) but it is unclear what "data" is (see
	// https://github.com/skuid/skuid-cli/issues/201).  Also, is is unclear from the docs if there is any official nomenclature for
	// what a specific "Item" of  "Metadata Type" is called. Depending on what the meaning of "data" is and/or if there is official
	// nomenclature, a more appropriate name for this flag may be "Objects".  For now, using "Entities" as I currently do not see that
	// the CLI supports any type of "data" nor was I able to find any official noemclature in the docs regarding what Skuid calls an
	// "Item" of a "Metadata Type".
	// TODO: Rename this flag if needed based on answers to the above.
	Entities = &Flag[[]metadata.MetadataEntity, metadata.MetadataEntity]{
		Name:  "entities",
		Usage: "Entity name(s), separated by a comma",
		Parse: ParseMetadataEntity,
	}
)
