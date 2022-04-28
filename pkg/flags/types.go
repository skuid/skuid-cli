package flags

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Generic flag type that can take a type variable and use a pointer
// to that type as the thing we're going to add
type Flag[T any] struct {
	// we're making argument private because I want you to access flags
	// through the command pointer that we get from the function
	argument *T // required

	Name       string // required
	Default    T      // optional, overridden by existing environment variable
	EnvVarName string // optional, overrides default value if exists
	Usage      string // required, text shown in usage
	Required   bool   // flag whether the command requires this flag
	Shorthand  string // optional, will change call to allow for shorthand

	Global bool // is this a global/persistent flag?
}

func CheckRequiredFields[T any](f *Flag[T]) error {
	if f.argument == nil {
		return fmt.Errorf("StringFlag Argument must be provided.")
	}

	if f.Name == "" {
		return fmt.Errorf("StringFlag FlagName must be provided.")
	}

	if f.Usage == "" {
		return fmt.Errorf("StringFlag UsageText must be provided.")
	}

	return nil
}

// AddFlagFunctions is a function that takes a command and adds a series of flags
// and if there's an issue adding the flags (usually just a null pointer)
// we'll log fatal
func AddFlagFunctions(cmd *cobra.Command, adds ...func(*cobra.Command) error) {
	for _, addFlag := range adds {
		if err := addFlag(cmd); err != nil {
			log.Fatalf("Unable to set up %v: %v", cmd.Name(), err.Error())
		}
	}
}

func AddFlags[T any](cmd *cobra.Command, flags ...*Flag[T]) {
	for _, flag := range flags {
		if err := Add(flag)(cmd); err != nil {
			log.Fatalf("Unable to set up flag (%v) for command (%v): %v", flag.Name, cmd.Name(), err.Error())
		}
	}
}

func FlagFunctions[T any](flags ...*Flag[T]) (adds []func(*cobra.Command) error) {
	for _, f := range flags {
		adds = append(adds, Add(f))
	}
	return
}

// If we find an environment variable indicate that in the usage text so that there
// is some additional information given
func EnvironmentVariableFound(environmentVariableName, usageText string) string {
	environmentVariableValue := os.Getenv(environmentVariableName)
	usageText = color.Gray.Sprint(usageText)
	if environmentVariableValue != "" {
		usageText = usageText + "\n" +
			color.Green.Sprintf(`(Environment Variable: "$%v" FOUND)`, environmentVariableName)
	} else {
		usageText = usageText + "\n" +
			color.Yellow.Sprintf(`(Environment Variable: "$%v")`, environmentVariableName)
	}
	return usageText
}

// TODO: when they add type switching
func Add[T any](flag *Flag[T]) func(*cobra.Command) error {
	switch f := any(flag).(type) {
	case *Flag[string]:
		return AddString(f)
	case *Flag[bool]:
		return AddBool(f)
	case *Flag[[]string]:
		return AddStringArray(f)
	default:
		return func(cmd *cobra.Command) error {
			return fmt.Errorf("No type definition found")
		}
	}
}

// AddTo takes a StringFlag and adds it to a command.
// There's a lot of duplicate logic that can be handled
// by this stuff
func AddString(flag *Flag[string]) func(*cobra.Command) error {
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
		defaultVar := flag.Default

		// override default values and usage text if
		// there is an environment variable provided
		if flag.EnvVarName != "" {
			defaultVar = os.Getenv(flag.EnvVarName)
			usageText = EnvironmentVariableFound(flag.EnvVarName, flag.Usage)
			if defaultVar != "" {
				// the only time we disabled required
				// is when we have the environment variable name
				// and we find an environment variable value
				required = false
			}
		}

		var flags *pflag.FlagSet
		if flag.Global {
			flags = to.PersistentFlags()
		} else {
			flags = to.Flags()
		}

		if flag.Shorthand != "" {
			flags.StringVarP(
				flag.argument,
				flag.Name,
				flag.Shorthand,
				defaultVar,
				usageText,
			)
		} else {
			flags.StringVar(
				flag.argument,
				flag.Name,
				defaultVar,
				usageText,
			)
		}

		if required {
			return to.MarkFlagRequired(flag.Name)
		} else {
			return nil
		}
	}
}

// AddTo takes a StringArrayFlag and adds it to a command.
// There's a lot of duplicate logic that can be handled
// by this stuff
func AddStringArray(flag *Flag[[]string]) func(*cobra.Command) error {
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
		defaultVar := flag.Default

		// override default values and usage text if
		// there is an environment variable provided
		if flag.EnvVarName != "" {
			defaultVar = strings.Split(os.Getenv(flag.EnvVarName), ",")
			usageText = EnvironmentVariableFound(flag.EnvVarName, flag.Usage)
			if len(defaultVar) > 0 {
				// the only time we disabled required
				// is when we have the environment variable name
				// and we find an environment variable value
				required = false
			}
		}

		var flags *pflag.FlagSet
		if flag.Global {
			flags = to.PersistentFlags()
		} else {
			flags = to.Flags()
		}

		if flag.Shorthand != "" {
			flags.StringArrayVarP(
				flag.argument,
				flag.Name,
				flag.Shorthand,
				defaultVar,
				usageText,
			)
		} else {
			flags.StringArrayVar(
				flag.argument,
				flag.Name,
				defaultVar,
				usageText,
			)
		}

		if required {
			return to.MarkFlagRequired(flag.Name)
		} else {
			return nil
		}
	}
}

func AddBool(flag *Flag[bool]) func(cmd *cobra.Command) error {
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
		defaultValue := flag.Default

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
			usageText = EnvironmentVariableFound(flag.EnvVarName, flag.Usage)
		}

		var flags *pflag.FlagSet
		if flag.Global {
			flags = to.PersistentFlags()
		} else {
			flags = to.Flags()
		}

		if flag.Shorthand != "" {
			flags.BoolVarP(
				flag.argument,
				flag.Name,
				flag.Shorthand,
				defaultValue,
				usageText,
			)
		} else {
			flags.BoolVar(
				flag.argument,
				flag.Name,
				defaultValue,
				usageText,
			)
		}

		if required {
			return to.MarkFlagRequired(flag.Name)
		} else {
			return nil
		}
	}
}
