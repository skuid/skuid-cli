package util

import "slices"

// StringSliceContainsKey returns true if a string is contained in a slice
func StringSliceContainsKey(strings []string, key string) bool {
	return slices.Contains(strings, key)
}

// same thing as above except we early exit at any given point going through the
// strings and keys
func StringSliceContainsAnyKey(strings []string, keys []string) bool {
	return slices.ContainsFunc(strings, func(item string) bool {
		return slices.Contains(keys, item)
	})
}
