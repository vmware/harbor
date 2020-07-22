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

package scheduler

import (
	"context"
	"testing"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/stretchr/testify/suite"
)

type daoTestSuite struct {
	suite.Suite
	dao     DAO
	execMgr task.ExecutionManager
	ctx     context.Context
	id      int64
	execID  int64
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = &dao{}
	d.execMgr = task.NewExecutionManager()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.Context()
}

func (d *daoTestSuite) SetupTest() {
	execID, err := d.execMgr.Create(d.ctx, "vendor", 0, "trigger")
	d.Require().Nil(err)
	d.execID = execID
	schedule := &schedule{
		CRON:              "0 * * * * *",
		ExecutionID:       execID,
		CallbackFuncName:  "callback_func_01",
		CallbackFuncParam: "callback_func_params",
	}
	id, err := d.dao.Create(d.ctx, schedule)
	d.Require().Nil(err)
	d.id = id
}

func (d *daoTestSuite) TearDownTest() {
	d.Require().Nil(d.dao.Delete(d.ctx, d.id))
	d.Require().Nil(d.execMgr.Delete(d.ctx, d.execID))
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass is covered in SetupTest

	// foreign key error
	_, err := d.dao.Create(d.ctx, &schedule{
		CRON:             "0 * * * * *",
		ExecutionID:      10000,
		CallbackFuncName: "callback_func",
	})
	d.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))
}

func (d *daoTestSuite) TestList() {
	schedules, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"CallbackFuncName": "callback_func_01",
		},
	})
	d.Require().Nil(err)
	d.Require().Len(schedules, 1)
	d.Equal(d.id, schedules[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// not found
	schedule, err := d.dao.Get(d.ctx, 10000)
	d.True(errors.IsNotFoundErr(err))

	// pass
	schedule, err = d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Equal(d.id, schedule.ID)
}

func (d *daoTestSuite) TestDelete() {
	// the happy pass is covered in TearDownTest

	// not found
	err := d.dao.Delete(d.ctx, 10000)
	d.True(errors.IsNotFoundErr(err))
}

func (d *daoTestSuite) TestUpdate() {
	// not found
	err := d.dao.Update(d.ctx, &schedule{
		ID: 10000,
	})
	d.True(errors.IsNotFoundErr(err))

	// pass
	err = d.dao.Update(d.ctx, &schedule{
		ID:   d.id,
		CRON: "* */2 * * * *",
	}, "CRON")
	d.Require().Nil(err)

	schedule, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Equal("* */2 * * * *", schedule.CRON)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
