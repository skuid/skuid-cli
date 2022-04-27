package main_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	tides "github.com/skuid/tides"
)

var (
	sortedString = `{
		"name": "name",
		"a": "a",
		"c": "c",
		"d": "d",
		"x": "x",
		"z": "z"
	}`

	unsortedString = `{
		"a": "a",
		"z": "z",
		"c": "c",
		"name": "name",
		"d": "d",
		"x":"x"
	}`

	sortedNested = `{
		"name": "name",
		"a": "a",
		"c": "c",
		"d": "d",
		"nested": {
			"name":"name",
			"a":"a"
		},
		"x":"x",
		"z": "z"
	}`

	unsortedNested = `{
		"a": "a",
		"z": "z",
		"c": "c",
		"name": "name",
		"nested": {
			"a":"a",
			"name":"name"
		},
		"d": "d",
		"x":"x"
	}`
)

func TestSortJson(t *testing.T) {
	for _, tc := range []struct {
		description string
		given       string
		expected    string
	}{
		{
			description: "sorted",
			given:       sortedString,
			expected:    sortedString,
		},
		{
			description: "unsorted",
			given:       unsortedString,
			expected:    sortedString,
		},
		{
			description: "sorted nested",
			given:       sortedNested,
			expected:    sortedNested,
		},
		{
			description: "unsorted nested",
			given:       unsortedNested,
			expected:    sortedNested,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual, _ := tides.ReSortJson([]byte(tc.given))
			expectedTrimmed := regexp.MustCompile(`\s`).ReplaceAllString(tc.expected, "")
			assert.Equal(t, expectedTrimmed, string(actual))
		})
	}
}
