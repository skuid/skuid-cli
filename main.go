package main

import (
	"os"

	"github.com/gookit/color"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		color.Errorf("Error executing: %v", err.Error())
		os.Exit(1)
	}
}
