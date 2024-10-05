package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/errutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

type legacyLoggingOptions struct {
	logging.LoggingOptions
	Trace      bool
	Diagnostic bool
	Verbose    bool
}

type rootCommander struct {
	factory     *cmdutil.Factory
	loggingOpts legacyLoggingOptions
}

func (c *rootCommander) GetCommand() *cobra.Command {
	rootTemplate := cmdutil.CmdTemplate{
		Use:   constants.PROJECT_NAME,
		Short: "A CLI for Skuid APIs",
		Long:  `A command-line interface used to retrieve and deploy Skuid NLX sites.`,
	}

	cmd := rootTemplate.ToCommand(nil)
	cmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	cmd.SilenceErrors = true
	cmd.Version = c.factory.AppVersion
	cmd.SetVersionTemplate(fmt.Sprintf(constants.VERSION_TEMPLATE+"\n", c.factory.AppVersion))
	cmd.PersistentPreRunE = c.setup

	cmdutil.AddValueFlag(cmd, &c.loggingOpts.Level, flags.LogLevel)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Diag, flags.Diag)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Verbose, flags.Verbose)
	// should never error - only error possible is if flag doesn't exist or usageMessage empty
	errutil.Must(cmd.PersistentFlags().MarkDeprecated(flags.Verbose.Name, fmt.Sprintf("it will be removed in a future release, please migrate to --%v", flags.LogLevel.Name)))
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Trace, flags.Trace)
	// should never error - only error possible is if flag doesn't exist or usageMessage empty
	errutil.Must(cmd.PersistentFlags().MarkDeprecated(flags.Trace.Name, fmt.Sprintf("it will be removed in a future release, please migrate to --%v", flags.LogLevel.Name)))
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Diagnostic, flags.Diagnostic)
	// should never error - only error possible is if flag doesn't exist or usageMessage empty
	errutil.Must(cmd.PersistentFlags().MarkDeprecated(flags.Diagnostic.Name, fmt.Sprintf("it will be removed in a future release, please migrate to --%v", flags.LogLevel.Name)))
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.FileLogging, flags.FileLogging)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.NoConsoleLogging, flags.NoConsoleLogging)
	cmdutil.AddStringFlag(cmd, &c.loggingOpts.FileLoggingDir, flags.LogDirectory)

	cmd.AddCommand(NewCmdDeploy(c.factory))
	cmd.AddCommand(NewCmdRetrieve(c.factory))
	cmd.AddCommand(NewCmdWatch(c.factory))

	// cobra's default version command starts with lower-case usage text but our
	// flags start with uppercase so ensure consistent format for the flag
	// version is only available on root command
	// we don't use cmdutil helpers because we don't want it registered
	// since its technically not our flag
	cmd.Flags().Bool("version", false, "Show version for "+cmd.Name())

	// can't specify verbose/trace/diagnostic which have been deprecated
	// when using log-level and/or debug which replaces them all
	cmd.MarkFlagsMutuallyExclusive(flags.LogLevel.Name, flags.Verbose.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.LogLevel.Name, flags.Trace.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.LogLevel.Name, flags.Diagnostic.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Diag.Name, flags.Verbose.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Diag.Name, flags.Trace.Name)
	cmd.MarkFlagsMutuallyExclusive(flags.Diag.Name, flags.Diagnostic.Name)

	return cmd
}

func NewCmdRoot(factory *cmdutil.Factory) *cobra.Command {
	commander := new(rootCommander)
	commander.factory = factory
	return commander.GetCommand()
}

// setup is responsible for configuring logging
// no logs should be emitted prior to it completing successfully
func (c *rootCommander) setup(cmd *cobra.Command, args []string) error {
	appliedEnvVars, err := cmdutil.ApplyEnvVars(cmd)
	if err != nil {
		return err
	}

	// get effective logging options to account for deprecated flags
	effectiveLoggingOptions := c.getLoggingOptions(cmd)
	if err := c.factory.LogConfig.Setup(&effectiveLoggingOptions); err != nil {
		return err
	}

	logger := logging.WithName("cmd.setup")
	logger.Debugf(constants.VERSION_TEMPLATE, constants.VERSION_NAME)

	for _, v := range appliedEnvVars {
		if v.Deprecated {
			logger.Warnf("%v: %v will be removed in a future release, please migrate to %v", logging.ColorWarning.Text("Deprecated environment variable detected"), logging.ColorEnvVar.Text(v.Name), logging.ColorEnvVar.Text(cmdutil.EnvVarName(v.Flag.Name)))
		}
		logger.Debugf("Using environment variable %v for flag %v", logging.ColorEnvVar.Text(v.Name), logging.ColorEnvVar.Text("--"+v.Flag.Name))
	}

	return nil
}

func (c *rootCommander) getLoggingOptions(cmd *cobra.Command) logging.LoggingOptions {
	fs := cmd.Flags()

	// if log-level and/or debug has been set, we ignore any deprecated flags
	if !fs.Changed(flags.LogLevel.Name) && !fs.Changed(flags.Diag.Name) {
		// map legacy flags to their equivalent level & debug values
		if fs.Changed(flags.Diagnostic.Name) {
			c.loggingOpts.Level = logrus.TraceLevel
			c.loggingOpts.Diag = true
		} else if fs.Changed(flags.Trace.Name) {
			c.loggingOpts.Level = logrus.TraceLevel
			c.loggingOpts.Diag = false
		} else if fs.Changed(flags.Verbose.Name) {
			c.loggingOpts.Level = logrus.DebugLevel
			c.loggingOpts.Diag = false
		}
	}

	return c.loggingOpts.LoggingOptions
}
