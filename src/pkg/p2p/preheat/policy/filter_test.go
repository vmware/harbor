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

package policy

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FilterTestSuite is a test suite of testing policy filter
type FilterTestSuite struct {
	suite.Suite

	candidates []*selector.Candidate
}

// TestFilter is an entry method of running FilterTestSuite
func TestFilter(t *testing.T) {
	suite.Run(t, &FilterTestSuite{})
}

// SetupSuite prepares env for running FilterTestSuite
func (suite *FilterTestSuite) SetupSuite() {
	suite.candidates = []*selector.Candidate{
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "sub/busybox",
			Tags:                  []string{"prod"},
			Digest:                "sha256@fake",
			Labels:                []string{"prod_ready", "approved"},
			Signatures:            map[string]bool{"prod": true},
			VulnerabilitySeverity: 3, // medium
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "sub/busybox",
			Tags:                  []string{"qa"},
			Digest:                "sha256@fake2",
			Labels:                []string{"prod_ready", "approved"},
			Signatures:            map[string]bool{"qa": true},
			VulnerabilitySeverity: 3, // medium
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "portal",
			Tags:                  []string{"prod"},
			Digest:                "sha256@fake3",
			Labels:                []string{"prod_ready", "approved"},
			Signatures:            map[string]bool{"prod": true},
			VulnerabilitySeverity: 3, // medium
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "sub/busybox",
			Tags:                  []string{"prod2"},
			Digest:                "sha256@fake4",
			Labels:                []string{"prod_ready", "approved"},
			Signatures:            map[string]bool{"prod2": true},
			VulnerabilitySeverity: 5, // critical
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "sub/busybox",
			Tags:                  []string{"prod3"},
			Digest:                "sha256@fake5",
			Labels:                []string{"prod_ready", "approved"},
			Signatures:            map[string]bool{"prod3": false},
			VulnerabilitySeverity: 3, // medium
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "sub/busybox",
			Tags:                  []string{"prod4"},
			Digest:                "sha256@fake6",
			Labels:                []string{"prod_ready"},
			Signatures:            map[string]bool{"prod4": true},
			VulnerabilitySeverity: 3, // medium
		},
		{
			NamespaceID:           1,
			Namespace:             "test",
			Kind:                  "image",
			Repository:            "portal",
			Tags:                  []string{"qa"},
			Digest:                "sha256@fake7",
			Labels:                []string{"staged"},
			Signatures:            map[string]bool{"qa": false},
			VulnerabilitySeverity: 4, // high
		},
	}
}

// TestInvalidFilters tests the invalid filters
func (suite *FilterTestSuite) TestInvalidFilters() {
	p1 := &policy.Schema{
		Filters: []*policy.Filter{
			{
				Type:  policy.FilterTypeRepository,
				Value: 100,
			},
		},
	}
	fl := NewFilter()
	_, err := fl.BuildFrom(p1).Filter(suite.candidates)
	suite.Errorf(err, "invalid filter: %s", policy.FilterTypeRepository)

	p2 := &policy.Schema{
		Filters: []*policy.Filter{
			{
				Type:  policy.FilterTypeSignature,
				Value: "true",
			},
		},
	}
	_, err = fl.BuildFrom(p2).Filter(suite.candidates)
	suite.Errorf(err, "invalid filter: %s", policy.FilterTypeSignature)

	p3 := &policy.Schema{
		Filters: []*policy.Filter{
			{
				Type:  policy.FilterTypeVulnerability,
				Value: "3",
			},
		},
	}
	_, err = fl.BuildFrom(p3).Filter(suite.candidates)
	suite.Errorf(err, "invalid filter: %s", policy.FilterTypeVulnerability)
}

// TestFilters test all the supported filters with candidates
func (suite *FilterTestSuite) TestFilters() {
	p := &policy.Schema{
		Filters: []*policy.Filter{
			{
				Type:  policy.FilterTypeRepository,
				Value: "sub/**",
			},
			{
				Type:  policy.FilterTypeTag,
				Value: "prod*",
			},
			{
				Type:  policy.FilterTypeLabel,
				Value: "prod_ready,approved",
			},
			{
				Type:  policy.FilterTypeSignature,
				Value: true, // signed
			},
			{
				Type:  policy.FilterTypeVulnerability,
				Value: 4, // < high
			},
		},
	}

	res, err := NewFilter().BuildFrom(p).Filter(suite.candidates)
	require.NoError(suite.T(), err, "do filters")
	require.Equal(suite.T(), 1, len(res), "number of matched candidates")
	suite.Equal("sha256@fake", res[0].Digest, "digest of matched candidate")
}
