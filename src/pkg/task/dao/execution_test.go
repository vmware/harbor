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

package dao

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/suite"
)

type executionDAOTestSuite struct {
	suite.Suite
	ctx          context.Context
	executionDAO *executionDAO
	executionID  int64
}

func (e *executionDAOTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
	e.ctx = orm.Context()
	e.executionDAO = &executionDAO{}
}

func (e *executionDAOTestSuite) SetupTest() {
	id, err := e.executionDAO.Create(e.ctx, &Execution{
		VendorType: "test",
		Trigger:    "test",
		ExtraAttrs: "{}",
	})
	e.Require().Nil(err)
	e.executionID = id
}

func (e *executionDAOTestSuite) TearDownTest() {
	err := e.executionDAO.Delete(e.ctx, e.executionID)
	e.Nil(err)
}

func (e *executionDAOTestSuite) TestCount() {
	count, err := e.executionDAO.Count(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": "test",
		},
	})
	e.Require().Nil(err)
	e.Equal(int64(1), count)
}

func (e *executionDAOTestSuite) TestList() {
	executions, err := e.executionDAO.List(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": "test",
		},
	})
	e.Require().Nil(err)
	e.Require().Len(executions, 1)
	e.Equal(e.executionID, executions[0].ID)
}

func (e *executionDAOTestSuite) TestGet() {
	// not exist
	_, err := e.executionDAO.Get(e.ctx, 10000)
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// exist
	execution, err := e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.NotNil(execution)
}

func (e *executionDAOTestSuite) TestCreate() {
	// happy pass is covered by SetupTest
}

func (e *executionDAOTestSuite) TestUpdate() {
	// not exist
	err := e.executionDAO.Update(e.ctx, &Execution{ID: 10000}, "Status")
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// exist
	err = e.executionDAO.Update(e.ctx, &Execution{
		ID:     e.executionID,
		Status: "failed",
	}, "Status")
	e.Require().Nil(err)
	execution, err := e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal("failed", execution.Status)
}

func (e *executionDAOTestSuite) TestDelete() {
	// not exist
	err := e.executionDAO.Delete(e.ctx, 10000)
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// happy pass is covered by TearDownTest
}

func TestExecutionDAOSuite(t *testing.T) {
	suite.Run(t, &executionDAOTestSuite{})
}
