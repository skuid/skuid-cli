package pkg

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/skuid/tides/pkg/logging"
)

type DataService struct {
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

func isDefaultDs(ds string) bool {
	defaultDS := map[string]interface{}{
		"":                       nil,
		fakeDefaultDataServiceId: nil,
		"default":                nil,
	}
	_, ok := defaultDS[ds]
	return ok
}

func getDataServices(api *NlxApi) (map[string]DataService, error) {
	dspath := "/objects/dataservice"
	dsStream, err := api.Connection.MakeAccessTokenRequest(
		http.MethodGet,
		dspath,
		nil,
		"application/json",
	)
	if err != nil {
		return nil, errors.New("Error requesting Data Service list.")
	}
	var dataservices []DataService
	if err = json.Unmarshal(dsStream, &dataservices); err != nil {
		return nil, errors.New("Could not parse Data Service list response.")
	}
	dsmap := make(map[string]DataService, len(dataservices))
	for _, ds := range dataservices {
		dsmap[ds.ID] = ds
	}
	return dsmap, nil
}

func findDataServiceId(api *NlxApi, variableDataService, name string) (string, error) {
	if isDefaultDs(variableDataService) {
		return fakeDefaultDataServiceId, nil
	}
	if _, err := uuid.FromString(variableDataService); err == nil {
		return variableDataService, nil
	}
	// Match the name with an existing Private Data Service
	dataservices, err := getDataServices(api)
	if err != nil {
		return "", err
	}
	for _, ds := range dataservices {
		if ds.Name == name {
			return ds.ID, nil
		}
	}
	return "", errors.New("Could not find specified Data Service by name.")
}

func GetEnvironmentSpecificConfigurations(api *NlxApi, mapDsName bool) ([]EnvSpecificConfig, error) {

	logging.VerboseSection("Getting Variables")

	escStart := time.Now()
	api.Connection.APIVersion = "1"
	result, err := api.Connection.MakeAccessTokenRequest(
		http.MethodGet,
		"/ui/variables",
		nil,
		"application/json",
	)
	if err != nil {
		return nil, err
	}

	var escs []EnvSpecificConfig
	if err := json.Unmarshal(result, &escs); err != nil {
		return nil, err
	}

	if mapDsName {
		dsmap, err := getDataServices(api)
		if err != nil {
			return nil, err
		}
		for i, esc := range escs {
			if esc.DataServiceID != "" {
				if esc.DataServiceID == fakeDefaultDataServiceId {
					esc.DataServiceName = "default"
					escs[i] = esc
				} else if ds, ok := dsmap[esc.DataServiceID]; ok {
					esc.DataServiceName = ds.Name
					escs[i] = esc
				}
			}
		}
	}

	logging.VerboseSuccess("Success getting variable values", escStart)

	return escs, nil
}

func SetEnvironmentSpecificConfiguration(api *NlxApi, variableName, variableValue, variableDataService string) error {

	logging.VerboseSection("Setting Variable")

	if variableName == "" {
		return errors.New("Variable name is required for this command.")
	}
	if variableValue == "" {
		logging.Println("Enter value:")
		valbytes, err := terminal.ReadPassword(0)
		if err != nil {
			return errors.New("Error reading value from prompt.")
		}
		variableValue = string(valbytes)
	}
	api.Connection.APIVersion = "1"

	dataServiceId, err := findDataServiceId(api, variableDataService, variableName)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	body["name"] = variableName
	body["value"] = variableValue
	if dataServiceId != "" {
		body["data_service_id"] = dataServiceId
	}

	verb := http.MethodPost
	path := "/ui/variables"
	existingEscs, err := GetEnvironmentSpecificConfigurations(api, false)
	if err != nil {
		return err
	}
	for _, existing := range existingEscs {
		if existing.Name == variableName && existing.DataServiceID == dataServiceId {
			verb = http.MethodPut
			path += "/" + existing.ID
			row := body
			row["id"] = existing.ID
			body = map[string]interface{}{
				"changes": map[string]interface{}{
					"value": variableValue,
				},
				"row": row,
				"originals": map[string]interface{}{
					existing.ID: map[string]interface{}{
						"value": "*****",
					},
				},
			}
			break
		}
	}

	bodybytes, err := json.Marshal(body)
	payload := string(bodybytes) + "\n"
	if err != nil {
		return err
	}
	escStart := time.Now()
	_, err = api.Connection.MakeAccessTokenRequest(
		verb,
		path,
		strings.NewReader(payload),
		"application/json",
	)
	if err != nil {
		return err
	}

	logging.VerboseSuccess("Success setting variable value", escStart)

	return nil
}

func RemoveEnvironmentSpecificConfigurations(api *NlxApi, variableName, variableDataService string) error {

	logging.VerboseSection("Deleting Variable")

	if variableName == "" {
		return errors.New("Variable name is required for this command.")
	}

	api.Connection.APIVersion = "1"
	dataServiceId, err := findDataServiceId(api, variableDataService, variableName)
	if err != nil {
		return err
	}

	// Find ID of ESC to delete
	escID := ""
	existingEscs, err := GetEnvironmentSpecificConfigurations(api, false)
	if err != nil {
		return err
	}
	for _, existing := range existingEscs {
		if existing.Name == variableName && existing.DataServiceID == dataServiceId {
			escID = existing.ID
			break
		}
	}
	if escID == "" {
		return errors.New("Could not find specified variable to delete.")
	}
	path := "/ui/variables/" + escID

	body := map[string]interface{}{
		"data_service_id": dataServiceId,
	}

	bodybytes, err := json.Marshal(body)
	payload := string(bodybytes) + "\n"
	if err != nil {
		return err
	}
	escStart := time.Now()
	_, err = api.Connection.MakeAccessTokenRequest(
		http.MethodDelete,
		path,
		strings.NewReader(payload),
		"application/json",
	)
	if err != nil {
		return err
	}

	logging.VerboseSuccess("Success deleting variable", escStart)

	return nil
}
