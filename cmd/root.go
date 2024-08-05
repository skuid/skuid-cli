package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/skuid/skuid-cli/pkg/logging"
)

type rootCommander struct {
	factory     *cmdutil.Factory
	loggingOpts logging.LoggingOptions
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

	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Verbose, flags.Verbose)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Trace, flags.Trace)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.Diagnostic, flags.Diagnostic)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.FileLogging, flags.FileLogging)
	cmdutil.AddBoolFlag(cmd, &c.loggingOpts.NoConsole, flags.NoConsole)
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

	if err := c.factory.LogConfig.Setup(&c.loggingOpts); err != nil {
		return err
	}

	logging.Get().Infof(constants.VERSION_TEMPLATE, constants.VERSION_NAME)

	for _, v := range appliedEnvVars {
		if v.Deprecated {
			logging.Get().Warnf("environment variable %q has been deprecated and will be removed in a future release - please migrate to %q", v.Name, cmdutil.EnvVarName(v.Flag.Name))
		}
		logging.Get().Debugf("Using environment variable %v for flag: %v", v.Name, v.Flag.Name)
	}

	return nil
}
