package pkg_test

import (
	"os"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/util"
)

var (
	authOptions = &pkg.AuthorizeOptions{
		os.Getenv(cmdutil.EnvVarName(flags.Host.Name)),
		os.Getenv(cmdutil.EnvVarName(flags.Username.Name)),
		flags.NewValueWithRedact(os.Getenv(cmdutil.EnvVarName(flags.Password.Name)), new(string), nil, nil),
	}
)

// if you have to run it by itself, add some environment variables
// otherwise this crap is going down lol
func TestAuthorizationMethods(t *testing.T) {
	util.SkipIntegrationTest(t)
	if err := util.LoadTestEnvironment(); err != nil {
		t.Fatal(err)
	}

	if accessToken, err := pkg.GetAccessToken(authOptions); err != nil {
		color.Red.Println(err)
		t.FailNow()
	} else if authorizationToken, err := pkg.GetAuthorizationToken(
		authOptions.Host, accessToken,
	); err != nil {
		color.Red.Println(err)
		t.FailNow()
	} else {

		auth, err := pkg.Authorize(authOptions)
		if err != nil {
			t.FailNow()
		}

		assert.NotEqual(t, auth.AccessToken, accessToken)
		assert.NotEqual(t, auth.AuthorizationToken, authorizationToken)
	}
}
