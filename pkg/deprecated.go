package pkg

import (
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
	"github.com/skuid/tides/pkg/logging"
)

func SkuidNlxLogin(cmd *cobra.Command) (api *NlxApi, err error) {
	var host, username, password string

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

	return SkuidNlxLogin2(host, username, password)
}

func SkuidNlxLogin2(host, username, password string) (api *NlxApi, err error) {

	loginStart := time.Now()

	for _, msg := range [][]string{
		{"Skuid NLX User:", username},
		{"Skuid NLX Host:", host},
	} {
		logging.VerboseF("%-20s\t%s\n", msg[0], color.Green.Sprint(msg[1]))
	}
	logging.VerboseSeparator()

	conn := NlxConnection{
		Host:     host,
		Username: username,
		Password: password,
	}

	err = conn.Refresh()

	if err != nil {
		return
	}

	err = conn.GetJWTAuthorizationToken()

	if err != nil {
		return
	}

	api = &NlxApi{
		Connection: &conn,
	}

	logging.VerboseSuccess("Login Success", loginStart)

	return
}
