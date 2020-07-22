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

package task

import (
	"errors"
	cjob "github.com/goharbor/harbor/src/common/job"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type taskManagerTestSuite struct {
	suite.Suite
	mgr      *manager
	dao      *mockTaskDAO
	jsClient *mockJobserviceClient
}

func (t *taskManagerTestSuite) SetupTest() {
	t.dao = &mockTaskDAO{}
	t.jsClient = &mockJobserviceClient{}
	t.mgr = &manager{
		dao:      t.dao,
		jsClient: t.jsClient,
	}
}

func (t *taskManagerTestSuite) TestCreate() {
	// success to submit job to jobservice
	t.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	t.jsClient.On("SubmitJob", mock.Anything).Return("1", nil)
	t.dao.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	id, err := t.mgr.Create(nil, 1, &Job{}, map[string]interface{}{"a": "b"})
	t.Require().Nil(err)
	t.Equal(int64(1), id)
	t.dao.AssertExpectations(t.T())
	t.jsClient.AssertExpectations(t.T())

	// reset mock
	t.SetupTest()

	// failed to submit job to jobservice
	t.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	t.jsClient.On("SubmitJob", mock.Anything).Return("", errors.New("error"))
	t.dao.On("Update", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	id, err = t.mgr.Create(nil, 1, &Job{}, map[string]interface{}{"a": "b"})
	t.Require().Nil(err)
	t.Equal(int64(1), id)
	t.dao.AssertExpectations(t.T())
	t.jsClient.AssertExpectations(t.T())
}

func (t *taskManagerTestSuite) TestStop() {
	// job not found
	t.dao.On("Get", mock.Anything, mock.Anything).Return(&dao.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.RunningStatus.String(),
	}, nil)
	t.jsClient.On("PostAction", mock.Anything, mock.Anything).Return(cjob.ErrJobNotFound)
	t.dao.On("Update", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := t.mgr.Stop(nil, 1)
	t.Require().Nil(err)
	t.dao.AssertExpectations(t.T())
	t.jsClient.AssertExpectations(t.T())

	// reset mock
	t.SetupTest()

	// pass
	t.dao.On("Get", mock.Anything, mock.Anything).Return(&dao.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.RunningStatus.String(),
	}, nil)
	t.jsClient.On("PostAction", mock.Anything, mock.Anything).Return(nil)
	err = t.mgr.Stop(nil, 1)
	t.Require().Nil(err)
	t.dao.AssertExpectations(t.T())
	t.jsClient.AssertExpectations(t.T())
}

func (t *taskManagerTestSuite) TestGet() {
	t.dao.On("Get", mock.Anything, mock.Anything).Return(&dao.Task{
		ID: 1,
	}, nil)
	task, err := t.mgr.Get(nil, 1)
	t.Require().Nil(err)
	t.Equal(int64(1), task.ID)
	t.dao.AssertExpectations(t.T())
}

func (t *taskManagerTestSuite) TestList() {
	t.dao.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID: 1,
		},
	}, nil)
	tasks, err := t.mgr.List(nil, nil)
	t.Require().Nil(err)
	t.Require().Len(tasks, 1)
	t.Equal(int64(1), tasks[0].ID)
	t.dao.AssertExpectations(t.T())
}

func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, &taskManagerTestSuite{})
}
