package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		PrintError("Error Encountered During Run:", err)
		os.Exit(1)
	}
}

func init() {
	// if we want to load environments for customers
	// instead of relying on autoenv/direnv
	if err := godotenv.Load(".env"); err != nil {
		PrintError("Error initializing environment:", err)
		os.Exit(1)
	}
}
