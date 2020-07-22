package artifactrash

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"time"
)

type fakeDao struct {
	mock.Mock
}

func (f *fakeDao) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) (err error) {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) Filter(ctx context.Context, timeWindow time.Time) (arts []model.ArtifactTrash, err error) {
	args := f.Called()
	return args.Get(0).([]model.ArtifactTrash), args.Error(1)
}
func (f *fakeDao) Flush(ctx context.Context, timeWindow time.Time) (err error) {
	args := f.Called()
	return args.Error(0)
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

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything).Return(1, nil)
	id, err := m.mgr.Create(nil, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
		Digest:            "5678",
	})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Equal(int64(1), id)
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything).Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestFilter() {
	m.dao.On("Filter", mock.Anything).Return([]model.ArtifactTrash{
		{
			ManifestMediaType: v1.MediaTypeImageManifest,
			RepositoryName:    "test/hello-world",
			Digest:            "5678",
		},
	}, nil)
	arts, err := m.mgr.Filter(nil, 0)
	m.Require().Nil(err)
	m.Equal(len(arts), 1)
}

func (m *managerTestSuite) TestFlush() {
	m.dao.On("Flush", mock.Anything).Return(nil)
	err := m.mgr.Flush(nil, 0)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}
