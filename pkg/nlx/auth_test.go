package nlx_test

import (
	"testing"

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

func TestAuthorizationMethods(t *testing.T) {
	util.SkipIntegrationTest(t)
	host := "https://jredhoop-subdomain.pliny.webserver:3000"
	logging.SetVerbose(true)

	if at, err := nlx.GetAccessToken(
		host, "jredhoop", "SkuidLocalDevelopment",
	); err != nil {
		color.Red.Println(err)
		t.FailNow()
	} else if _, err := nlx.GetAuthorizationToken(
		host, at,
	); err != nil {
		color.Red.Println(err)
		t.FailNow()
	}
}
