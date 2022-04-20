package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/types"
)

func TestUnmarshalMetadata(t *testing.T) {
	for _, tc := range []struct {
		description string
		given       []byte
		expected    types.Metadata
	}{
		{
			description: "original",
			given: []byte(`
				{
					"profiles": ["A","B","C"]					
				}
			`),
			expected: types.Metadata{
				Profiles: []string{"A", "B", "C"},
			},
		},
		{
			description: "new",
			given: []byte(`
				{
					"sitepermissionsets": ["A","B","C"]					
				}
			`),
			expected: types.Metadata{
				Profiles: []string{"A", "B", "C"},
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual := types.Metadata{}
			err := json.Unmarshal(tc.given, &actual)
			assert.Equal(t, err, nil)
			assert.Equal(t, tc.expected, actual)
		})
	}

}
