package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skuid/tides/pkg/util"
)

func TestGetAbs(t *testing.T) {
	wd, _ := os.Getwd()
	for _, tc := range []struct {
		description string
		given       string
		expected    string
	}{
		{
			description: "relative",
			given:       ".relative",
			expected: func() string {
				p, _ := filepath.Abs(filepath.Join(wd, ".relative"))
				return p
			}(),
		},
		{
			description: "absolute",
			given: func() string {
				return filepath.Join(wd, ".absolute")
			}(),
			expected: func() string {
				return filepath.Join(wd, ".absolute")
			}(),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			actual := util.GetAbs(tc.given)
			if actual != tc.expected {
				t.Logf("actual %v not equal %v", actual, tc.expected)
				t.FailNow()
			}
		})
	}
}