package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/skuid/tides/pkg/logging"
)

// hopefully this doesn't mess with anything lol
// might want to use foreign keys or something. if a
// page has the word "%NAME%" in it, this might interfere.
const (
	PLACEHOLDER = `%NAME%`
)

// InsertNamePlaceholders replaces the thing that we want to go first (name) with
// a placeholder that will show up alphabetically first (%NAME%). when golang
// marshals maps into json, it sorts the keys. so we're going to piggyback off
// that logic and then do a replacement on the string
func InsertNamePlaceholders(in map[string]interface{}) map[string]interface{} {
	for k, v := range in {
		if k == "name" {
			delete(in, k)
			in[fmt.Sprintf("%v", PLACEHOLDER)] = v
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
			fmt.Sprintf(`"%v"`, PLACEHOLDER),
			`"name"`),
	)
}

func ReSortJson(data []byte) (replaced []byte, err error) {
	return ReSortJsonIndent(data, false)
}

// this takes marshalled json in bytes and resorts it recursively the way
// we want it, with "name" field first.
func ReSortJsonIndent(data []byte, indent bool) (replaced []byte, err error) {
	var unsorted map[string]interface{}
	err = json.Unmarshal(data, &unsorted)
	if err != nil {
		logging.Logger.Tracef("Error Unmarshalling into map[string]interface{}: %v", string(data))
		return
	}

	// insert the name placeholders
	InsertNamePlaceholders(unsorted)

	// marshal, which uses the native ordering
	// (will do alphanumerically)
	var sorted []byte
	if indent {
		sorted, err = json.MarshalIndent(unsorted, "", "\t")
	} else {
		sorted, err = json.Marshal(unsorted)
	}
	if err != nil {
		return
	}

	// replace the name placeholders
	replaced = ReplaceNamePlaceholders(sorted)

	return
}
