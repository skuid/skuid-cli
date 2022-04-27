package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

func PlatformLogin(cmd *cobra.Command) (api *PlatformRestApi, err error) {
	var host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy string
	var verbose bool

	f := cmd.Flags()

	if host, err = f.GetString(FlagNameHost); err != nil {
		return
	}

	if username, err = f.GetString(FlagNameUsername); err != nil {
		return
	}

	if password, err = f.GetString(FlagNamePassword); err != nil {
		return
	}

	if verbose, err = f.GetBool(FlagNameVerbose); err != nil {
		return
	}

	if apiVersion, err = f.GetString(FlagNameApiVersion); err != nil {
		return
	}

	if metadataServiceProxy, err = f.GetString(FlagNameMetadataServiceProxy); err != nil {
		return
	}

	if dataServiceProxy, err = f.GetString(FlagNameDataserviceProxy); err != nil {
		return
	}

	if apiVersion == "" {
		apiVersion = "2"
	}

	loginStart := time.Now()

	if verbose {
		for _, msg := range [][]string{
			{"Logging in to Skuid Platform as user:", username},
			{"Connecting to host:", host},
			{"API Version:", apiVersion},
		} {
			Printf("%-45s\t%s\n", color.Yellow.Sprint(msg[0]), color.Green.Sprint(msg[1]))
		}
		Separator()
	}

	conn := PlatformRestConnection{
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

	api = &PlatformRestApi{
		Connection: &conn,
	}

	if verbose {
		SuccessWithTime("Login Success", loginStart)
		Println("Access Token: " + conn.AccessToken)
	}

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
