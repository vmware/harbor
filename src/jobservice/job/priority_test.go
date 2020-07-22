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

package job

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PrioritySamplerSuite is test suite for PrioritySampler.
type PrioritySamplerSuite struct {
	suite.Suite

	sampler *defaultSampler
}

// TestPrioritySampler is entry point of PrioritySamplerSuite.
func TestPrioritySampler(t *testing.T) {
	suite.Run(t, &PrioritySamplerSuite{})
}

// SetupSuite prepares the testing env
func (suite *PrioritySamplerSuite) SetupSuite() {
	suite.sampler = &defaultSampler{}
}

// Test for method
func (suite *PrioritySamplerSuite) Test() {
	p1 := suite.sampler.For(SampleJob)
	suite.Equal((uint)(1), p1, "Job priority for %s", SampleJob)

	p2 := suite.sampler.For(Retention)
	suite.Equal(defaultPriority, p2, "Job priority for %s", Retention)

	p3 := suite.sampler.For(Replication)
	suite.Equal(defaultPriority, p3, "Job priority for %s", Replication)

	p4 := suite.sampler.For(SlackJob)
	suite.Equal((uint)(1), p4, "Job priority for %s", SlackJob)
}
