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

	if apiVersion == "" {
		apiVersion = "2"
	}

	loginStart := time.Now()

	for _, msg := range [][]string{
		{"Logging in to Skuid Platform as user:", username},
		{"Connecting to host:", host},
		{"API Version:", apiVersion},
	} {
		logging.VerboseF("%-45s\t%s\n", color.Yellow.Sprint(msg[0]), color.Green.Sprint(msg[1]))
	}
	logging.VerboseSeparator()

	conn := NlxConnection{
		Host:       host,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
	}

	if metadataServiceProxy != "" {
		proxyURL, err := url.Parse(metadataServiceProxy)
		if err != nil {
			return nil, err
		}
		conn.MetadataServiceProxy = proxyURL
	}

	if dataServiceProxy != "" {
		proxyURL, err := url.Parse(dataServiceProxy)
		if err != nil {
			return nil, err
		}
		conn.DataServiceProxy = proxyURL
	}

	err = conn.Refresh()

	if err != nil {
		return nil, err
	}

	err = conn.GetJWT()

	if err != nil {
		return nil, err
	}

	api = &NlxApi{
		Connection: &conn,
	}

	logging.VerboseSuccess("Login Success", loginStart)
	logging.VerboseLn("Access Token: " + conn.AccessToken)

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
