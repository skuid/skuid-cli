package flags

import (
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/pkg/errors"
	"github.com/skuid/tides/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Generic flag type that can take a type variable and use a pointer
// to that type as the thing we're going to add
// just reflecting the external pflags stuff
type Flag[T any] struct {
	// we're making argument private because I want you to access flags
	// through the command pointer that we get from the function
	argument *T // required

	Name string // required
	// Aliases    []string // optional
	Default    T      // optional, overridden by existing environment variable
	EnvVarName string // optional, overrides default value if exists
	Usage      string // required, text shown in usage
	Required   bool   // flag whether the command requires this flag
	Shorthand  string // optional, will change call to allow for shorthand

	Global bool // is this a global/persistent flag?
}

func CheckRequiredFields[T any](f *Flag[T]) error {
	if f.argument == nil {
		return errors.Critical("Flag Argument must be provided.")
	}

	if f.Name == "" {
		return errors.Critical("Flag FlagName must be provided.")
	}

	if f.Usage == "" {
		return errors.Critical("Flag UsageText must be provided.")
	}

	return nil
}

// AddFlagFunctions is a function that takes a command and adds a series of flags
// and if there's an issue adding the flags (usually just a null pointer)
// we'll log fatal
func AddFlagFunctions(cmd *cobra.Command, adds ...func(*cobra.Command) error) {
	for _, addFlag := range adds {
		if err := addFlag(cmd); err != nil {
			logging.Fatalf("Unable to set up %v: %v", cmd.Name(), err.Error())
		}
	}
}

// Same as AddFlagFunctions but for any type of flag (of consistent type... ugh)
func AddFlags[T any](cmd *cobra.Command, flags ...*Flag[T]) {
	for _, flag := range flags {
		if err := Add(flag)(cmd); err != nil {
			logging.Fatalf("Unable to set up flag (%v) for command (%v): %v", flag.Name, cmd.Name(), err.Error())
		}
	}
}

// If we find an environment variable indicate that in the usage text so that there
// is some additional information given
func environmentVariableDetectionString(environmentVariableName, usageText string) string {
	environmentVariableValue := os.Getenv(environmentVariableName)
	usageText = color.White.Sprint(usageText)
	if environmentVariableValue != "" {
		usageText = usageText + "\n" +
			color.Gray.Sprintf("%v %v", color.Green.Sprint(environmentVariableName), color.Green.Sprint(constants.CHECKMARK))
	} else {
		usageText = usageText + "\n" +
			color.Gray.Sprintf("Available as environment variable: %v", color.Yellow.Sprint(environmentVariableName))
	}
	return usageText
}

func aliasInformationString(flagName, usageText string) string {
	return color.Gray.Sprintf("Alias for '--%v'\n", flagName) + usageText
}

func Add[T any](flag *Flag[T]) func(*cobra.Command) error {
	return func(to *cobra.Command) error {
		// three required fields
		if err := CheckRequiredFields(flag); err != nil {
			return err
		}

		// these are the three variables
		// that are going to change based on what
		// we're getting
		required := flag.Required
		usageText := flag.Usage

		var flags *pflag.FlagSet
		if flag.Global {
			flags = to.PersistentFlags()
		} else {
			flags = to.Flags()
		}

		// TODO: comment this and explain what we're doing with
		// the type switch
		switch f := any(flag).(type) {
		// handle string
		case *Flag[string]:
			defaultVar := f.Default
			// override default values and usage text if
			// there is an environment variable provided
			if flag.EnvVarName != "" {
				defaultVar = os.Getenv(flag.EnvVarName)
				usageText = environmentVariableDetectionString(flag.EnvVarName, flag.Usage)
				if defaultVar != "" {
					// the only time we disabled required
					// is when we have the environment variable name
					// and we find an environment variable value
					required = false
				}
			}
			if flag.Shorthand != "" {
				flags.StringVarP(f.argument, flag.Name, flag.Shorthand, defaultVar, usageText)
			} else {
				flags.StringVar(f.argument, flag.Name, defaultVar, usageText)
			}
			// if len(flag.Aliases) > 0 {
			// 	for _, alias := range flag.Aliases {
			// 		flags.StringVar(f.argument, alias, defaultVar, aliasInformationString(flag.Name, usageText))
			// 	}
			// }

		// handle bools
		case *Flag[bool]:
			defaultValue := f.Default
			// override default values and usage text if
			// there is an environment variable provided
			if flag.EnvVarName != "" {
				defaultValue = func() bool {
					val := strings.ToLower(os.Getenv(flag.EnvVarName))
					// if there's something set
					// disable requirement, we're going to use
					// the environment variable as the default value
					if val != "" && required {
						required = false
					}
					// we will accept VALUE == "true/TRUE/tRuE"
					if strings.EqualFold(val, "true") {
						return true
					}
					// we will accept VALUE != 0 (VALUE=1)
					if i, err := strconv.Atoi(val); err == nil && i != 0 {
						return true
					}
					// otherwise, it's false
					return false
				}()
				usageText = environmentVariableDetectionString(flag.EnvVarName, flag.Usage)
			}

			if flag.Shorthand != "" {
				flags.BoolVarP(f.argument, flag.Name, flag.Shorthand, defaultValue, usageText)
			} else {
				flags.BoolVar(f.argument, flag.Name, defaultValue, usageText)
			}

			// if len(flag.Aliases) > 0 {
			// 	for _, alias := range flag.Aliases {
			// 		flags.BoolVar(f.argument, alias, defaultValue, aliasInformationString(flag.Name, usageText))
			// 	}
			// }

		// handle string arrays
		case *Flag[[]string]:
			defaultVar := f.Default
			// override default values and usage text if
			// there is an environment variable provided
			if flag.EnvVarName != "" {
				defaultVar = strings.Split(os.Getenv(flag.EnvVarName), ",")
				usageText = environmentVariableDetectionString(flag.EnvVarName, flag.Usage)
				if len(defaultVar) > 0 {
					// the only time we disabled required
					// is when we have the environment variable name
					// and we find an environment variable value
					required = false
				}
			}

			if flag.Shorthand != "" {
				flags.StringArrayVarP(f.argument, flag.Name, flag.Shorthand, defaultVar, usageText)
			} else {
				flags.StringArrayVar(f.argument, flag.Name, defaultVar, usageText)
			}

			// if len(flag.Aliases) > 0 {
			// 	for _, alias := range flag.Aliases {
			// 		flags.StringArrayVar(f.argument, alias, defaultVar, aliasInformationString(flag.Name, usageText))
			// 	}
			// }

		default:
			return errors.Critical("No type definition found")
		}

		if required {
			return to.MarkFlagRequired(flag.Name)
		} else {
			return nil
		}
	}
}
