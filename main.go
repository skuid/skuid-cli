package main

import (
	_ "embed"
	"os"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/cmd"
	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/logging"
)

// VERSION_NAME has to be here instead of in the constants package
// because embed does not allow relative paths (parent directories)

//go:embed .version
var VERSION_NAME string

func main() {
	Run()
}

// Run is a function so that TestMain can execute it
func Run() {
	cmd.VERSION_NAME = VERSION_NAME
	pkg.VERSION_NAME = VERSION_NAME
	logging.Get().Infof("Skuid CLI Version %v", VERSION_NAME)
	if err := cmd.SkuidCmd.Execute(); err != nil {
		logging.Get().WithError(err).Errorf("Error Encountered During Run: %v", color.Red.Sprint(err))
		os.Exit(1)
	}
}

//func init() {
//	// if we want to load environments for customers
//	// instead of relying on autoenv/direnv
//	if err := godotenv.Load(".env"); err != nil {
//		logging.Get().Tracef("Error Initializing Environment: %v", err)
//	}
//}
