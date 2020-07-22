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
	"context"
	"testing"
	"time"

	beego_orm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	"github.com/stretchr/testify/suite"
)

type daoTestSuite struct {
	suite.Suite

	dao           DAO
	ctx           context.Context
	defaultPolicy *policy.Schema
}

// TestDaoTestSuite tests policy dao.
func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}

// SetupSuite setups testing env.
func (d *daoTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	d.dao = New()
	d.ctx = orm.NewContext(nil, beego_orm.NewOrm())
	d.defaultPolicy = &policy.Schema{
		ID:          1,
		Name:        "default-policy",
		Description: "test",
		ProjectID:   1,
		ProviderID:  1,
		Filters:     nil,
		FiltersStr:  "",
		Trigger:     nil,
		TriggerStr:  "",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedTime: time.Now(),
	}

	_, err := d.dao.Create(d.ctx, d.defaultPolicy)
	d.Require().Nil(err)
}

// TearDownTest cleans testing env.
func (d *daoTestSuite) TearDownSuite() {
	err := d.dao.Delete(d.ctx, d.defaultPolicy.ID)
	d.Require().Nil(err)
}

// TestCount tests count total
func (d *daoTestSuite) TestCount() {
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

// TestCreate tests create a policy schema.
func (d *daoTestSuite) TestCreate() {
	// create duplicate policy should return error
	_, err := d.dao.Create(d.ctx, d.defaultPolicy)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))

	// same name and project id should error
	sameNamePolicy := *d.defaultPolicy
	sameNamePolicy.ID = 1000
	_, err = d.dao.Create(d.ctx, &sameNamePolicy)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))

	// same name but different project id should not error
	sameNamePolicyWithDiffProjectID := sameNamePolicy
	sameNamePolicyWithDiffProjectID.ProjectID = 10
	_, err = d.dao.Create(d.ctx, &sameNamePolicyWithDiffProjectID)
	d.Require().Nil(err)
	// clean
	err = d.dao.Delete(d.ctx, sameNamePolicyWithDiffProjectID.ID)
	d.Require().Nil(err)
}

// Delete tests delete a policy schema.
func (d *daoTestSuite) TestDelete() {
	// delete a not exist policy
	err := d.dao.Delete(d.ctx, 0)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

// Get tests get a policy schema by id.
func (d *daoTestSuite) TestGet() {
	policy, err := d.dao.Get(d.ctx, 1)
	d.Require().Nil(err)
	d.Require().NotNil(policy)
	d.Equal(d.defaultPolicy.Name, policy.Name, "get a default policy")

	// not found
	_, err = d.dao.Get(d.ctx, 1000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

// GetByName tests get a policy schema by name.
func (d *daoTestSuite) TestGetByName() {
	policy, err := d.dao.GetByName(d.ctx, 1, "default-policy")
	d.Require().Nil(err)
	d.Require().NotNil(policy)
	d.Equal(d.defaultPolicy.Name, policy.Name, "get a default policy")

	// not found
	_, err = d.dao.GetByName(d.ctx, 2, "default-policy")
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

// Update tests update a policy schema.
func (d *daoTestSuite) TestUpdate() {
	newDesc := "test update"
	newPolicy := *d.defaultPolicy
	newPolicy.Description = newDesc

	err := d.dao.Update(d.ctx, &newPolicy)
	d.Require().Nil(err)

	policy, err := d.dao.Get(d.ctx, 1)
	d.Require().Nil(err)
	d.Require().NotNil(policy)
	d.Equal(newDesc, policy.Description, "update a policy description")
}

func (d *daoTestSuite) TestList() {
	newPolicy := &policy.Schema{
		ID:          2,
		Name:        "new-policy",
		Description: "new",
		ProjectID:   2,
		ProviderID:  2,
		Filters:     nil,
		FiltersStr:  "",
		Trigger:     nil,
		TriggerStr:  "",
		Enabled:     false,
		CreatedAt:   time.Time{},
		UpdatedTime: time.Time{},
	}

	_, err := d.dao.Create(d.ctx, newPolicy)
	d.Require().Nil(err)
	// clean up
	defer func() {
		err = d.dao.Delete(d.ctx, 2)
		d.Require().Nil(err)
	}()

	policies, err := d.dao.List(d.ctx, &q.Query{})
	d.Require().Nil(err)
	d.Len(policies, 2, "list all policy schemas")

	// list policy filter by project
	query := &q.Query{
		Keywords: map[string]interface{}{
			"project_id": 1,
		},
	}
	policies, err = d.dao.List(d.ctx, query)
	d.Require().Nil(err)
	d.Len(policies, 1, "list policy schemas by project")
	d.Equal(d.defaultPolicy.Name, policies[0].Name)
}
