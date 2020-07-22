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

package preheat

import (
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	p "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JobTestSuite is test suite of preheating job.
type JobTestSuite struct {
	suite.Suite

	dragonfly *httptest.Server
	kraken    *httptest.Server

	context job.Context

	preheatingImage *provider.PreheatImage
}

// TestJob is the entry method of JobTestSuite
func TestJob(t *testing.T) {
	suite.Run(t, &JobTestSuite{})
}

// SetupSuite prepares the env for JobTestSuite.
func (suite *JobTestSuite) SetupSuite() {
	suite.dragonfly = provider.MockDragonflyProvider()
	suite.dragonfly.StartTLS()

	suite.kraken = provider.MockKrakenProvider()
	suite.kraken.StartTLS()

	suite.preheatingImage = &provider.PreheatImage{
		Type:      "image",
		ImageName: "busybox",
		Tag:       "latest",
		URL:       "https://harbor.com",
		Headers: map[string]interface{}{
			"robot$my": "jwt-token",
		},
	}

	ctx := &jobservice.MockJobContext{}
	logger := &jobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	suite.context = ctx
}

// TearDownSuite clears the env for JobTestSuite.
func (suite *JobTestSuite) TearDownSuite() {
	suite.dragonfly.Close()
	suite.kraken.Close()
}

// TestJobWithDragonflyDriver test preheat job running with Dragonfly driver.
func (suite *JobTestSuite) TestJobWithDragonflyDriver() {
	ins := &p.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   provider.DriverDragonfly,
		Endpoint: suite.dragonfly.URL,
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   provider.DriverStatusHealthy,
	}

	suite.runJob(ins)
}

// TestJobWithKrakenDriver test preheat job running with Kraken driver.
func (suite *JobTestSuite) TestJobWithKrakenDriver() {
	ins := &p.Instance{
		ID:       2,
		Name:     "test-instance2",
		Vendor:   provider.DriverKraken,
		Endpoint: suite.kraken.URL,
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   provider.DriverStatusHealthy,
	}

	suite.runJob(ins)
}

func (suite *JobTestSuite) validateJob(j job.Interface, params job.Parameters) {
	require.Equal(suite.T(), uint(1), j.MaxFails(), "max fails")
	require.Equal(suite.T(), false, j.ShouldRetry(), "should retry")
	require.Equal(suite.T(), uint(0), j.MaxCurrency(), "max concurrency")
	require.NoError(suite.T(), j.Validate(params), "validate job parameters")
}

func (suite *JobTestSuite) runJob(ins *p.Instance) {
	params := make(job.Parameters)
	data, err := ins.ToJSON()
	require.NoError(suite.T(), err, "encode parameter", PreheatParamProvider)
	params[PreheatParamProvider] = data

	data, err = suite.preheatingImage.ToJSON()
	require.NoError(suite.T(), err, "encode parameter", PreheatParamImage)
	params[PreheatParamImage] = data

	j := &Job{}
	suite.validateJob(j, params)
	err = j.Run(suite.context, params)
	suite.NoError(err, "run preheating job with driver %s", ins.Vendor)
}
