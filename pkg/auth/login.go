package auth

import (
	"time"

	"github.com/spf13/cobra"
)

// TODO: This is how I'm going to store the credentials
// and login information for each run

// For each host we want to store an authentication
// that can be easily retrieved
type Store map[string]Authentication

// Authentication is a struct that we store
// to allow users to be able to type in
// multiple passwords and login whenever
// they jump in and more
type Authentication struct {
	Host        string
	Expiration  time.Time
	AccessToken string
	AuthToken   string
}

// Credentials is for logging into Skuid
// this is going to be like what is
// prompted when TokenStore expires
type Credentials struct {
	Username string
	Password string
}

type CommandExecution struct {
	source    *cobra.Command
	arguments []string

	// todo: return information / logs
}

type Runs []CommandExecution
