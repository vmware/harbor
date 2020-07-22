// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// KrakenTestSuite is a test suite of testing Kraken driver.
type KrakenTestSuite struct {
	suite.Suite

	kraken *httptest.Server
	driver *KrakenDriver
}

// TestKraken is the entry method of running KrakenTestSuite.
func TestKraken(t *testing.T) {
	suite.Run(t, &KrakenTestSuite{})
}

// SetupSuite prepares the env for KrakenTestSuite.
func (suite *KrakenTestSuite) SetupSuite() {
	suite.kraken = MockKrakenProvider()

	suite.kraken.StartTLS()

	suite.driver = &KrakenDriver{
		instance: &provider.Instance{
			ID:       2,
			Name:     "test-instance2",
			Vendor:   DriverKraken,
			Endpoint: suite.kraken.URL,
			AuthMode: auth.AuthModeNone,
			Enabled:  true,
			Default:  true,
			Insecure: true,
			Status:   DriverStatusHealthy,
		},
	}
}

// TearDownSuite clears the env for KrakenTestSuite.
func (suite *KrakenTestSuite) TearDownSuite() {
	suite.kraken.Close()
}

// TestSelf tests Self method.
func (suite *KrakenTestSuite) TestSelf() {
	m := suite.driver.Self()
	suite.Equal(DriverKraken, m.ID, "self metadata")
}

// TestGetHealth tests GetHealth method.
func (suite *KrakenTestSuite) TestGetHealth() {
	st, err := suite.driver.GetHealth()
	require.NoError(suite.T(), err, "get health")
	suite.Equal(DriverStatusHealthy, st.Status, "healthy status")
}

// TestPreheat tests Preheat method.
func (suite *KrakenTestSuite) TestPreheat() {
	st, err := suite.driver.Preheat(&PreheatImage{
		Type:      "image",
		ImageName: "busybox",
		Digest:    "sha256@fake",
		Tag:       "latest",
		URL:       "https://harbor.com",
	})
	require.NoError(suite.T(), err, "preheat image")
	suite.Equal(provider.PreheatingStatusSuccess, st.Status, "preheat image result")
	suite.NotEmptyf(st.FinishTime, "finish time")
}

// TestCheckProgress tests CheckProgress method.
func (suite *KrakenTestSuite) TestCheckProgress() {
	st, err := suite.driver.CheckProgress("kraken-id")
	require.NoError(suite.T(), err, "get preheat status")
	suite.Equal(provider.PreheatingStatusSuccess, st.Status, "preheat status")
	suite.NotEmptyf(st.FinishTime, "finish time")
}
