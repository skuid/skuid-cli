package cmd

import (
	"fmt"
	"os"
	"strconv"

	"net/url"

	"encoding/json"

	"github.com/skuid/skuid/force"
	"github.com/skuid/skuid/types"
	"github.com/spf13/cobra"
)

var pagePackModule string
var outputFile string

// page-packCmd represents the page-pack command
var pagePackCmd = &cobra.Command{
	Use:   "page-pack",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//login to the Force API
		api, err := force.Login(
			appClientID,
			appClientSecret,
			sfdcHost,
			sfdcUsername,
			sfdcPassword,
			apiVersion,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		query := url.Values{
			"module": []string{pagePackModule},
			"as":     []string{"pagePack"},
		}

		result, err := api.Connection.Get("/skuid/api/v1/pages", query)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		unquoted, _ := strconv.Unquote(string(result))

		pagePack := &types.PagePackResponse{}

		err = json.Unmarshal([]byte(unquoted), pagePack)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		err = pagePack.WritePagePack(outputFile, pagePackModule)
	},
}

func init() {
	RootCmd.AddCommand(pagePackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pagePackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pagePackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pagePackCmd.Flags().StringVarP(&pagePackModule, "module", "m", "", "Module to pull.")
	pagePackCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Filename of output file")
}
