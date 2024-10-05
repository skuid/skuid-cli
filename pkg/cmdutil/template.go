package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/errutil"
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

	cmd := &cobra.Command{
		SilenceUsage: true,
		Use:          ct.Use,
		Short:        ct.Short,
		Long:         ct.Long,
		Example:      ct.Example,
		RunE:         run,
	}

	// cobra's default help command starts with lower-case usage text but our
	// flags start with uppercase so ensure consistent format for the flag
	// help is available on every command
	// we don't use cmdutil helpers because we don't want it registered
	// since its technically not our flag
	cmd.Flags().Bool("help", false, "Show help for "+cmd.Name())
	return cmd
}

func (ct *CmdTemplate) checkValidTemplate(run CobraRunE) {
	// should never happen in production
	errutil.MustConditionf(ct.Use != "", "must provide a value for Use")
	errutil.MustConditionf(ct.Short != "", "must provide a value for Short")
	errutil.MustConditionf(ct.Long != "", "must provide a value for Long")
	errutil.MustConditionf(run == nil || ct.Example != "", "must provide a value for Example")
}
