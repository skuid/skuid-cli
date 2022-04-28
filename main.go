package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/skuid/tides/cmd"
	"github.com/skuid/tides/pkg/logging"
)

func main() {
	if err := cmd.TidesCmd.Execute(); err != nil {
		logging.PrintError("Error Encountered During Run:", err)
		os.Exit(1)
	}
}

func init() {
	// if we want to load environments for customers
	// instead of relying on autoenv/direnv
	if err := godotenv.Load(".env"); err != nil {
		logging.PrintError("Error initializing environment:", err)
		os.Exit(1)
	} else {
		logging.VerboseLn("Loaded Environment")
	}
}
