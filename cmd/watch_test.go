package cmd_test

import (
	"testing"

	"github.com/skuid/skuid-cli/cmd"
	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdWatch(t *testing.T) {
	factory := &cmdutil.Factory{}
	c := cmd.NewCmdWatch(factory)
	assert.NotNil(t, c)
}

// TODO: Add Tests for command execution once code issues regarding testability (e.g., using
// dependency injection for things like HttpClient, Logger) are addressed - see https://github.com/skuid/skuid-cli/issues/166
