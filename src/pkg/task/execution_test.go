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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type executionManagerTestSuite struct {
	suite.Suite
	execMgr *executionManager
	taskMgr *mockTaskManager
	execDAO *mockExecutionDAO
	taskDAO *mockTaskDAO
}

func (e *executionManagerTestSuite) SetupTest() {
	e.taskMgr = &mockTaskManager{}
	e.execDAO = &mockExecutionDAO{}
	e.taskDAO = &mockTaskDAO{}
	e.execMgr = &executionManager{
		executionDAO: e.execDAO,
		taskMgr:      e.taskMgr,
		taskDAO:      e.taskDAO,
	}
}

func (e *executionManagerTestSuite) TestCreate() {
	e.execDAO.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := e.execMgr.Create(nil, "vendor", 0, ExecutionTriggerManual,
		map[string]interface{}{"k": "v"})
	e.Require().Nil(err)
	e.Equal(int64(1), id)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestMarkDone() {
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.MarkDone(nil, 1, "success")
	e.Require().Nil(err)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestMarkError() {
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.MarkError(nil, 1, "error")
	e.Require().Nil(err)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestStop() {
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
		},
	}, nil)
	e.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.Stop(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.taskMgr.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestDelete() {
	// try to delete the execution which contains running tasks
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.RunningStatus.String(),
		},
	}, nil)
	err := e.execMgr.Delete(nil, 1)
	e.Require().NotNil(err)
	e.True(errors.IsErr(err, errors.PreconditionCode))
	e.taskDAO.AssertExpectations(e.T())

	// reset the mock
	e.SetupTest()

	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.SuccessStatus.String(),
		},
	}, nil)
	e.taskDAO.On("Delete", mock.Anything, mock.Anything).Return(nil)
	e.execDAO.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err = e.execMgr.Delete(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestGet() {
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.SuccessStatus.String(),
	}, nil)
	exec, err := e.execMgr.Get(nil, 1)
	e.Require().Nil(err)
	e.Equal(int64(1), exec.ID)
	e.Equal(job.SuccessStatus.String(), exec.Status)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestList() {
	e.execDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Execution{
		{
			ID:     1,
			Status: job.SuccessStatus.String(),
		},
	}, nil)
	execs, err := e.execMgr.List(nil, nil)
	e.Require().Nil(err)
	e.Require().Len(execs, 1)
	e.Equal(int64(1), execs[0].ID)
	e.Equal(job.SuccessStatus.String(), execs[0].Status)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestPopulateExecutionMetrics() {
	e.taskDAO.On("ListStatusCount", mock.Anything, mock.Anything).Return([]*dao.StatusCount{
		{
			Status: job.SuccessStatus.String(),
			Count:  1,
		},
		{
			Status: job.ErrorStatus.String(),
			Count:  1,
		},
		{
			Status: job.StoppedStatus.String(),
			Count:  1,
		},
		{
			Status: job.RunningStatus.String(),
			Count:  1,
		},
		{
			Status: job.PendingStatus.String(),
			Count:  1,
		},
		{
			Status: job.ScheduledStatus.String(),
			Count:  1,
		},
	}, nil)
	exec := &Execution{}
	e.execMgr.populateExecutionMetrics(nil, exec)
	e.Require().NotNil(exec.Metrics)
	e.Equal(int64(6), exec.Metrics.TaskCount)
	e.Equal(int64(1), exec.Metrics.SuccessTaskCount)
	e.Equal(int64(1), exec.Metrics.ErrorTaskCount)
	e.Equal(int64(1), exec.Metrics.StoppedTaskCount)
	e.Equal(int64(1), exec.Metrics.PendingTaskCount)
	e.Equal(int64(1), exec.Metrics.RunningTaskCount)
	e.Equal(int64(1), exec.Metrics.ScheduledTaskCount)
	e.taskDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestPopulateExecutionStatus() {
	// running
	exec := &Execution{}
	e.execMgr.populateExecutionStatus(exec)
	e.Equal(job.RunningStatus.String(), exec.Status)

	// running
	exec = &Execution{
		Metrics: &Metrics{
			SuccessTaskCount:   1,
			ErrorTaskCount:     1,
			PendingTaskCount:   1,
			RunningTaskCount:   1,
			ScheduledTaskCount: 1,
			StoppedTaskCount:   1,
		},
	}
	e.execMgr.populateExecutionStatus(exec)
	e.Equal(job.RunningStatus.String(), exec.Status)

	// error
	exec = &Execution{
		Metrics: &Metrics{
			SuccessTaskCount:   1,
			ErrorTaskCount:     1,
			PendingTaskCount:   0,
			RunningTaskCount:   0,
			ScheduledTaskCount: 0,
			StoppedTaskCount:   1,
		},
	}
	e.execMgr.populateExecutionStatus(exec)
	e.Equal(job.ErrorStatus.String(), exec.Status)

	// stopped
	exec = &Execution{
		Metrics: &Metrics{
			SuccessTaskCount:   1,
			ErrorTaskCount:     0,
			PendingTaskCount:   0,
			RunningTaskCount:   0,
			ScheduledTaskCount: 0,
			StoppedTaskCount:   1,
		},
	}
	e.execMgr.populateExecutionStatus(exec)
	e.Equal(job.StoppedStatus.String(), exec.Status)

	// success
	exec = &Execution{
		Metrics: &Metrics{
			SuccessTaskCount:   1,
			ErrorTaskCount:     0,
			PendingTaskCount:   0,
			RunningTaskCount:   0,
			ScheduledTaskCount: 0,
			StoppedTaskCount:   0,
		},
	}
	e.execMgr.populateExecutionStatus(exec)
	e.Equal(job.SuccessStatus.String(), exec.Status)
}

func (e *executionManagerTestSuite) TestPopulateExecutionEndTime() {
	// isn't final status
	exec := &Execution{
		Status: job.RunningStatus.String(),
	}
	e.execMgr.populateExecutionEndTime(nil, exec)
	e.Equal(time.Time{}, exec.EndTime)

	// final status
	now := time.Now()
	exec = &Execution{
		Status: job.SuccessStatus.String(),
	}
	e.taskDAO.On("GetMaxEndTime", mock.Anything, mock.Anything).Return(now, nil)
	e.execMgr.populateExecutionEndTime(nil, exec)
	e.Equal(now, exec.EndTime)
	e.taskDAO.AssertExpectations(e.T())
}

func TestExecutionManagerSuite(t *testing.T) {
	suite.Run(t, &executionManagerTestSuite{})
}
