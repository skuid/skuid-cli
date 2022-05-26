package cmd

import (
	"github.com/spf13/cobra"

	"github.com/skuid/tides/pkg/flags"
)

var streamCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "stream",
	Short:         "Stream between Tides and Marina.",
	Long:          "Stream between Tides and Marina.",
	// PersistentPreRunE: common.PrerunValidation,
	RunE: Stream,
}

func init() {
	TidesCmd.AddCommand(streamCmd)
	flags.AddFlags(streamCmd, flags.MarinaHost, flags.PlinyHost)
}

func Stream(cmd *cobra.Command, _ []string) error {
	return nil
}
