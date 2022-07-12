package util_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/util"
)

var (
	testPassword = "testPassword"
)

func TestRemovePassword(t *testing.T) {
	for _, tc := range []struct {
		description    string
		givenPassword  string
		givenString    string
		expectedString string
	}{
		{
			description:    "no password",
			givenPassword:  "",
			givenString:    "max-fried",
			expectedString: "max-fried",
		},
		{
			description:    "password",
			givenPassword:  "max-fried",
			givenString:    "max-fried",
			expectedString: constants.PasswordPlaceholder,
		},
		{
			description:    "password x2",
			givenPassword:  "max-fried",
			givenString:    "max-fried, max-fried",
			expectedString: fmt.Sprintf("%v, %v", constants.PasswordPlaceholder, constants.PasswordPlaceholder),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			flags.Password.Set(&tc.givenPassword)
			result := util.RemovePassword(tc.givenString)
			assert.Equal(t, tc.expectedString, result)
		})
	}
}
