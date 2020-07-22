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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	car "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/selector"
	ar "github.com/goharbor/harbor/src/pkg/artifact"
	po "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	pr "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	ta "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/instance"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/policy"
	"github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EnforcerTestSuite is a test suite of testing preheat enforcer
type EnforcerTestSuite struct {
	suite.Suite

	enforcer *defaultEnforcer
	server   *httptest.Server
}

// TestEnforcer is an entry method of running EnforcerTestSuite
func TestEnforcer(t *testing.T) {
	suite.Run(t, &EnforcerTestSuite{})
}

// SetupSuite prepares env for running EnforcerTestSuite
func (suite *EnforcerTestSuite) SetupSuite() {
	// Start mock server
	suite.server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	suite.server.StartTLS()

	fakePolicies := mockPolicies()
	fakePolicyManager := &policy.FakeManager{}
	fakePolicyManager.On("Get",
		context.TODO(),
		mock.AnythingOfType("int64")).
		Return(fakePolicies[0], nil)
	fakePolicyManager.On("ListPoliciesByProject",
		context.TODO(),
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("*q.Query"),
	).Return(fakePolicies, nil)

	fakeExecManager := &task.FakeExecutionManager{}
	fakeExecManager.On("Create",
		context.TODO(),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("map[string]interface {}"),
	).Return(time.Now().Unix(), nil)

	fakeTaskManager := &task.FakeManager{}
	fakeTaskManager.On("Create",
		context.TODO(),
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("*task.Job"),
		mock.AnythingOfType("map[string]interface {}"),
	).Return(time.Now().Unix(), nil)

	fakeArtCtl := &artifact.Controller{}
	fakeArtCtl.On("List",
		context.TODO(),
		mock.AnythingOfType("*q.Query"),
		mock.AnythingOfType("*artifact.Option"),
	).Return(mockArtifacts(), nil)

	fakeScanCtl := &scan.Controller{}
	fakeScanCtl.On("GetSummary",
		context.TODO(),
		mock.AnythingOfType("*artifact.Artifact"),
		[]string{v1.MimeTypeNativeReport},
		mock.AnythingOfType("report.Option"),
	).Return(mockVulnerabilitySummary(), nil)

	fakeProCtl := &project.Controller{}
	fakeProCtl.On("Get",
		context.TODO(),
		(int64)(1),
		mock.Anything,
		mock.Anything,
	).Return(&models.Project{
		ProjectID:    1,
		Name:         "library",
		CVEAllowlist: models.CVEAllowlist{},
		Metadata: map[string]string{
			proMetaKeyContentTrust:  "true",
			proMetaKeyVulnerability: "true",
			proMetaKeySeverity:      "low",
		},
	}, nil)

	fakeInstanceMgr := &instance.FakeManager{}
	fakeInstanceMgr.On("Get",
		context.TODO(),
		mock.AnythingOfType("int64"),
	).Return(&pr.Instance{
		ID:       1,
		Name:     "my_preheat_provider1",
		Vendor:   provider.DriverKraken,
		Endpoint: suite.server.URL,
		Status:   provider.DriverStatusHealthy,
		AuthMode: auth.AuthModeNone,
		Insecure: true,
	}, nil)

	suite.enforcer = &defaultEnforcer{
		policyMgr:    fakePolicyManager,
		executionMgr: fakeExecManager,
		taskMgr:      fakeTaskManager,
		artCtl:       fakeArtCtl,
		scanCtl:      fakeScanCtl,
		proCtl:       fakeProCtl,
		instMgr:      fakeInstanceMgr,
		fullURLGetter: func(c *selector.Candidate) (s string, e error) {
			r := fmt.Sprintf("%s/%s", c.Namespace, c.Repository)
			return fmt.Sprintf(manifestAPIPattern, "https://testing.harbor.com", r, c.Tags[0]), nil
		},
		credMaker: func(c *selector.Candidate) (s string, e error) {
			return "fake-token", nil
		},
	}
}

// TearDownSuite cleans the testing env
func (suite *EnforcerTestSuite) TearDownSuite() {
	suite.server.Close()
}

// TestEnforcePolicy tests the policy enforcement case.
func (suite *EnforcerTestSuite) TestEnforcePolicy() {
	eid, err := suite.enforcer.EnforcePolicy(context.TODO(), 1)
	require.NoError(suite.T(), err, "enforce policy")
	suite.Condition(func() (success bool) {
		return eid > 0
	}, "execution created")
}

// TestPreheatArtifact tests the artifact preheating case
func (suite *EnforcerTestSuite) TestPreheatArtifact() {
	ids, err := suite.enforcer.PreheatArtifact(context.TODO(), mockArtifacts()[1])
	require.NoError(suite.T(), err, "preheat given artifact")
	suite.Equal(1, len(ids), "executions created")
}

// mock policies for reusing
func mockPolicies() []*po.Schema {
	return []*po.Schema{
		{
			ID:          1,
			Name:        "manual_policy",
			Description: "for testing",
			ProjectID:   1,
			ProviderID:  1,
			Filters: []*po.Filter{
				{
					Type:  po.FilterTypeRepository,
					Value: "sub/**",
				},
				{
					Type:  po.FilterTypeTag,
					Value: "prod*",
				},
				{
					Type:  po.FilterTypeLabel,
					Value: "approved,ready",
				},
				{
					Type:  po.FilterTypeSignature,
					Value: false,
				},
				{
					Type:  po.FilterTypeVulnerability,
					Value: 3, // medium
				},
			},
			Trigger: &po.Trigger{
				Type: po.TriggerTypeManual,
			},
			Enabled:     true,
			CreatedAt:   time.Now().UTC(),
			UpdatedTime: time.Now().UTC(),
		}, {
			ID:          2,
			Name:        "event_based_policy",
			Description: "for testing",
			ProjectID:   1,
			ProviderID:  1,
			Filters: []*po.Filter{
				{
					Type:  po.FilterTypeRepository,
					Value: "busy*",
				},
				{
					Type:  po.FilterTypeTag,
					Value: "stage*",
				},
				{
					Type:  po.FilterTypeLabel,
					Value: "staged",
				},
			},
			Trigger: &po.Trigger{
				Type: po.TriggerTypeEventBased,
			},
			Enabled:     true,
			CreatedAt:   time.Now().UTC(),
			UpdatedTime: time.Now().UTC(),
		},
	}
}

// mock artifacts
func mockArtifacts() []*car.Artifact {
	// Skip all the unused properties
	return []*car.Artifact{
		{
			Artifact: ar.Artifact{
				ID:             1,
				Type:           "image",
				ProjectID:      1,
				RepositoryName: "library/sub/busybox",
				Digest:         "sha256@fake1",
			},
			Tags: []*tag.Tag{
				{
					Tag: ta.Tag{
						Name: "prod",
					},
					Signed: true,
				}, {
					Tag: ta.Tag{
						Name: "stage",
					},
					Signed: false,
				},
			},
			Labels: []*models.Label{
				{
					Name: "approved",
				}, {
					Name: "ready",
				},
			},
		}, {
			Artifact: ar.Artifact{
				ID:             2,
				Type:           "image",
				ProjectID:      1,
				RepositoryName: "library/busybox",
				Digest:         "sha256@fake2",
			},
			Tags: []*tag.Tag{
				{
					Tag: ta.Tag{
						Name: "latest",
					},
					Signed: true,
				}, {
					Tag: ta.Tag{
						Name: "stage",
					},
					Signed: false,
				},
			},
			Labels: []*models.Label{
				{
					Name: "approved",
				}, {
					Name: "staged",
				},
			},
		},
	}
}

// mock vulnerability summary
func mockVulnerabilitySummary() map[string]interface{} {
	// skip all unused properties
	return map[string]interface{}{
		v1.MimeTypeNativeReport: &vuln.NativeReportSummary{
			Severity: vuln.Low,
		},
	}
}
