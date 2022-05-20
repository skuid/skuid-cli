package pkg_test

import (
	"os"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	"github.com/skuid/tides/pkg"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/util"
)

var (
	authHost = os.Getenv(constants.ENV_SKUID_HOST)
	authUser = os.Getenv(constants.ENV_SKUID_USERNAME)
	authPass = os.Getenv(constants.ENV_SKUID_PASSWORD)
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

		var auth pkg.Authorization
		if auth, err := pkg.Authorize(authHost, authUser, authPass); err != nil {
			t.FailNow()
		} else if err := auth.Refresh(); err != nil {
			t.FailNow()
		}

		assert.NotEqual(t, auth.AccessToken, accessToken)
		assert.NotEqual(t, auth.AuthorizationToken, authorizationToken)
	}
}
