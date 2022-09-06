package errors

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
