package cmdutil

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gookit/color"
	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AppliedEnvVar struct {
	Name       string
	Deprecated bool
	Flag       flags.FlagInfo
}

var flagNameValidator = regexp.MustCompile(`^[a-z0-9_-]{3,}$`)

const FlagSetBySkuidCliAnnotation = "skuidcli_annotation_flag_set_by_skuidcli"

type CommandInformer interface {
	AddFlags(cmd *cobra.Command, flags *CommandFlags)
	MarkFlagsMutuallyExclusive(cmd *cobra.Command, flagGroups [][]string)
	ApplyEnvVars(cmd *cobra.Command) ([]*AppliedEnvVar, error)
	SetupLogging(cmd *cobra.Command, li logging.LogInformer) error
}

type Commander struct {
	// we do not need a full blown configuration system, only require environment variables per flag so keep it simple
	FlagManager FlagTrackerFinder
}

func (c *Commander) AddFlags(cmd *cobra.Command, flags *CommandFlags) {
	addFlagType(cmd, c.FlagManager, flags.Bool)
	addFlagType(cmd, c.FlagManager, flags.Int)
	addFlagType(cmd, c.FlagManager, flags.String)
	addFlagType(cmd, c.FlagManager, flags.StringSlice)
}

func (c *Commander) MarkFlagsMutuallyExclusive(cmd *cobra.Command, flagGroups [][]string) {
	for _, flagGroup := range flagGroups {
		cmd.MarkFlagsMutuallyExclusive(flagGroup...)
	}
}

func (c *Commander) ApplyEnvVars(cmd *cobra.Command) ([]*AppliedEnvVar, error) {
	// tracks all errors via error wrapping - no way to iterate flags without going through them all
	var err error
	var appliedVars []*AppliedEnvVar
	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		if _, ok := pf.Annotations[FlagSetBySkuidCliAnnotation]; !ok {
			return
		}

		if applied, errApply := applyEnvVar(cmd, c.FlagManager, pf); errApply != nil {
			if err == nil {
				err = errApply
			} else {
				err = fmt.Errorf("%w; %w", err, errApply)
			}
		} else if applied != nil {
			appliedVars = append(appliedVars, applied)
		}
	})

	return appliedVars, err
}

func (c *Commander) SetupLogging(cmd *cobra.Command, li logging.LogInformer) (err error) {
	verbose, err := cmd.Flags().GetBool(flags.Verbose.Name)
	if err != nil {
		return err
	}

	trace, err := cmd.Flags().GetBool(flags.Trace.Name)
	if err != nil {
		return err
	}

	diagnostic, err := cmd.Flags().GetBool(flags.Diagnostic.Name)
	if err != nil {
		return err
	}

	fileLoggingEnabled, err := cmd.Flags().GetBool(flags.FileLogging.Name)
	if err != nil {
		return err
	}

	loggingDirectory, err := cmd.Flags().GetString(flags.LogDirectory.Name)
	if err != nil {
		return err
	}

	if diagnostic {
		li.SetDiagnostic()
	} else if trace {
		li.SetTrace()
	} else if verbose {
		li.SetVerbose()
	}

	if fileLoggingEnabled {
		if err := li.SetFileLogging(loggingDirectory); err != nil {
			return err
		}
	}

	return nil
}

func NewCommander() CommandInformer {
	return &Commander{
		FlagManager: NewFlagManager(),
	}
}

func applyEnvVar(cmd *cobra.Command, ff FlagFinder, pf *pflag.Flag) (*AppliedEnvVar, error) {
	if f, ok := ff.Find(pf); !ok {
		return nil, errors.Error("unable to locate flag information for: %v", pf.Name)
	} else {
		return applyFlagEnvVar(cmd, pf, f)
	}
}

func applyFlagEnvVar(cmd *cobra.Command, pf *pflag.Flag, r flags.FlagInfo) (*AppliedEnvVar, error) {
	// only attempt to apply env if user did not provide on command line
	if pf.Changed {
		return nil, nil
	}

	// !!!!!
	// WARNING: DO NOT log the envVal as it may contain confidential information
	// !!!!!
	if envVarName, envVal, deprecated, ok := findEnvVar(r); ok {
		// MUST use flagSet.Set to ensure that flag.Changed is updated
		// calling flag.Value.Set will only set the value
		if errSet := cmd.Flags().Set(pf.Name, envVal); errSet != nil {
			return nil, fmt.Errorf("unable to use value from environment variable %v for flag %q: %v", envVarName, pf.Name, errSet)
		}

		// cmd.Flags().Lookup(pf.Name).Value could be used to log the value of the flag at this point
		// because anything confidential will be a Redacted* type and will emit its mask when
		// converted to string
		return &AppliedEnvVar{envVarName, deprecated, r}, nil
	}

	return nil, nil
}

func findEnvVar(r flags.FlagInfo) (string, string, bool, bool) {
	// check if the default environment variable name is set as it takes precendence over legacy
	defaultEnvName := r.EnvVarName()
	if envVal, ok := os.LookupEnv(defaultEnvName); ok {
		return defaultEnvName, envVal, false, true
	}

	for _, envName := range r.LegacyEnvVarNames() {
		if envVal, ok := os.LookupEnv(envName); ok {
			return envName, envVal, true, true
		}
	}

	return "", "", false, false
}

func addFlagType[T ~[]*flags.Flag[F], F flags.FlagType](cmd *cobra.Command, ft FlagTracker, addFlags T) {
	for _, f := range addFlags {
		addFlag(cmd, ft, f)
	}
}

func addFlag[T flags.FlagType](cmd *cobra.Command, ft FlagTracker, flag *flags.Flag[T]) {
	checkValidFlag(flag)

	var fs *pflag.FlagSet
	if flag.Global {
		fs = cmd.PersistentFlags()
	} else {
		fs = cmd.Flags()
	}

	envSuffix := color.Gray.Sprintf("Available as environment variable: %v", flag.EnvVarName())
	usage := color.White.Sprintf("%v\n", flag.Usage) + envSuffix

	switch f := any(flag).(type) {
	case *flags.Flag[string]:
		var parse flags.Parse[string]
		if f.Parse != nil {
			parse = f.Parse
		} else {
			parse = func(val string) (string, error) { return val, nil }
		}
		p := new(string)
		var v *flags.Value[string]
		if f.Redact != nil {
			v = flags.NewValueWithRedact(f.Default, p, parse, f.Redact)
		} else {
			v = flags.NewValue(f.Default, p, parse)
		}
		fs.VarP(v, flag.Name, flag.Shorthand, usage)
	case *flags.Flag[bool]:
		fs.BoolP(f.Name, f.Shorthand, f.Default, usage)
	case *flags.Flag[flags.StringSlice]:
		fs.StringSliceP(f.Name, f.Shorthand, f.Default, usage)
	case *flags.Flag[int]:
		fs.IntP(f.Name, f.Shorthand, f.Default, usage)
	default:
		// should never get here due to type constraints
		panic(fmt.Errorf("unable to handle type for flag %q", flag.Name))
	}

	// mark the flag as created by us since Cobra (and any plugin that may be used in future) will/could
	// create their own flags (e.g., cobra automatically creates a "help" flag)
	// should never error - only error possible is if flag doesn't exist and we just added it
	errors.Must(fs.SetAnnotation(flag.Name, FlagSetBySkuidCliAnnotation, []string{"true"}))

	if flag.Required {
		if flag.Global {
			// should never error - only error possible is if flag doesn't exist and we just added it
			errors.Must(cmd.MarkPersistentFlagRequired(flag.Name))
		} else {
			// should never error - only error possible is if flag doesn't exist and we just added it
			errors.Must(cmd.MarkFlagRequired(flag.Name))
		}
	}

	pf := fs.Lookup(flag.Name)
	// should never happen since we just added it
	errors.MustConditionf(pf != nil, "unable to find flag %q", flag.Name)
	ft.Track(pf, flag)
}

func checkValidFlag[T flags.FlagType](f *flags.Flag[T]) {
	// should never happen in production
	errors.MustConditionf(flagNameValidator.MatchString(f.Name), "flag name %q is invalid", f.Name)
	errors.MustConditionf(len(f.Usage) >= 10, "flag usage %q is invalid for flag name %q", f.Usage, f.Name)
	errors.MustConditionf(f.Parse == nil || supportsParse(f), "flag type %T does not support Parse for flag name %q", f, f.Name)
	errors.MustConditionf(f.Redact == nil || supportsRedact(f), "flag type %T does not support Redact for flag name %q", f, f.Name)
}

func supportsRedact[T flags.FlagType](f *flags.Flag[T]) bool {
	switch any(f).(type) {
	case *flags.Flag[string]:
		return true
	default:
		return false
	}
}

func supportsParse[T flags.FlagType](f *flags.Flag[T]) bool {
	switch any(f).(type) {
	case *flags.Flag[string]:
		return true
	default:
		return false
	}
}
