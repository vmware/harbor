package instance

import (
	"context"
	"testing"

	beego_orm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	models "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultInstance = &models.Instance{
		ID:             1,
		Name:           "dragonfly-cn-1",
		Description:    "fake dragonfly server",
		Vendor:         "dragonfly",
		Endpoint:       "https://cn-1.dragonfly.com",
		AuthMode:       "basic",
		AuthData:       "{\"username\": \"admin\", \"password\": \"123456\"}",
		Status:         "healthy",
		Enabled:        true,
		SetupTimestamp: 1582721396,
	}
)

type instanceSuite struct {
	suite.Suite
	ctx context.Context
	dao DAO
}

func (is *instanceSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	is.ctx = orm.NewContext(nil, beego_orm.NewOrm())
	is.dao = New()
}

func (is *instanceSuite) SetupTest() {
	t := is.T()
	_, err := is.dao.Create(is.ctx, defaultInstance)
	assert.Nil(t, err)
}

func (is *instanceSuite) TearDownTest() {
	t := is.T()
	err := is.dao.Delete(is.ctx, defaultInstance.ID)
	assert.Nil(t, err)
}

func (is *instanceSuite) TestGet() {
	t := is.T()
	i, err := is.dao.Get(is.ctx, defaultInstance.ID)
	assert.Nil(t, err)
	assert.Equal(t, defaultInstance.Name, i.Name)

	// not exist
	i, err = is.dao.Get(is.ctx, 0)
	assert.Nil(t, i)
}

// TestCreate tests create instance.
func (is *instanceSuite) TestCreate() {
	// test create same name instance, should error
	sameNameInstance := *defaultInstance
	sameNameInstance.ID = 1000
	_, err := is.dao.Create(is.ctx, &sameNameInstance)
	is.True(errors.IsConflictErr(err))
}

// TestGetByName tests get a instance by name.
func (is *instanceSuite) TestGetByName() {
	instance, err := is.dao.GetByName(is.ctx, defaultInstance.Name)
	is.Require().Nil(err)
	is.Require().NotNil(instance)
	is.Equal(defaultInstance.Name, instance.Name, "get a default instance")

	// not found
	_, err = is.dao.GetByName(is.ctx, "default-instance")
	is.Require().NotNil(err)
	is.True(errors.IsErr(err, errors.NotFoundCode))
}
func (is *instanceSuite) TestUpdate() {
	t := is.T()
	i, err := is.dao.Get(is.ctx, defaultInstance.ID)
	assert.Nil(t, err)
	assert.NotNil(t, i)

	i.Enabled = false
	err = is.dao.Update(is.ctx, i, "enabled")
	assert.Nil(t, err)

	i.Default = true
	err = is.dao.Update(is.ctx, i, "default")
	assert.NotNil(t, err)

	i, err = is.dao.Get(is.ctx, defaultInstance.ID)
	assert.Nil(t, err)
	assert.NotNil(t, i)
	assert.False(t, i.Enabled)
}

func (is *instanceSuite) TestList() {
	t := is.T()
	// add more instances
	testInstance1 := &models.Instance{
		ID:             2,
		Name:           "kraken-us-1",
		Description:    "fake kraken server",
		Vendor:         "kraken",
		Endpoint:       "https://us-1.kraken.com",
		AuthMode:       "none",
		AuthData:       "",
		Status:         "success",
		Enabled:        true,
		SetupTimestamp: 0,
	}
	_, err := is.dao.Create(is.ctx, testInstance1)
	assert.Nilf(t, err, "Create %d", testInstance1.ID)
	defer func() {
		// clean data
		err = is.dao.Delete(is.ctx, testInstance1.ID)
		assert.Nilf(t, err, "delete instance %d", testInstance1.ID)
	}()

	total, err := is.dao.Count(is.ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, total, int64(2))
	// limit 1
	total, err = is.dao.Count(is.ctx, &q.Query{PageSize: 1, PageNumber: 1})
	assert.Nil(t, err)
	assert.Equal(t, total, int64(2))

	// without limit should return all instances
	instances, err := is.dao.List(is.ctx, nil)
	assert.Nil(t, err)
	assert.Len(t, instances, 2)

	// limit 1
	instances, err = is.dao.List(is.ctx, &q.Query{PageSize: 1, PageNumber: 1})
	assert.Nil(t, err)
	assert.Len(t, instances, 1, "instances number")
	assert.Equal(t, defaultInstance.ID, instances[0].ID)

	// keyword search
	keywords := make(map[string]interface{})
	keywords["name"] = "kraken-us-1"
	instances, err = is.dao.List(is.ctx, &q.Query{Keywords: keywords})
	assert.Nil(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, testInstance1.Name, instances[0].Name)

}

func TestInstance(t *testing.T) {
	suite.Run(t, &instanceSuite{})
}
