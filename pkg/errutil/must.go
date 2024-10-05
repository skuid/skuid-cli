package errutil

import "fmt"

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustConditionf(condition bool, format string, a ...any) {
	if !condition {
		Must(fmt.Errorf(format, a...))
	}
}
