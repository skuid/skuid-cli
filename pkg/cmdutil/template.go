package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/skuid/skuid-cli/pkg/logging"
	"github.com/spf13/cobra"
)

type CommandFunc func(factory *Factory, cmd *cobra.Command, args []string) error

type CmdTemplate struct {
	Use                    string
	Short                  string
	Long                   string
	Example                string
	Flags                  *CommandFlags
	MutuallyExclusiveFlags [][]string
	Commands               []*cobra.Command
}

func (ct CmdTemplate) ToCommand(factory *Factory, pPreRun CommandFunc, preRun CommandFunc, run CommandFunc) *cobra.Command {
	ct.checkValidTemplate()

	c := cobra.Command{
		SilenceUsage: true,
		Use:          ct.Use,
		Short:        ct.Short,
		Long:         ct.Long,
		Example:      ct.Example,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// initialize is responsible for configuring logging
			// no logs should be emitted prior to it completing successfully
			appliedEnvVars, err := ct.initialize(factory, cmd)
			if err != nil {
				return err
			}

			logging.Get().Infof(constants.VERSION_TEMPLATE, constants.VERSION_NAME)

			for _, v := range appliedEnvVars {
				if v.Deprecated {
					logging.Get().Warnf("environment variable %q has been deprecated and will be removed in a future release - please migrate to %q", v.Name, v.Flag.EnvVarName())
				}
				logging.Get().Debugf("Using environment variable %v for flag: %v", v.Name, v.Flag.FlagName())
			}

			if pPreRun != nil {
				return pPreRun(factory, cmd, args)
			}

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if preRun != nil {
				return preRun(factory, cmd, args)
			}

			return nil
		},
	}

	if ct.Flags != nil {
		factory.Commander.AddFlags(&c, ct.Flags)
		factory.Commander.MarkFlagsMutuallyExclusive(&c, ct.MutuallyExclusiveFlags)
	}

	for _, childCmd := range ct.Commands {
		c.AddCommand(childCmd)
	}

	if run != nil {
		c.RunE = func(cmd *cobra.Command, args []string) error {
			return run(factory, cmd, args)
		}
	}

	return &c
}

func (ct *CmdTemplate) checkValidTemplate() {
	// should never happen in production
	errors.MustConditionf(ct.Use != "", "must provide a value for Use")
	errors.MustConditionf(len(ct.MutuallyExclusiveFlags) == 0 || ct.Flags != nil, "command %q must contain flags to have flags marked mutually exclusive", ct.Use)
}

// applies any environment variables and configures logging
// anything called within this function should not emit any log statements until after logging is configured
func (ct *CmdTemplate) initialize(factory *Factory, cmd *cobra.Command) ([]*AppliedEnvVar, error) {
	appliedEnvVars, err := factory.Commander.ApplyEnvVars(cmd)
	if err != nil {
		return nil, err
	}

	if err := factory.Commander.SetupLogging(cmd, factory.LogConfig); err != nil {
		return nil, err
	}

	return appliedEnvVars, nil
}
