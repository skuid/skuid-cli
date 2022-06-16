package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/skuid/tides/cmd"
	"github.com/skuid/tides/pkg/logging"
)

func main() {
	Run()
}

func Run() {
	if err := cmd.TidesCmd.Execute(); err != nil {
		logging.Get().Errorf("Error Encountered During Run: %v", err)
		os.Exit(1)
	}
}

func init() {
	// if we want to load environments for customers
	// instead of relying on autoenv/direnv
	if err := godotenv.Load(".env"); err != nil {
		logging.Get().Tracef("Error Initializing Environment: %v", err)
	} else {
		// logging.Get().Debug("Initialized Environment")
	}
}
