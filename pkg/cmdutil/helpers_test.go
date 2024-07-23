package cmdutil_test

import (
	"bytes"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func emptyCobraRun(*cobra.Command, []string) error              { return nil }
func emptyRun(*cmdutil.Factory, *cobra.Command, []string) error { return nil }
func executeCommand(root *cobra.Command, args ...string) (err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	return root.Execute()
}
