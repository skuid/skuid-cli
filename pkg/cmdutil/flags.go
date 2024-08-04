package cmdutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	FlagSetBySkuidCliAnnotation = "skuidcli_annotation_flag_set_by_skuidcli"
	LegacyEnvVarsAnnotation     = "skuidcli_annotation_legacy_env_vars"
)

var flagNameValidator = regexp.MustCompile(`^[a-z0-9_-]{3,}$`)

func AddStringFlag(cmd *cobra.Command, valPtr *string, flagInfo *flags.Flag[string]) {
	checkValidFlag(flagInfo, false, false)
	*valPtr = flagInfo.Default
	fs := getFlagSet(cmd, flagInfo)
	fs.StringVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, usage(flagInfo))
	register(fs, flagInfo)
}

func AddBoolFlag(cmd *cobra.Command, valPtr *bool, flagInfo *flags.Flag[bool]) {
	checkValidFlag(flagInfo, false, false)
	*valPtr = flagInfo.Default
	fs := getFlagSet(cmd, flagInfo)
	fs.BoolVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, usage(flagInfo))
	register(fs, flagInfo)
}

func AddIntFlag(cmd *cobra.Command, valPtr *int, flagInfo *flags.Flag[int]) {
	checkValidFlag(flagInfo, false, false)
	*valPtr = flagInfo.Default
	fs := getFlagSet(cmd, flagInfo)
	fs.IntVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, usage(flagInfo))
	register(fs, flagInfo)
}

func AddStringSliceFlag(cmd *cobra.Command, valPtr *[]string, flagInfo *flags.Flag[[]string]) {
	checkValidFlag(flagInfo, false, false)
	*valPtr = flagInfo.Default
	fs := getFlagSet(cmd, flagInfo)
	fs.StringSliceVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, usage(flagInfo))
	register(fs, flagInfo)
}

func AddValueFlag[T flags.FlagType](cmd *cobra.Command, valPtr *T, flagInfo *flags.Flag[T]) {
	checkValidFlag(flagInfo, true, false)
	fs := getFlagSet(cmd, flagInfo)
	v := flags.NewValue(flagInfo.Default, valPtr, flagInfo.Parse)
	fs.VarP(v, flagInfo.Name, flagInfo.Shorthand, usage(flagInfo))
	register(fs, flagInfo)
}

func AddValueWithRedactFlag[T flags.FlagType](cmd *cobra.Command, valPtr *T, flagInfo *flags.Flag[T]) *flags.Value[T] {
	checkValidFlag(flagInfo, true, true)
	fs := getFlagSet(cmd, flagInfo)
	v := flags.NewValueWithRedact(flagInfo.Default, valPtr, flagInfo.Parse, flagInfo.Redact)
	fs.VarP(v, flagInfo.Name, flagInfo.Shorthand, usage(flagInfo))
	register(fs, flagInfo)
	return v
}

func AddAuthFlags(cmd *cobra.Command, opts *pkg.AuthorizeOptions) {
	AddValueFlag(cmd, &opts.Host, flags.Host)
	AddValueFlag(cmd, &opts.Username, flags.Username)
	p := new(string)
	opts.Password = AddValueWithRedactFlag(cmd, p, flags.Password)
}

func EnvVarName(name string) string {
	envVarName := strings.ToUpper(name)
	envVarName = strings.ReplaceAll(envVarName, "-", "_")
	envVarName = constants.ENV_PREFIX + "_" + envVarName
	return envVarName
}

func register[T flags.FlagType](fs *pflag.FlagSet, flagInfo *flags.Flag[T]) {
	setAnnotations(fs, flagInfo)
	if flagInfo.Required {
		cobra.MarkFlagRequired(fs, flagInfo.Name)
	}
}

func setAnnotations[T flags.FlagType](fs *pflag.FlagSet, flagInfo *flags.Flag[T]) {
	// mark the flag as created by us since Cobra (and any plugin that may be used in future) will/could
	// create their own flags (e.g., cobra automatically creates a "help" flag)
	// should never error - only error possible is if flag doesn't exist and we just added it
	errors.Must(fs.SetAnnotation(flagInfo.Name, FlagSetBySkuidCliAnnotation, []string{"true"}))
	errors.Must(fs.SetAnnotation(flagInfo.Name, LegacyEnvVarsAnnotation, flagInfo.LegacyEnvVars))
}

func checkValidFlag[T flags.FlagType](f *flags.Flag[T], allowParse bool, allowRedact bool) {
	// should never happen in production
	errors.MustConditionf(flagNameValidator.MatchString(f.Name), "flag name %q is invalid", f.Name)
	errors.MustConditionf(len(f.Usage) >= 10, "flag usage %q is invalid for flag name %q", f.Usage, f.Name)
	errors.MustConditionf(f.Parse == nil || allowParse, "flag type %T does not support Parse for flag name %q", f, f.Name)
	errors.MustConditionf(f.Redact == nil || allowRedact, "flag type %T does not support Redact for flag name %q", f, f.Name)
}

func usage[T flags.FlagType](flagInfo *flags.Flag[T]) string {
	return fmt.Sprintf("%v\nEnvironment variable: %v", flagInfo.Usage, EnvVarName(flagInfo.Name))
}

func getFlagSet[T flags.FlagType](cmd *cobra.Command, flagInfo *flags.Flag[T]) *pflag.FlagSet {
	if flagInfo.Global {
		return cmd.PersistentFlags()
	}
	return cmd.Flags()
}
