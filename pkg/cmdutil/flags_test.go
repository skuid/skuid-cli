package cmdutil_test

import (
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagManagerTestSuite struct {
	suite.Suite
}

func (suite *FlagManagerTestSuite) TestTrackFlag() {
	t := suite.T()
	pf := &pflag.Flag{Name: "foobar"}
	fi := &flags.Flag[string]{Name: "foobar", Required: true, LegacyEnvVars: []string{"legacy1", "legacy2"}}
	mgr := &cmdutil.FlagManager{}

	mgr.Track(pf, fi)
	actualFlag, ok := mgr.Find(pf)

	assert.True(t, ok)
	assert.NotNil(t, actualFlag)
	assert.Same(t, fi, actualFlag)
}

func (suite *FlagManagerTestSuite) TestFindNotFound() {
	t := suite.T()
	pf := &pflag.Flag{Name: "foobar"}
	mgr := &cmdutil.FlagManager{}

	actualFlag, ok := mgr.Find(pf)

	assert.False(t, ok)
	assert.Nil(t, actualFlag)
}

func (suite *FlagManagerTestSuite) TestNewFlagManager() {
	t := suite.T()
	mgr := cmdutil.NewFlagManager()
	assert.NotNil(t, mgr)
}

func TestFlagManagerTestSuite(t *testing.T) {
	suite.Run(t, new(FlagManagerTestSuite))
}
