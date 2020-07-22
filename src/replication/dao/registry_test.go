package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultRegistry = &models.Registry{
		Name:           "daoTestDefault",
		URL:            "test.harbor.io",
		CredentialType: "basic",
		AccessKey:      "key1",
		AccessSecret:   "secret1",
		Type:           "harbor",
	}
	testRegistry1 = &models.Registry{
		Name:           "daoTest2",
		URL:            "test2.harbor.io",
		CredentialType: "basic",
		AccessKey:      "key1",
		AccessSecret:   "secret1",
		Type:           "harbor",
	}
)

type RegistrySuite struct {
	suite.Suite
	defaultID int64
}

func (suite *RegistrySuite) SetupTest() {
	assert := assert.New(suite.T())
	id, err := AddRegistry(defaultRegistry)
	assert.Nil(err)
	suite.defaultID = id
}

func (suite *RegistrySuite) TearDownTest() {
	assert := assert.New(suite.T())
	err := DeleteRegistry(suite.defaultID)
	assert.Nil(err)
}

func (suite *RegistrySuite) TestGetRegistry() {
	assert := assert.New(suite.T())

	// Get non-existed registry, should fail
	r, _ := GetRegistry(0)
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestGetRegistryByName() {
	assert := assert.New(suite.T())

	// Get registry by empty name, should fail
	r, _ := GetRegistryByName("")
	assert.Nil(r)

	// Get non-existed registry, should fail
	r, _ = GetRegistryByName("non-exist")
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistryByName(defaultRegistry.Name)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestGetRegistryByURL() {
	assert := assert.New(suite.T())

	// Get registry by empty url, should fail
	r, _ := GetRegistryByURL("")
	assert.Nil(r)

	// Get non-existed registry, should fail
	r, _ = GetRegistryByURL("non-exist.harbor.io")
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistryByURL(defaultRegistry.URL)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestListRegistries() {
	ctx := orm.Context()

	// nil query
	count, registries, err := ListRegistries(ctx, nil)
	suite.Require().Nil(err)
	if count < 1 {
		suite.T().Errorf("At least %d should be found in total, but got %d", 1, count)
	}

	// query by name
	count, registries, err = ListRegistries(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Name": "daoTestDefault",
		},
	})
	suite.Require().Nil(err)
	suite.Require().Equal(int64(1), count)
	suite.Assert().Equal("daoTestDefault", registries[0].Name)

}

func (suite *RegistrySuite) TestUpdate() {
	assert := assert.New(suite.T())

	// Get registry, should succeed
	r, err := GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.NotNil(r)

	r.AccessKey = "key2"
	err = UpdateRegistry(r)
	assert.Nil(err)

	r, err = GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.NotNil(r)
	assert.Equal("key2", r.AccessKey)
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
