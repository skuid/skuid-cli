package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/constants"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/cmd"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

// VERSION_NAME has to be here instead of in the constants package
// because embed does not allow relative paths (parent directories)

//go:embed .version
var VERSION_NAME string

func init() {
	VERSION_NAME = strings.TrimSpace(VERSION_NAME)
	constants.VERSION_NAME = VERSION_NAME
}

func main() {
	code := Run()
	os.Exit(int(code))
}

// Run is a function so that TestMain can execute it
func Run() exitCode {
	if err := cmd.NewCmdRoot(cmdutil.NewFactory(VERSION_NAME)).Execute(); err != nil {
		// logging might not be setup so output directly to stderr
		fmt.Fprintf(os.Stderr, "Error: %v\n", color.Red.Sprint(err))
		return exitError
	}

	return exitOK
}
