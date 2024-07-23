package flags

import (
	"fmt"
	"strings"

	"github.com/mmatczuk/anyflag"
	"github.com/spf13/pflag"

	"github.com/skuid/skuid-cli/pkg/constants"
)

type StringSlice []string
type RedactedString string

type FlagType interface {
	int | string | RedactedString | bool | StringSlice
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

// flags are command specific so grab the password from the flags for the specific command
//
// NOTE - This is a fairly inelgent solution but password is a one-off use case so keeping it simple.  If/When other
// flags are needed for things like RestrictedString, a more universal approach can be taken.
//
// NOTE - The auth package will unredact the password and clear its value immediately after so this is essentially a
// one-time use password albeit somewhat inelegant approach.
//
// TODO: Implement a solution for secure storage of the password while in memory and implement a proper one-time use
// approach assuming Skuid supports refresh tokens (see https://github.com/skuid/skuid-cli/issues/172)
func GetPassword(fs *pflag.FlagSet) (*anyflag.Value[RedactedString], error) {
	v := fs.Lookup(Password.Name)
	if v == nil {
		return nil, fmt.Errorf("unable to obtain the value for the password specified (code=1)")
	}
	rs, ok := v.Value.(*anyflag.Value[RedactedString])
	if !ok {
		return nil, fmt.Errorf("unable to obtain the value for the password specified (code=2)")
	}
	return rs, nil
}
