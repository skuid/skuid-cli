package logging

import "github.com/skuid/tides/pkg/errors"

func Error(err error) {
	switch err.(type) {
	case errors.CriticalError:
		Fatal(err)
	case errors.BenignError:
		VerboseF("Error Discarded: %v", err)
	default:
		DebugF("Error Encountered: %v", err)
	}
}
