package util

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/skuid/skuid-cli/pkg/logging"
)

// hopefully this doesn't mess with anything lol
// might want to use foreign keys or something. if a
// page has the word "%NAME%" in it, this might interfere.
const (
	Placeholder = `%NAME%`
)

// InsertNamePlaceholders replaces the thing that we want to go first (name) with
// a placeholder that will show up alphabetically first (%NAME%). when golang
// marshals maps into json, it sorts the keys. so we're going to piggyback off
// that logic and then do a replacement on the string
func InsertNamePlaceholders(in map[string]interface{}) map[string]interface{} {
	for k, v := range in {
		if k == "name" {
			delete(in, k)
			k = fmt.Sprintf("%v", Placeholder)
			in[k] = v
		}
		if m, anotherMap := v.(map[string]interface{}); anotherMap {
			in[k] = InsertNamePlaceholders(m)
		}
	}
	return in
}

// we want to replace all of the placeholders with the actual
// name string and return the bytes
func ReplaceNamePlaceholders(in []byte) []byte {
	return []byte(
		strings.ReplaceAll(
			string(in),
			fmt.Sprintf(`"%v"`, Placeholder),
			`"name"`),
	)
}

// this takes marshalled json in bytes and resorts it recursively the way
// we want it, with "name" field first.
func ReSortJson(data []byte, indent bool) (sorted []byte, err error) {
	message := "Re-sorting JSON"
	fields := logging.Fields{
		"dataLen": len(data),
		"indent":  indent,
	}
	logger := logging.WithTraceTracking("util.ReSortJsonIndent", message, fields).StartTracking()
	defer func() { logger.FinishTracking(err) }()

	var unsorted map[string]interface{}
	if err := json.Unmarshal(data, &unsorted); err != nil {
		logger.WithError(err).Debugf("Error unmarshalling into map[string]interface{}: %s", data)
		return nil, fmt.Errorf("unable to re-sort json indent due to parsing error: %w", err)
	}

	// insert the name placeholders
	unsorted = InsertNamePlaceholders(unsorted)

	// marshal, which uses the native ordering
	// (will do alphanumerically)
	if indent {
		sorted, err = json.MarshalIndent(unsorted, "", "\t")
	} else {
		sorted, err = json.Marshal(unsorted)
	}
	if err != nil {
		logger.WithError(err).Debugf("Error marshalling data: %v", unsorted)
		return nil, fmt.Errorf("unable to convert data to JSON bytes: %w", err)
	}

	logger = logger.WithSuccess()
	return ReplaceNamePlaceholders(sorted), nil
}
