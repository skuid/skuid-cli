package logging

import "github.com/skuid/tides/pkg/errors"

func Error(err error) {
	switch err.(type) {
	case errors.CriticalError:
		Fatal(err)
	case errors.BenignError:
		Debugf("Error Discarded: %v", err)
	default:
		TraceF("Error Encountered: %v", err)
	}
}
