package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/skuid/skuid-cli/platform"
	"github.com/skuid/skuid-cli/text"
	"github.com/spf13/cobra"
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

func getDataServices(api *platform.RestApi) (map[string]DataService, error) {
	//searchterm := html.EscapeString(variabledataservice)
	//dspath := "/objects/dataservice?search="+searchterm+"&limit=1",
	dspath := "/objects/dataservice"
	dsStream, err := api.Connection.MakeRequest(
		http.MethodGet,
		dspath,
		nil,
		"application/json",
	)
	if err != nil {
		return nil, errors.New("Error requesting Data Service list from platform.")
	}
	var dataservices []DataService
	if err = json.NewDecoder(*dsStream).Decode(&dataservices); err != nil {
		return nil, errors.New("Could not parse Data Service from platform response.")
	}
	dsmap := make(map[string]DataService, len(dataservices))
	for _, ds := range dataservices {
		dsmap[ds.ID] = ds
	}
	return dsmap, nil
}

var getvarCmd = &cobra.Command{
	Use:   "getvariables",
	Short: "Get a list of Skuid environment variables.",
	Long:  "Get a list of existing Skuid environment variables.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(text.RunCommand("Get Variables"))

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			metadataServiceProxy,
			dataServiceProxy,
			verbose,
		)

		if err != nil {
			fmt.Println(text.PrettyError("Error logging in to Skuid site", err))
			os.Exit(1)
		}

		escResult, err := getEscs(api)
		if err != nil {
			fmt.Println(text.PrettyError("Error getting variables from Skuid site", err))
			os.Exit(1)
		}
		body := strings.Builder{}
		body.WriteString("Name\tDataService\n")
		for _, esc := range escResult {
			body.WriteString(esc.Name + "\t" + esc.DataServiceName + "\n")
		}
		if verbose {
			successMessage := "Successfully retrieved variables from Skuid Site\n"
			fmt.Println(successMessage + text.Separator() + body.String())
		} else {
			fmt.Println(body.String())
		}
	},
}

func getEscs(api *platform.RestApi) ([]EnvSpecificConfig, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Getting Variables"))
	}

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
	if err := json.NewDecoder(*result).Decode(&escs); err != nil { return nil, err }

	dsmap, err := getDataServices(api)
	if err != nil { return nil, err }
	for i, esc := range escs {
		if esc.DataServiceID != "" {
			if ds, ok := dsmap[esc.DataServiceID]; ok {
				esc.DataServiceName = ds.Name
				escs[i] = esc
			}
		}
	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success getting variable values", escStart))
	}
	return escs, nil
}

// setvarCmd represents the setvariable command
var setvarCmd = &cobra.Command{
	Use:   "setvariable",
	Short: "Set a Skuid environment variable.",
	Long:  "Set a Skuid envorinment variable. These can be used as credentials for a data source without revealing them.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(text.RunCommand("Set Variable"))

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			metadataServiceProxy,
			dataServiceProxy,
			verbose,
		)

		if err != nil {
			fmt.Println(text.PrettyError("Error logging in to Skuid site", err))
			os.Exit(1)
		}

		variableStart := time.Now()
		err = setEsc(api)
		if err != nil {
			fmt.Println(text.PrettyError("Error setting variable in Skuid site", err))
			os.Exit(1)
		}
		successMessage := "Successfully set variable in Skuid Site"
		if verbose {
			fmt.Println(text.SuccessWithTime(successMessage, variableStart))
		} else {
			fmt.Println(successMessage + ".")
		}
	},
}

func setEsc(api *platform.RestApi) error {
	if verbose {
		fmt.Println(text.VerboseSection("Setting Variable"))
	}
	if variablename == "" || variablevalue == "" {
		return errors.New("Variable name and value are required for this command.")
	}

	dataServiceId := ""
	if variabledataservice != "" {
		if _, err := uuid.FromString(variabledataservice); err == nil {
			dataServiceId = variabledataservice
		} else {
			dataservices, err := getDataServices(api)
			if err != nil {
				return err
			}

			found := false
			for _, ds := range dataservices {
				if ds.Name == variabledataservice {
					found = true
					dataServiceId = ds.ID
				}
			}
			if !found {
				return errors.New("Could not find specified Data Service by name.")
			}
		}
	}

	body := "{\n"
	body += "\"name\":\"" + variablename + "\",\n"
	body += "\"value\":\"" + variablevalue + "\",\n"
	if dataServiceId != "" {
		body += "\"data_service_id\":\"" + dataServiceId + "\",\n"
	}
	body += "}\n"

	escStart := time.Now()
	api.Connection.APIVersion = "1"
	_, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/ui/variables",
		strings.NewReader(body),
		"application/json",
	)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Println(text.SuccessWithTime("Success setting variable value", escStart))
	}
	return nil
}

func init() {
	RootCmd.AddCommand(setvarCmd)
	RootCmd.AddCommand(getvarCmd)
}
