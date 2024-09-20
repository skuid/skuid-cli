package errutil

import "fmt"

type CriticalError error
type BenignError error

func Critical(msg string, args ...interface{}) (critical CriticalError) {
	critical = fmt.Errorf(msg, args...)
	return
}

func Error(msg string, args ...interface{}) (benign BenignError) {
	benign = fmt.Errorf(msg, args...)
	return
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustConditionf(condition bool, format string, a ...any) {
	if !condition {
		Must(Critical(format, a...))
	}
}
