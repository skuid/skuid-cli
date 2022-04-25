package main

// StringSliceContainsKey returns true if a string is contained in a slice
func StringSliceContainsKey(strings []string, key string) bool {
	for _, item := range strings {
		if item == key {
			return true
		}
	}
	return false
}

func StringSliceContainsAnyKey(strings []string, keys []string) bool {
	for _, item := range strings {
		for _, key := range keys {
			if item == key {
				return true
			}
		}
	}
	return false
}
