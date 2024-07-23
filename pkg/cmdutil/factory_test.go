package cmdutil_test

import (
	"testing"

	"github.com/skuid/skuid-cli/pkg/cmdutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type NewFactoryTestSuite struct {
	suite.Suite
}

func (suite *NewFactoryTestSuite) TestNewFactory() {
	t := suite.T()

	appVersion := "1.2.30"
	f := cmdutil.NewFactory(appVersion)
	assert.NotNil(t, f)
	assert.Equal(t, appVersion, f.AppVersion)
	assert.NotNil(t, f.LogConfig)
	assert.NotNil(t, f.Commander)
}

func TestNewFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(NewFactoryTestSuite))
}
