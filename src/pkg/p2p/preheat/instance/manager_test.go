package instance

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
	dao "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	providerModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type fakeDao struct {
	mock.Mock
}

var _ dao.DAO = (*fakeDao)(nil)
var lists = []*providerModel.Instance{
	{Name: "abc"},
}

func (d *fakeDao) Create(ctx context.Context, instance *provider.Instance) (int64, error) {
	var args = d.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (d *fakeDao) Get(ctx context.Context, id int64) (*provider.Instance, error) {
	var args = d.Called()
	var instance *provider.Instance
	if args.Get(0) != nil {
		instance = args.Get(0).(*provider.Instance)
	}
	return instance, args.Error(1)
}

func (d *fakeDao) GetByName(ctx context.Context, name string) (*provider.Instance, error) {
	var args = d.Called()
	var instance *provider.Instance
	if args.Get(0) != nil {
		instance = args.Get(0).(*provider.Instance)
	}
	return instance, args.Error(1)
}

func (d *fakeDao) Update(ctx context.Context, instance *provider.Instance, props ...string) error {
	var args = d.Called()
	return args.Error(0)
}

func (d *fakeDao) Delete(ctx context.Context, id int64) error {
	var args = d.Called()
	return args.Error(0)
}

func (d *fakeDao) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	var args = d.Called()

	return int64(args.Int(0)), args.Error(1)
}

func (d *fakeDao) List(ctx context.Context, query *q.Query) (ins []*provider.Instance, err error) {
	var args = d.Called()
	var instances []*provider.Instance
	if args.Get(0) != nil {
		instances = args.Get(0).([]*provider.Instance)
	}
	return instances, args.Error(1)
}

type instanceManagerSuite struct {
	suite.Suite
	dao     *fakeDao
	ctx     context.Context
	manager Manager
}

func (im *instanceManagerSuite) SetupSuite() {
	im.dao = &fakeDao{}
	im.manager = &manager{dao: im.dao}
	im.dao.On("List").Return(lists, nil)
}

func (im *instanceManagerSuite) TestSave() {
	im.dao.On("Create").Return(1, nil)
	id, err := im.manager.Save(im.ctx, nil)
	im.Require().Nil(err)
	im.Require().Equal(int64(1), id)
}

func (im *instanceManagerSuite) TestDelete() {
	im.dao.On("Delete").Return(nil)
	err := im.manager.Delete(im.ctx, 1)
	im.Require().Nil(err)
}

func (im *instanceManagerSuite) TestUpdate() {
	im.dao.On("Update").Return(nil)
	err := im.manager.Update(im.ctx, nil)
	im.Require().Nil(err)
}

func (im *instanceManagerSuite) TestGet() {
	ins := &providerModel.Instance{Name: "abc"}
	im.dao.On("Get").Return(ins, nil)
	res, err := im.manager.Get(im.ctx, 1)
	im.Require().Nil(err)
	im.Require().Equal(ins, res)
}

func (im *instanceManagerSuite) TestGetByName() {
	im.dao.On("GetByName").Return(lists[0], nil)
	res, err := im.manager.GetByName(im.ctx, "abc")
	im.Require().Nil(err)
	im.Require().Equal(lists[0], res)
}

func (im *instanceManagerSuite) TestCount() {
	im.dao.On("Count").Return(2, nil)
	count, err := im.manager.Count(im.ctx, nil)
	assert.Nil(im.T(), err)
	assert.Equal(im.T(), int64(2), count)
}

func (im *instanceManagerSuite) TestList() {
	lists := []*providerModel.Instance{
		{Name: "abc"},
	}
	im.dao.On("List").Return(lists, nil)
	res, err := im.manager.List(im.ctx, nil)
	assert.Nil(im.T(), err)
	assert.Len(im.T(), res, 1)
	assert.Equal(im.T(), lists, res)
}

func TestInstanceManager(t *testing.T) {
	suite.Run(t, &instanceManagerSuite{})
}
