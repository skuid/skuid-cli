package cmdutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AppliedEnvVar struct {
	Name       string
	Deprecated bool
	Flag       *pflag.Flag
}

func ApplyEnvVars(cmd *cobra.Command) ([]*AppliedEnvVar, error) {
	// tracks all errors via error wrapping - no way to iterate flags without going through them all
	var err error
	var appliedVars []*AppliedEnvVar
	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		// ignore everything except our flags - cobra (and any plugin that may be used in the future) will/could
		// add its own flags (e.g. cobra automatically created a "help" flag)
		if _, ok := pf.Annotations[FlagSetBySkuidCliAnnotation]; !ok {
			return
		}

		if applied, errApply := applyEnvVar(cmd, pf); errApply != nil {
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

func applyEnvVar(cmd *cobra.Command, pf *pflag.Flag) (*AppliedEnvVar, error) {
	// only attempt to apply env if user did not provide on command line
	if pf.Changed {
		return nil, nil
	}

	// !!!!!
	// WARNING: DO NOT log the envVal as it may contain confidential information
	// !!!!!
	if envVarName, envVal, deprecated, ok := findEnvVar(pf); ok {
		// MUST use flagSet.Set to ensure that flag.Changed is updated
		// calling flag.Value.Set will only set the value
		if errSet := cmd.Flags().Set(pf.Name, envVal); errSet != nil {
			return nil, fmt.Errorf("unable to use value from environment variable %v for flag %q: %v", envVarName, pf.Name, errSet)
		}

		// cmd.Flags().Lookup(pf.Name).Value could be used to log the value of the flag at this point
		// because anything confidential will be a Redacted* type and will emit its mask when
		// converted to string
		return &AppliedEnvVar{envVarName, deprecated, pf}, nil
	}

	return nil, nil
}

func findEnvVar(pf *pflag.Flag) (string, string, bool, bool) {
	// check if the default environment variable name is set as it takes precendence over legacy
	defaultEnvName := EnvVarName(pf.Name)
	if envVal, ok := os.LookupEnv(defaultEnvName); ok {
		return defaultEnvName, envVal, false, true
	}

	if legacyEnvVarNames, ok := pf.Annotations[LegacyEnvVarsAnnotation]; ok {
		for _, envName := range legacyEnvVarNames {
			if envVal, ok := os.LookupEnv(envName); ok {
				return envName, envVal, true, true
			}
		}
	}

	return "", "", false, false
}
