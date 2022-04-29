package pkg

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

func SkuidNlxLogin(cmd *cobra.Command) (api *NlxApi, err error) {
	var host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy string

	f := cmd.Flags()

	if host, err = f.GetString(flags.Host.Name); err != nil {
		return
	}

	if username, err = f.GetString(flags.Username.Name); err != nil {
		return
	}

	if password, err = f.GetString(flags.Password.Name); err != nil {
		return
	}

	if apiVersion, err = f.GetString(flags.ApiVersion.Name); err != nil {
		return
	}

	if metadataServiceProxy, err = f.GetString(flags.MetadataServiceProxy.Name); err != nil {
		return
	}

	if dataServiceProxy, err = f.GetString(flags.DataServiceProxy.Name); err != nil {
		return
	}

	return SkuidNlxLogin2(host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy)
}

func SkuidNlxLogin2(host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy string) (api *NlxApi, err error) {

	if apiVersion == "" {
		apiVersion = "2"
	}

	loginStart := time.Now()

	for _, msg := range [][]string{
		{"Skuid NLX User:", username},
		{"Skuid NLX Host:", host},
		{"Skuid NLX API Version:", apiVersion},
	} {
		logging.VerboseF("%-20s\t%s\n", msg[0], color.Green.Sprint(msg[1]))
	}
	logging.VerboseSeparator()

	conn := NlxConnection{
		Host:       host,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
	}

	if metadataServiceProxy != "" {
		logging.VerboseF("Using Metadata Service Proxy: %v", color.Green.Sprint(metadataServiceProxy))
		if conn.MetadataServiceProxy, err = url.Parse(metadataServiceProxy); err != nil {
			return
		}
	}

	if dataServiceProxy != "" {
		logging.VerboseF("Using Data Service Proxy: %v", color.Green.Sprint(dataServiceProxy))
		if conn.DataServiceProxy, err = url.Parse(dataServiceProxy); err != nil {
			return
		}
	}

	err = conn.Refresh()

	if err != nil {
		return
	}

	err = conn.GetJWT()

	if err != nil {
		return
	}

	api = &NlxApi{
		Connection: &conn,
	}

	logging.VerboseSuccess("Login Success", loginStart)

	return
}

func getClientForProxyURL(url *url.URL) *http.Client {
	if url != nil {
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(url),
			},
		}
	}
	return http.DefaultClient
}
