package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// KnownDriverTestSuite is a test suite of testing known driver related.
type KnownDriverTestSuite struct {
	suite.Suite
}

// TestKnownDriver is the entry of running KnownDriverTestSuite.
func TestKnownDriver(t *testing.T) {
	suite.Run(t, &KnownDriverTestSuite{})
}

func (suite *KnownDriverTestSuite) TestListProviders() {
	metadata, err := ListProviders()
	require.NoError(suite.T(), err, "list providers")
	suite.Equal(len(knownDrivers), len(metadata))
	suite.Equal(DriverDragonfly, metadata[0].ID)
}

func (suite *KnownDriverTestSuite) TestGetProvider() {
	f, ok := GetProvider(DriverDragonfly)
	require.Equal(suite.T(), true, ok)

	_, err := f(nil)
	suite.NoError(err, "dragonfly factory")
}
