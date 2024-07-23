package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/errors"
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
			if ct.Flags != nil {
				if err := factory.Commander.ApplyEnvVars(cmd); err != nil {
					return err
				}

				if err := factory.Commander.SetupLogging(cmd, factory.LogConfig); err != nil {
					return err
				}
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
