package nlx_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/skuid/tides/pkg/logging"
	"github.com/skuid/tides/pkg/nlx"
	"github.com/skuid/tides/pkg/util"
)

type NlxDataService struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        string    `json:"port"`
	Version     string    `json:"version"`
	IsActive    bool      `json:"is_active"`
	CreatedByID string    `json:"created_by_id"`
	UpdatedByID string    `json:"updated_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EnvSpecificConfig struct {
	ID              string `json:"id"`
	OrganizationID  string `json:"organization_id"`
	Name            string `json:"name"`
	Value           string `json:"value"`
	IsSecret        bool   `json:"is_secret"`
	IsManaged       bool   `json:"is_managed"`
	DataServiceID   string `json:"data_service_id"`
	DataServiceName string
	CreatedByID     string    `json:"created_by_id"`
	UpdatedByID     string    `json:"updated_by_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

const fakeDefaultDataServiceId = "153b1f3e-e35a-4bb8-90a4-abbcc95fe15c"

func IsDefaultDs(ds string) bool {
	defaultDS := map[string]interface{}{
		"":                       nil,
		fakeDefaultDataServiceId: nil,
		"default":                nil,
	}
	_, ok := defaultDS[ds]
	return ok
}

func GetDataServices(auth *nlx.Authorization) (results map[string]NlxDataService, err error) {
	headers := nlx.GenerateHeaders(auth.Host, auth.AccessToken)

	var dataservices []NlxDataService
	if dataservices, err = nlx.FastJsonBodyRequest[[]NlxDataService](
		fmt.Sprintf("%s/api/v1/objects/dataservice", auth.Host),
		http.MethodGet,
		nil,
		headers,
	); err != nil {
		return
	}

	results = make(map[string]NlxDataService, len(dataservices))
	for _, ds := range dataservices {
		results[ds.ID] = ds
	}

	return
}

func TestVariables(t *testing.T) {
	util.SkipIntegrationTest(t)
	logging.SetVerbose(true)

	auth, err := nlx.Authorize(authHost, authUser, authPass)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	services, err := GetDataServices(auth)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	fmt.Println(services)
}
