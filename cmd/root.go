package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/constants"
	"github.com/skuid/skuid-cli/pkg/flags"
)

const VERSION_TEMPLATE = "Skuid CLI Version %v"

func NewCmdRoot(factory *cmdutil.Factory) *cobra.Command {
	rootTemplate := &cmdutil.CmdTemplate{
		Use:   constants.PROJECT_NAME,
		Short: "A CLI for Skuid APIs",
		Long:  `A command-line interface used to retrieve and deploy Skuid NLX sites.`,
		Flags: &cmdutil.CommandFlags{
			Bool:   []*flags.Flag[bool]{flags.Verbose, flags.Trace, flags.FileLogging, flags.Diagnostic},
			String: []*flags.Flag[string]{flags.LogDirectory},
		},
		Commands: []*cobra.Command{
			NewCmdDeploy(factory),
			NewCmdRetrieve(factory),
			NewCmdWatch(factory),
		},
	}

	cmd := rootTemplate.ToCommand(factory, nil, nil, nil)
	cmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	cmd.SilenceErrors = true
	cmd.Version = factory.AppVersion
	cmd.SetVersionTemplate(fmt.Sprintf(VERSION_TEMPLATE+"\n", factory.AppVersion))

	return cmd
}
