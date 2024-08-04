package testutil

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

type EnvVar[T any] struct {
	EnvValue string
	Value    T
}

func SetupEnv[T any](t *testing.T, vars map[string]EnvVar[T]) {
	for key, val := range vars {
		t.Setenv(key, val.EnvValue)
	}
}

// testify evaluates pointer equality using referenced value, not pointer itself
// so this can be used to test pointer equality
func SameMatcher(a any) func(fn any) bool {
	return func(fn any) bool {
		return a == fn
	}
}

func EmptyCobraRun(*cobra.Command, []string) error { return nil }

func ExecuteCommand(root *cobra.Command, args ...string) (err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	return root.Execute()
}
