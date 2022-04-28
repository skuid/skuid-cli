package pkg_test

import (
	"testing"

	"github.com/gookit/color"
)

func TestRoot(t *testing.T) {
	for _, tc := range []struct {
		description string
	}{
		{
			description: "test!",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			color.Red.Println("running:", tc.description)
		})
	}
}
