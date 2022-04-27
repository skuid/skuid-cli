package main_test

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gookit/color"
	"github.com/joho/godotenv"
)

func loadEnv() error {
	return godotenv.Load(".testenv")
}

func init() {
	err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	rc := m.Run()

	var threshold float64 = 0

	for _, arg := range os.Args {
		if strings.Contains(arg, "coverfail") {
			if strings.Contains(arg, "true") {
				threshold = 0.8
				color.Green.Printf("failing if coverage is below threshold of %v\n", threshold)
			}
		}
	}

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < threshold {
			color.Red.Printf("\nTests passed but coverage failed at: %2.2f%%\n", c*100)
			rc = -1
		}
	}
	os.Exit(rc)

}
