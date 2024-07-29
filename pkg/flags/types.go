package flags

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mmatczuk/anyflag"
	"github.com/spf13/pflag"

	"github.com/skuid/skuid-cli/pkg/constants"
)

type StringSlice []string
type RedactedString string
type CustomString string
type Parse[T FlagType] func(string) (T, error)

type FlagType interface {
	int | string | RedactedString | bool | StringSlice | CustomString
}

type FlagInfo interface {
	EnvVarName() string
	LegacyEnvVarNames() []string
	IsRequired() bool
}

// Flag is a generic flag type that can take a type variable
// just reflecting the external pflags stuff
type Flag[T FlagType] struct {
	Name  string // required
	Usage string // required, text shown in usage

	Default       T        // optional
	LegacyEnvVars []string // optional, additional env var names to check in order of priority
	Required      bool     // flag whether the command requires this flag
	Shorthand     string   // optional, will change call to allow for shorthand
	Global        bool     // is this a global/persistent flag?
	Parse         Parse[T] // optional, parse user-input during pflag.Set() - only supported by CustomString and RedactedString currently
}

func (f *Flag[T]) EnvVarName() string {
	envVarName := strings.ToUpper(f.Name)
	envVarName = strings.ReplaceAll(envVarName, "-", "_")
	envVarName = constants.ENV_PREFIX + "_" + envVarName
	return envVarName
}

func (f *Flag[T]) LegacyEnvVarNames() []string {
	return f.LegacyEnvVars
}

func (f *Flag[T]) IsRequired() bool {
	return f.Required
}

func GetRedactedString(fs *pflag.FlagSet, name string) (*Value[RedactedString], error) {
	if rs, err := getFlagValue[RedactedString](fs, name); err != nil {
		return nil, err
	} else {
		return rs, nil
	}
}

func GetCustomString(fs *pflag.FlagSet, name string) (string, error) {
	if cs, err := getFlagValue[CustomString](fs, name); err != nil {
		return "", err
	} else {
		return cs.String(), nil
	}
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

func getFlagValue[T FlagType](fs *pflag.FlagSet, name string) (*Value[T], error) {
	f := fs.Lookup(name)
	if f == nil {
		return nil, fmt.Errorf("flag accessed but not defined: %s", name)
	}
	v, ok := f.Value.(*Value[T])
	if !ok {
		return nil, fmt.Errorf("trying to get %T value of flag of type %s", *new(T), f.Value.Type())
	}
	return v, nil
}
