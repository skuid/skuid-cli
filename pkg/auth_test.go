package pkg_test

import (
	"os"
	"testing"

	"github.com/gookit/color"
	"github.com/mmatczuk/anyflag"
	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	authHost = os.Getenv(constants.ENV_SKUID_HOST)
	authUser = os.Getenv(constants.ENV_SKUID_USERNAME)
	authPass = anyflag.NewValueWithRedact(flags.RedactedString(os.Getenv(constants.ENV_SKUID_PASSWORD)), new(flags.RedactedString), nil, nil)
)

// if you have to run it by itself, add some environment variables
// otherwise this crap is going down lol
func TestAuthorizationMethods(t *testing.T) {
	util.SkipIntegrationTest(t)
	if err := util.LoadTestEnvironment(); err != nil {
		t.Fatal(err)
	}

	if accessToken, err := pkg.GetAccessToken(
		authHost, authUser, authPass,
	); err != nil {
		color.Red.Println(err)
		t.FailNow()
	} else if authorizationToken, err := pkg.GetAuthorizationToken(
		authHost, accessToken,
	); err != nil {
		color.Red.Println(err)
		t.FailNow()
	} else {

		auth, err := pkg.Authorize(authHost, authUser, authPass)
		if err != nil {
			t.FailNow()
		}

		assert.NotEqual(t, auth.AccessToken, accessToken)
		assert.NotEqual(t, auth.AuthorizationToken, authorizationToken)
	}
}
