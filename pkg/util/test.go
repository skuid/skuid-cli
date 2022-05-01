package util

import "testing"

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

func LoadTestEnvironment() (err error) {

	return
}
