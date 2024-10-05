package flags

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mmatczuk/anyflag"
	"github.com/sirupsen/logrus"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/skuid/skuid-cli/pkg/metadata"
)

type Parse[F FlagType] func(string) (F, error)
type Redact[T FlagType | ~[]F, F FlagType] func(T) string

type FlagType interface {
	int | string | bool | *time.Time | logrus.Level | metadata.MetadataEntity
}

// Flag is a generic flag type that can take type variables
// just reflecting the external pflags stuff
// T - The type of the flag
// F - When T is a slice, the type of a slice, else the same type as T
//
// Note - Go does not support using Types as the Type of a Type parameter and also doesn't
// support using a constraint when specyfing a slice type.  This presents a couple of
// limitations in the below implementation:
//  1. Requires a two-part generic (e.g., T FlagType | ~[]FlagType would be a one part)
//  2. When specyfing a non-slice type for T, the type for T & F could be different (e.g.,
//     Flag[int, string].  A runtime check to ensure they are always the same is performed
//     within checkValidFlag function in the cmdutil package.  There is discussion in Go
//     regarding supporting type parameters as types themselves so hopefully this will change
//     in future and the the below could be written as Flag[T F | ~[]F, F FlagType].  See the
//     following to track progress of this being added to the language:
//     - https://github.com/golang/go/issues/58590
//     - https://github.com/golang/go/issues/67513
//
// Additionally, Parse & Redact are only supported for Flags that use Value/ValueWithRedact,
// however moving those properties to separate flag structures (e.g., ValueFlag, ValueWithRedactFlag)
// complicates the code with little gain.  Instead, boolean values are utilized within checkValidFlag
// to ensure proper use of these properties.  It's far from ideal but the clarity of having separate
// structures, due to the lack of inheritance support in Go, isn't woth the complexity that it introduces
// in terms of flag creation, testing, etc.
type Flag[T FlagType | ~[]F, F FlagType] struct {
	Name  string // required
	Usage string // required, text shown in usage

	Default       T            // optional
	LegacyEnvVars []string     // optional, additional env var names to check in order of priority
	Required      bool         // flag whether the command requires this flag
	Shorthand     string       // optional, will change call to allow for shorthand
	Global        bool         // is this a global/persistent flag?
	Parse         Parse[F]     // optional, parse user-input during pflag.Set() - only supported by Value & ValueWithRedact
	Redact        Redact[T, F] // optional, mask value - only supported by ValueWithRedact
}

func UseDebugMessage() string {
	return fmt.Sprintf("use %v for details", logging.QuoteText(fmt.Sprintf("--%v %v", LogLevel.Name, logrus.DebugLevel)))
}

func FormatSince(since *time.Time) string {
	if since == nil {
		return ""
	} else {
		return since.Format(time.RFC3339)
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

// SliceValue[T] is a wrapper for anyflag.SliceValue[T] to address anyflag's issue
// of correctly resolving the Type() of T
type SliceValue[T any] struct {
	*anyflag.SliceValue[T]
}

func (v *SliceValue[T]) Type() string {
	val := *new([]T)
	asAny := any(val)

	// NOTE - Expand cases as needed to force a type name
	switch asAny.(type) {
	case []string:
		return "stringSlice"
	case []bool:
		return "boolSlice"
	case []int:
		return "intSlice"
	}

	return getTypeName(reflect.TypeOf(val))
}

func NewSliceValue[T any](val []T, p *[]T, parse func(val string) (T, error)) *SliceValue[T] {
	return &SliceValue[T]{
		anyflag.NewSliceValue(val, p, parse),
	}
}

func NewSliceValueWithRedact[T any](val []T, p *[]T, parse func(val string) (T, error), redact func(T) string) *SliceValue[T] {
	return &SliceValue[T]{
		anyflag.NewSliceValueWithRedact(val, p, parse, redact),
	}
}

// anyflag does not correctly resolve the type so we wrap Value[T] and SliceValue[T] in order to override Type()
func getTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return fmt.Sprint(t)
}
