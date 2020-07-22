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

package label

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fakeDao struct {
	mock.Mock
}

func (f *fakeDao) Get(ctx context.Context, id int64) (*models.Label, error) {
	args := f.Called()
	var label *models.Label
	if args.Get(0) != nil {
		label = args.Get(0).(*models.Label)
	}
	return label, args.Error(1)
}
func (f *fakeDao) Create(ctx context.Context, label *models.Label) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) ListByArtifact(ctx context.Context, artifactID int64) ([]*models.Label, error) {
	args := f.Called()
	var labels []*models.Label
	if args.Get(0) != nil {
		labels = args.Get(0).([]*models.Label)
	}
	return labels, args.Error(1)
}
func (f *fakeDao) CreateReference(ctx context.Context, reference *Reference) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) DeleteReference(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) DeleteReferences(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *fakeDao
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &fakeDao{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestGet() {
	m.dao.On("Get").Return(nil, nil)
	_, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
}

func (m *managerTestSuite) TestListArtifact() {
	m.dao.On("ListByArtifact").Return(nil, nil)
	_, err := m.mgr.ListByArtifact(nil, 1)
	m.Require().Nil(err)
}

func (m *managerTestSuite) TestAddTo() {
	m.dao.On("CreateReference").Return(1, nil)
	err := m.mgr.AddTo(nil, 1, 1)
	m.Require().Nil(err)
}

func (m *managerTestSuite) TestRemoveFrom() {
	// success
	m.dao.On("DeleteReferences").Return(1, nil)
	err := m.mgr.RemoveFrom(nil, 1, 1)
	m.Require().Nil(err)

	// reset mock
	m.SetupTest()

	// not found
	m.dao.On("DeleteReferences").Return(0, nil)
	err = m.mgr.RemoveFrom(nil, 1, 1)
	m.Require().NotNil(err)
	m.True(errors.IsErr(err, errors.NotFoundCode))
}

func (m *managerTestSuite) TestRemoveAllFrom() {
	m.dao.On("DeleteReferences").Return(2, nil)
	err := m.mgr.RemoveAllFrom(nil, 1)
	m.Require().Nil(err)
}

func (m *managerTestSuite) TestRemoveFromAllArtifacts() {
	m.dao.On("DeleteReferences").Return(2, nil)
	err := m.mgr.RemoveFromAllArtifacts(nil, 1)
	m.Require().Nil(err)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
