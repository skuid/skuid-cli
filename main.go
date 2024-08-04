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
	factory := cmdutil.NewFactory(VERSION_NAME)
	defer Teardown(factory)
	if err := cmd.NewCmdRoot(factory).Execute(); err != nil {
		// logging might not be setup so output directly to stderr and enable colors
		// which may have been disabled if file logging was enabled
		color.Enable = true
		fmt.Fprintf(os.Stderr, "%v %v\n", color.Red.Sprint("X"), err)
		return exitError
	}

	return exitOK
}

func Teardown(factory *cmdutil.Factory) {
	// We teardown logging (e.g., closing file) here instead of in the command
	// for two reasons:
	//   1. The Cobra *PostRun(E) functions are not called when RunE returns an error
	//   2. Cobra.OnFinalize is global hook so during tests, registering a finalizer
	//      will result in it being called every time a cobra command is executing and
	//      there is no way to "clear/reset" finalizers.  In short, while OnFinalize
	//      approach would work during normal execution as it would run only once, it
	//      impacts testability
	factory.LogConfig.Teardown()
}
