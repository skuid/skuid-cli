package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/errors"
	"github.com/spf13/cobra"
)

type CobraRunE func(cmd *cobra.Command, args []string) error

type CmdTemplate struct {
	Use     string
	Short   string
	Long    string
	Example string
}

func (ct *CmdTemplate) ToCommand(run CobraRunE) *cobra.Command {
	ct.checkValidTemplate(run)

	return &cobra.Command{
		SilenceUsage: true,
		Use:          ct.Use,
		Short:        ct.Short,
		Long:         ct.Long,
		Example:      ct.Example,
		RunE:         run,
	}
}

func (ct *CmdTemplate) checkValidTemplate(run CobraRunE) {
	// should never happen in production
	errors.MustConditionf(ct.Use != "", "must provide a value for Use")
	errors.MustConditionf(ct.Short != "", "must provide a value for Short")
	errors.MustConditionf(ct.Long != "", "must provide a value for Long")
	errors.MustConditionf(run == nil || ct.Example != "", "must provide a value for Example")
}
