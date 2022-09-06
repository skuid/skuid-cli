package util_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/skuid/tides/pkg/util"
)

func TestFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("a", "b", "c")
	cmd.Flags().Bool("d", true, "e")
	cmd.PersistentFlags().Int("f", 1, "g")
	flags := util.AllFlags(cmd)
	assert.Equal(t, 3, len(flags))
}
