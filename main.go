package main

import (
	"os"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/cmd"
	"github.com/skuid/skuid-cli/pkg/logging"
)

func main() {
	Run()
}

// Run is a function so that TestMain can execute it
func Run() {
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
