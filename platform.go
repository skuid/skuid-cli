package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gookit/color"
)

// Login logs a given user into a target Skuid Platform site and returns a RestApi connection
// that can be used to make HTTP requests
func PlatformLogin(host, username, password, apiVersion, metadataServiceProxy, dataServiceProxy string, verbose bool) (api *PlatformRestApi, err error) {

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

	return api, nil
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
