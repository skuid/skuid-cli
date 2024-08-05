package flags

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mmatczuk/anyflag"
	"github.com/sirupsen/logrus"
)

type Parse[T FlagType] func(string) (T, error)
type Redact[T FlagType] func(T) string

type FlagType interface {
	int | string | bool | []string | *time.Time | logrus.Level
}

// Flag is a generic flag type that can take a type variable
// just reflecting the external pflags stuff
type Flag[T FlagType] struct {
	Name  string // required
	Usage string // required, text shown in usage

	Default       T         // optional
	LegacyEnvVars []string  // optional, additional env var names to check in order of priority
	Required      bool      // flag whether the command requires this flag
	Shorthand     string    // optional, will change call to allow for shorthand
	Global        bool      // is this a global/persistent flag?
	Parse         Parse[T]  // optional, parse user-input during pflag.Set() - only supported by Value & ValueWithRedact
	Redact        Redact[T] // optional, mask value - only supported by ValueWithRedact
}

// Value[T] is a wrapper for anyflag.Value[T] to address anyflag's issue
// of correctly resolving the Type() of T
type Value[T any] struct {
	*anyflag.Value[T]
}

func (v *Value[T]) Type() string {
	return getTypeName(reflect.TypeOf(*new(T)))
}

func NewValue[T any](val T, p *T, parse func(val string) (T, error)) *Value[T] {
	return &Value[T]{
		anyflag.NewValue(val, p, parse),
	}
}

func NewValueWithRedact[T any](val T, p *T, parse func(val string) (T, error), redact func(T) string) *Value[T] {
	return &Value[T]{
		anyflag.NewValueWithRedact(val, p, parse, redact),
	}
}

func getTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return fmt.Sprint(t)
}
