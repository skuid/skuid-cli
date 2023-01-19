package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg/util"
)

func TestStringSliceContainsKey(t *testing.T) {
	for _, tc := range []struct {
		description string
		array       []string
		key         string
		expected    bool
	}{
		{
			description: "found",
			array:       []string{"a", "b", "c"},
			key:         "b",
			expected:    true,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			key:         "d",
			expected:    false,
		},
		{
			description: "not found",
			key:         "b",
			expected:    false,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			key:         "",
			expected:    false,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			result := util.StringSliceContainsKey(tc.array, tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStringSliceContainsAnyKey(t *testing.T) {
	for _, tc := range []struct {
		description string
		array       []string
		keys        []string
		expected    bool
	}{
		{
			description: "found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"b"},
			expected:    true,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			keys:        []string{"d"},
			expected:    false,
		},
		{
			description: "not found",
			keys:        []string{"b"},
			expected:    false,
		},
		{
			description: "not found",
			array:       []string{"a", "b", "c"},
			expected:    false,
		},
		{
			description: "not found",
			expected:    false,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			result := util.StringSliceContainsAnyKey(tc.array, tc.keys)
			assert.Equal(t, tc.expected, result)
		})
	}
}
