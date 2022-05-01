package nlx_test

import (
	"testing"

	"github.com/gookit/color"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

type DeploymentPlan struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Url      string   `json:"url"`
	Type     string   `json:"type"`
	Warnings []string `json:"warnings"`
	Metadata struct {
		Apps               []string `json:"apps"`
		AuthProviders      []string `json:"authproviders"`
		ComponentPacks     []string `json:"componentpacks"`
		DataServices       []string `json:"dataservices"`
		DataSources        []string `json:"datasources"`
		DesignSystems      []string `json:"designsystems"`
		Variables          []string `json:"variables"`
		Files              []string `json:"files"`
		Pages              []string `json:"pages"`
		PermissionSets     []string `json:"permissionsets"`
		SitePermissionSets []string `json:"sitepermissionsets"`
		Site               []string `json:"site"`
		Themes             []string `json:"themes"`
	} `json:"metadata"`
}

func GetDeployPlans(host, authToken string) (plans map[string]DeploymentPlan, err error) {

	return
}

func TestFasthttpMethods(t *testing.T) {
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
