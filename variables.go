package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
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

func getDataServices(api *PlatformRestApi) (map[string]DataService, error) {
	dspath := "/objects/dataservice"
	dsStream, err := api.Connection.MakeRequest(
		http.MethodGet,
		dspath,
		nil,
		"application/json",
	)
	if err != nil {
		return nil, errors.New("Error requesting Data Service list.")
	}
	var dataservices []DataService
	if err = json.NewDecoder(*dsStream).Decode(&dataservices); err != nil {
		return nil, errors.New("Could not parse Data Service list response.")
	}
	dsmap := make(map[string]DataService, len(dataservices))
	for _, ds := range dataservices {
		dsmap[ds.ID] = ds
	}
	return dsmap, nil
}

func findDataServiceId(api *PlatformRestApi, name string) (string, error) {
	if isDefaultDs(ArgVariableDataService) {
		return fakeDefaultDataServiceId, nil
	}
	if _, err := uuid.FromString(ArgVariableDataService); err == nil {
		return ArgVariableDataService, nil
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

var getvarCmd = &cobra.Command{
	Use:   "variables",
	Short: "Get a list of Skuid site environment variables.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		VerboseCommand("Get Variables")

		api, err := PlatformLogin(
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
			ArgMetadataServiceProxy,
			ArgDataServiceProxy,
			ArgVerbose,
		)

		if err != nil {
			PrintError("Error logging in to Skuid site", err)
			return
		}

		escResult, err := getEscs(api, true)
		if err != nil {
			PrintError("Error getting variables from Skuid site", err)
			return
		}

		body := tablewriter.NewWriter(os.Stdout)
		body.SetHeader([]string{"Name", "Data Service"})
		for _, esc := range escResult {
			body.Append([]string{esc.Name, esc.DataServiceName})
		}

		VerboseLn("Successfully retrieved variables from Skuid site")

		body.Render()

		return
	},
}

func getEscs(api *PlatformRestApi, mapDsName bool) ([]EnvSpecificConfig, error) {

	VerboseSection("Getting Variables")

	escStart := time.Now()
	api.Connection.APIVersion = "1"
	result, err := api.Connection.MakeRequest(
		http.MethodGet,
		"/ui/variables",
		nil,
		"application/json",
	)
	if err != nil {
		return nil, err
	}

	var escs []EnvSpecificConfig
	if err := json.NewDecoder(*result).Decode(&escs); err != nil {
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

	VerboseSuccess("Success getting variable values", escStart)

	return escs, nil
}

// setvarCmd represents the setvariable command
var setvarCmd = &cobra.Command{
	Use:   "set-variable",
	Short: "Set a Skuid site environment variable",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		VerboseCommand("Set Variable")

		api, err := PlatformLogin(
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
			ArgMetadataServiceProxy,
			ArgDataServiceProxy,
			ArgVerbose,
		)

		if err != nil {
			PrintError("Error logging in to Skuid site", err)
			return
		}

		variableStart := time.Now()
		err = setEsc(api)
		if err != nil {
			PrintError("Error setting variable in Skuid site", err)
			return
		}

		successMessage := "Successfully set variable in Skuid Site"

		VerboseSuccess(successMessage, variableStart)

		return
	},
}

func setEsc(api *PlatformRestApi) error {

	VerboseSection("Setting Variable")

	if ArgVariableName == "" {
		return errors.New("Variable name is required for this command.")
	}
	if ArgVariableValue == "" {
		Println("Enter value:")
		valbytes, err := terminal.ReadPassword(0)
		if err != nil {
			return errors.New("Error reading value from prompt.")
		}
		ArgVariableValue = string(valbytes)
	}
	api.Connection.APIVersion = "1"

	dataServiceId, err := findDataServiceId(api, ArgVariableDataService)
	if err != nil {
		return err
	}
	body := map[string]interface{}{}
	body["name"] = ArgVariableName
	body["value"] = ArgVariableValue
	if dataServiceId != "" {
		body["data_service_id"] = dataServiceId
	}

	verb := http.MethodPost
	path := "/ui/variables"
	existingEscs, err := getEscs(api, false)
	if err != nil {
		return err
	}
	for _, existing := range existingEscs {
		if existing.Name == ArgVariableName && existing.DataServiceID == dataServiceId {
			verb = http.MethodPut
			path += "/" + existing.ID
			row := body
			row["id"] = existing.ID
			body = map[string]interface{}{
				"changes": map[string]interface{}{
					"value": ArgVariableValue,
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
	_, err = api.Connection.MakeRequest(
		verb,
		path,
		strings.NewReader(payload),
		"application/json",
	)
	if err != nil {
		return err
	}

	VerboseSuccess("Success setting variable value", escStart)

	return nil
}

var rmvarCmd = &cobra.Command{
	Use:   "rm-variable",
	Short: "Delete a Skuid site environment variable",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		VerboseCommand("Delete Variable")

		api, err := PlatformLogin(
			ArgHost,
			ArgUsername,
			ArgPassword,
			ArgApiVersion,
			ArgMetadataServiceProxy,
			ArgDataServiceProxy,
			ArgVerbose,
		)

		if err != nil {
			PrintError("Error logging in to Skuid site", err)
			return
		}

		variableStart := time.Now()
		err = rmEsc(api)
		if err != nil {
			PrintError("Error deleting variable in Skuid site", err)
			return
		}

		successMessage := "Successfully deleted variable in Skuid Site"

		VerboseSuccess(successMessage, variableStart)

		return
	},
}

func rmEsc(api *PlatformRestApi) error {

	VerboseSection("Deleting Variable")

	if ArgVariableName == "" {
		return errors.New("Variable name is required for this command.")
	}

	api.Connection.APIVersion = "1"
	dataServiceId, err := findDataServiceId(api, ArgVariableDataService)
	if err != nil {
		return err
	}

	// Find ID of ESC to delete
	escID := ""
	existingEscs, err := getEscs(api, false)
	if err != nil {
		return err
	}
	for _, existing := range existingEscs {
		if existing.Name == ArgVariableName && existing.DataServiceID == dataServiceId {
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
	_, err = api.Connection.MakeRequest(
		http.MethodDelete,
		path,
		strings.NewReader(payload),
		"application/json",
	)
	if err != nil {
		return err
	}

	VerboseSuccess("Success deleting variable", escStart)

	return nil
}

func init() {
	RootCmd.AddCommand(setvarCmd)
	RootCmd.AddCommand(getvarCmd)
	RootCmd.AddCommand(rmvarCmd)
}
