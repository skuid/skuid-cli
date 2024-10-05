package util

import (
	"testing"

	"github.com/joho/godotenv"

	"github.com/skuid/skuid-cli/pkg/constants"
)

func SkipIntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}

func SkipBenchmark(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}
}

func LoadTestEnvironment() error {
	return godotenv.Load(constants.TestEnvironmentFile)
}
