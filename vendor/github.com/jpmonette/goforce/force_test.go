package force

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TestLimits is testing the Limits endpoint
func TestLimits(t *testing.T) {
	conn, err := Login(
		os.Getenv("CONSUMERKEY"),
		os.Getenv("CONSUMERSECRET"),
		os.Getenv("INSTANCEURL"),
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("VERSION"),
	)

	if assert.Nil(t, err) {

		limits, err := conn.Limits()

		assert.Nil(t, err)

		assert.Equal(t, 999999999, limits.HourlyDashboardStatuses.Remaining)
	}
}
