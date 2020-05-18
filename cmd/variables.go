package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/skuid/skuid-cli/platform"
	"github.com/skuid/skuid-cli/text"
	"github.com/spf13/cobra"
)

var getvarCmd = &cobra.Command{
	Use:  "getvariables",
	Short: "Get a list of Skuid environment variables.",
	Long: "Get a list of existing Skuid environment variables.",
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
		successMessage := "Successfully retrieved variables from Skuid Site"
		if verbose {
			fmt.Println(successMessage + text.Separator() + escResult)
		} else {
			fmt.Println(successMessage + escResult)
		}
	},
}

func getEscs(api *platform.RestApi) (string, error) {
	if verbose {
		fmt.Println(text.VerboseSection("Getting Variables"))
	}

	escStart := time.Now()
	result, err := api.Connection.MakeRequest(
		http.MethodGet,
		"/api/v1/ui/variables",
		nil,
		"application/json",
	)
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Println(text.SuccessWithTime("Success getting variable values", escStart))
	}
	defer (*result).Close()
	body, err := ioutil.ReadAll(*result)
	if err != nil {
		return "", err
	}
	return string(body), nil
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
	body := "{\n"
	body += "\"name\":\"" + variablename + "\",\n"
	body += "\"value\":\"" + variablevalue + "\",\n"
	body += "}\n"
	// TODO: allow specifying the data service by name (fetch its id from pliny) and post with that as "data_service_id"
	escStart := time.Now()
	_, err := api.Connection.MakeRequest(
		http.MethodPost,
		"/api/v1/ui/variables",
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
