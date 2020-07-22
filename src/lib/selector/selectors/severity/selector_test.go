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

package severity

import (
	"testing"

	"github.com/stretchr/testify/require"

	sl "github.com/goharbor/harbor/src/lib/selector"
	"github.com/stretchr/testify/suite"
)

// SeveritySelectorTestSuite is a test suite of testing severity selector
type SeveritySelectorTestSuite struct {
	suite.Suite

	candidates []*sl.Candidate
}

// TestSeveritySelector is an entry method of running SeveritySelectorTestSuite
func TestSeveritySelector(t *testing.T) {
	suite.Run(t, &SeveritySelectorTestSuite{})
}

// SetupSuite prepares the env of running SeveritySelectorTestSuite.
func (suite *SeveritySelectorTestSuite) SetupSuite() {
	suite.candidates = []*sl.Candidate{
		{
			Namespace:   "test",
			NamespaceID: 1,
			Repository:  "busybox",
			Kind:        "image",
			Digest:      "sha256@fake",
			Tags: []string{
				"latest",
				"1.0",
			},
			VulnerabilitySeverity: 3, // medium
		}, {
			Namespace:   "test",
			NamespaceID: 1,
			Repository:  "core",
			Kind:        "image",
			Digest:      "sha256@fake",
			Tags: []string{
				"latest",
				"1.1",
			},
			VulnerabilitySeverity: 4, // high
		}, {
			Namespace:   "test",
			NamespaceID: 1,
			Repository:  "portal",
			Kind:        "image",
			Digest:      "sha256@fake",
			Tags: []string{
				"latest",
				"1.2",
			},
			VulnerabilitySeverity: 5, // critical
		},
	}
}

// TestGte test >=
func (suite *SeveritySelectorTestSuite) TestGte() {
	s := New(Gte, 3, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by vulnerability severity")
	suite.Equal(3, len(l), "number of matched candidates")
}

// TestGte test >
func (suite *SeveritySelectorTestSuite) TestGt() {
	s := New(Gt, 3, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by vulnerability severity")
	require.Equal(suite.T(), 2, len(l), "number of matched candidates")
	suite.Condition(func() (success bool) {
		for _, a := range l {
			if a.VulnerabilitySeverity <= 3 {
				return false
			}
		}

		return true
	}, "severity checking of matched candidates")
}

// TestGte test =
func (suite *SeveritySelectorTestSuite) TestEqual() {
	s := New(Equal, 3, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by vulnerability severity")
	require.Equal(suite.T(), 1, len(l), "number of matched candidates")
	suite.Equal("busybox", l[0].Repository, "repository comparison of matched candidate")
}

// TestGte test <=
func (suite *SeveritySelectorTestSuite) TestLte() {
	s := New(Lte, 4, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by vulnerability severity")
	require.Equal(suite.T(), 2, len(l), "number of matched candidates")
	suite.Equal("busybox", l[0].Repository, "repository comparison of matched candidate")
	suite.Equal("core", l[1].Repository, "repository comparison of matched candidate")
}

// TestGte test <
func (suite *SeveritySelectorTestSuite) TestLt() {
	s := New(Lt, 5, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by vulnerability severity")
	require.Equal(suite.T(), 2, len(l), "number of matched candidates")
	suite.Equal("busybox", l[0].Repository, "repository comparison of matched candidate")
	suite.Equal("core", l[1].Repository, "repository comparison of matched candidate")
}
