package main_test

import (
	"log"
	"os"
	"testing"

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
	os.Exit(m.Run())
}
