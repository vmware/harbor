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

package scanner

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	mocktesting "github.com/goharbor/harbor/src/testing/mock"
	v1testing "github.com/goharbor/harbor/src/testing/pkg/scan/rest/v1"
	scannertesting "github.com/goharbor/harbor/src/testing/pkg/scan/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ControllerTestSuite is test suite to test the basic api controller.
type ControllerTestSuite struct {
	suite.Suite

	c     *basicController
	mMgr  *scannertesting.Manager
	mMeta *MockProMetaManager

	sample *scanner.Registration
}

// TestController is the entry of controller test suite
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupTest prepares env for the controller test suite
func (suite *ControllerTestSuite) SetupTest() {
	suite.mMgr = &scannertesting.Manager{}
	suite.mMeta = new(MockProMetaManager)

	m := &v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name:    "Clair",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Capabilities: []*v1.ScannerCapability{{
			ConsumesMimeTypes: []string{
				v1.MimeTypeOCIArtifact,
				v1.MimeTypeDockerArtifact,
			},
			ProducesMimeTypes: []string{
				v1.MimeTypeNativeReport,
				v1.MimeTypeRawReport,
			},
		}},
		Properties: v1.ScannerProperties{
			"extra": "testing",
		},
	}

	suite.sample = &scanner.Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	mc := &v1testing.Client{}
	mc.On("GetMetadata").Return(m, nil)

	mcp := &v1testing.ClientPool{}
	mocktesting.OnAnything(mcp, "Get").Return(mc, nil)
	suite.c = &basicController{
		manager:    suite.mMgr,
		proMetaMgr: suite.mMeta,
		clientPool: mcp,
	}
}

// Clear test case
func (suite *ControllerTestSuite) TearDownTest() {
	suite.sample.UUID = ""
}

// TestListRegistrations tests ListRegistrations
func (suite *ControllerTestSuite) TestListRegistrations() {
	query := &q.Query{
		PageSize:   10,
		PageNumber: 1,
	}

	suite.sample.UUID = "uuid"
	l := []*scanner.Registration{suite.sample}

	suite.mMgr.On("List", query).Return(l, nil)

	rl, err := suite.c.ListRegistrations(query)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(rl))
}

// TestCreateRegistration tests CreateRegistration
func (suite *ControllerTestSuite) TestCreateRegistration() {
	suite.mMgr.On("Create", suite.sample).Return("uuid", nil)

	uid, err := suite.mMgr.Create(suite.sample)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), uid, "uuid")
}

// TestGetRegistration tests GetRegistration
func (suite *ControllerTestSuite) TestGetRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)

	rr, err := suite.c.GetRegistration("uuid")
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), rr)
	assert.Equal(suite.T(), "forUT", rr.Name)
}

// TestRegistrationExists tests RegistrationExists
func (suite *ControllerTestSuite) TestRegistrationExists() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)

	exists := suite.c.RegistrationExists("uuid")
	assert.Equal(suite.T(), true, exists)

	suite.mMgr.On("Get", "uuid2").Return(nil, nil)

	exists = suite.c.RegistrationExists("uuid2")
	assert.Equal(suite.T(), false, exists)
}

// TestUpdateRegistration tests UpdateRegistration
func (suite *ControllerTestSuite) TestUpdateRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Update", suite.sample).Return(nil)

	err := suite.c.UpdateRegistration(suite.sample)
	require.NoError(suite.T(), err)
}

// TestDeleteRegistration tests DeleteRegistration
func (suite *ControllerTestSuite) TestDeleteRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)
	suite.mMgr.On("Delete", "uuid").Return(nil)

	r, err := suite.c.DeleteRegistration("uuid")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestSetDefaultRegistration tests SetDefaultRegistration
func (suite *ControllerTestSuite) TestSetDefaultRegistration() {
	suite.mMgr.On("SetAsDefault", "uuid").Return(nil)

	err := suite.c.SetDefaultRegistration("uuid")
	require.NoError(suite.T(), err)
}

// TestSetRegistrationByProject tests SetRegistrationByProject
func (suite *ControllerTestSuite) TestSetRegistrationByProject() {
	m := make(map[string]string, 1)
	mm := make(map[string]string, 1)
	mmm := make(map[string]string, 1)
	mm[proScannerMetaKey] = "uuid"
	mmm[proScannerMetaKey] = "uuid2"

	var pid, pid2 int64 = 1, 2

	// not set before
	suite.mMeta.On("Get", pid, []string{proScannerMetaKey}).Return(m, nil)
	suite.mMeta.On("Add", pid, mm).Return(nil)

	err := suite.c.SetRegistrationByProject(pid, "uuid")
	require.NoError(suite.T(), err)

	// Set before
	suite.mMeta.On("Get", pid2, []string{proScannerMetaKey}).Return(mm, nil)
	suite.mMeta.On("Update", pid2, mmm).Return(nil)

	err = suite.c.SetRegistrationByProject(pid2, "uuid2")
	require.NoError(suite.T(), err)
}

// TestGetRegistrationByProject tests GetRegistrationByProject
func (suite *ControllerTestSuite) TestGetRegistrationByProject() {
	m := make(map[string]string, 1)
	m[proScannerMetaKey] = "uuid"

	// Configured at project level
	var pid int64 = 1
	suite.sample.UUID = "uuid"

	suite.mMeta.On("Get", pid, []string{proScannerMetaKey}).Return(m, nil)
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)

	r, err := suite.c.GetRegistrationByProject(pid)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "forUT", r.Name)

	// Not configured at project level, return system default
	suite.mMeta.On("Get", pid, []string{proScannerMetaKey}).Return(nil, nil)
	suite.mMgr.On("GetDefault").Return(suite.sample, nil)

	r, err = suite.c.GetRegistrationByProject(pid)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestGetRegistrationByProjectWhenPingError tests GetRegistrationByProject
func (suite *ControllerTestSuite) TestGetRegistrationByProjectWhenPingError() {
	m := make(map[string]string, 1)
	m[proScannerMetaKey] = "uuid"

	// Configured at project level
	var pid int64 = 1
	suite.sample.UUID = "uuid"

	suite.mMeta.On("Get", pid, []string{proScannerMetaKey}).Return(m, nil)
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)

	// Ping error
	mc := &v1testing.Client{}
	mc.On("GetMetadata").Return(nil, fmt.Errorf("getMetadata error"))

	mcp := &v1testing.ClientPool{}
	mocktesting.OnAnything(mcp, "Get").Return(mc, nil)
	suite.c.clientPool = mcp

	r, err := suite.c.GetRegistrationByProject(pid)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unhealthy", r.Health)
}

// TestPing ...
func (suite *ControllerTestSuite) TestPing() {
	meta, err := suite.c.Ping(suite.sample)
	require.NoError(suite.T(), err)
	suite.NotNil(meta)
}

// TestGetMetadata ...
func (suite *ControllerTestSuite) TestGetMetadata() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", "uuid").Return(suite.sample, nil)

	meta, err := suite.c.GetMetadata(suite.sample.UUID)
	require.NoError(suite.T(), err)
	suite.NotNil(meta)
	suite.Equal(1, len(meta.Capabilities))
}

// MockProMetaManager is the mock of the ProjectMetadataManager
type MockProMetaManager struct {
	mock.Mock
}

// Add ...
func (m *MockProMetaManager) Add(projectID int64, meta map[string]string) error {
	args := m.Called(projectID, meta)
	return args.Error(0)
}

// Delete ...
func (m *MockProMetaManager) Delete(projecdtID int64, meta ...string) error {
	args := m.Called(projecdtID, meta)
	return args.Error(0)
}

// Update ...
func (m *MockProMetaManager) Update(projectID int64, meta map[string]string) error {
	args := m.Called(projectID, meta)
	return args.Error(0)
}

// Get ...
func (m *MockProMetaManager) Get(projectID int64, meta ...string) (map[string]string, error) {
	args := m.Called(projectID, meta)
	return args.Get(0).(map[string]string), args.Error(1)
}

// List ...
func (m *MockProMetaManager) List(name, value string) ([]*models.ProjectMetadata, error) {
	args := m.Called(name, value)
	return args.Get(0).([]*models.ProjectMetadata), args.Error(1)
}
