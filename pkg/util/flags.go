package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func AllFlags(cmd *cobra.Command) (flags []*pflag.Flag) {
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})
	return
}
