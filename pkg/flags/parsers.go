package flags

import (
	"time"

	"github.com/skuid/skuid-cli/pkg/util"
)

func ParseSince(val string) (CustomString, error) {
	if ts, err := util.GetTimestamp(val, time.Now(), true); err != nil {
		return "", err
	} else {
		return CustomString(ts), nil
	}
}
