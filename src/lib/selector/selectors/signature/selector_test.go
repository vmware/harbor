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

package signature

import (
	"testing"

	sl "github.com/goharbor/harbor/src/lib/selector"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SignatureSelectorTestSuite is a test suite for testing the signature selector
type SignatureSelectorTestSuite struct {
	suite.Suite

	candidates []*sl.Candidate
}

// TestSignatureSelector is the entry method of running SignatureSelectorTestSuite
func TestSignatureSelector(t *testing.T) {
	suite.Run(t, &SignatureSelectorTestSuite{})
}

// SetupSuite prepares the env for running SeveritySelectorTestSuite
func (suite *SignatureSelectorTestSuite) SetupSuite() {
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
			Signatures: map[string]bool{
				"latest": false,
				"1.0":    true,
			},
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
			Signatures: map[string]bool{
				"latest": false,
				"1.1":    false,
			},
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
			Signatures: map[string]bool{
				"latest": true,
				"1.2":    true,
			},
		},
	}
}

// TestAnySigned tests the 'any' decoration with expected=true
func (suite *SignatureSelectorTestSuite) TestAnySigned() {
	s := New(Any, true, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by signature")
	suite.Equal(2, len(l), "number of matched candidates")
	suite.Equal("busybox", l[0].Repository)
	suite.Equal("portal", l[1].Repository)
}

// TestAnyUnSigned tests the 'any' decoration with expected=false
func (suite *SignatureSelectorTestSuite) TestAnyUnSigned() {
	s := New(Any, false, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by signature")
	suite.Equal(2, len(l), "number of matched candidates")
	suite.Equal("busybox", l[0].Repository)
	suite.Equal("core", l[1].Repository)
}

// TestAllSigned tests the 'all' decoration with expected=true
func (suite *SignatureSelectorTestSuite) TestAllSigned() {
	s := New(All, true, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by signature")
	suite.Equal(1, len(l), "number of matched candidates")
	suite.Equal("portal", l[0].Repository)
}

// TestAllUnSigned tests the 'all' decoration with expected=false
func (suite *SignatureSelectorTestSuite) TestAllUnSigned() {
	s := New(All, false, "")
	l, err := s.Select(suite.candidates)
	require.NoError(suite.T(), err, "filter candidates by signature")
	suite.Equal(1, len(l), "number of matched candidates")
	suite.Equal("core", l[0].Repository)
}
