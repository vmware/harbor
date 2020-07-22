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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
	"testing"
)

type schedulerTestSuite struct {
	suite.Suite
	scheduler *scheduler
	dao       *mockDAO
	execMgr   *tasktesting.FakeExecutionManager
	taskMgr   *tasktesting.FakeManager
}

func (s *schedulerTestSuite) SetupTest() {
	registry = map[string]CallbackFunc{}
	err := RegisterCallbackFunc("callback", func(interface{}) error { return nil })
	s.Require().Nil(err)

	s.dao = &mockDAO{}
	s.execMgr = &tasktesting.FakeExecutionManager{}
	s.taskMgr = &tasktesting.FakeManager{}

	s.scheduler = &scheduler{
		dao:     s.dao,
		execMgr: s.execMgr,
		taskMgr: s.taskMgr,
	}
}

func (s *schedulerTestSuite) TestSchedule() {
	// invalid cron
	id, err := s.scheduler.Schedule(nil, "", "callback", nil)
	s.NotNil(err)

	// callback function not exist
	id, err = s.scheduler.Schedule(nil, "0 * * * * *", "not-exist", nil)
	s.NotNil(err)

	// failed to submit to jobservice
	s.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.ErrorStatus.String(),
	}, nil)
	_, err = s.scheduler.Schedule(nil, "0 * * * * *", "callback", "param")
	s.Require().NotNil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
	s.taskMgr.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// pass
	s.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.SuccessStatus.String(),
	}, nil)
	id, err = s.scheduler.Schedule(nil, "0 * * * * *", "callback", "param")
	s.Require().Nil(err)
	s.Equal(int64(1), id)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
	s.taskMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestUnSchedule() {
	// not existing schedule
	s.dao.On("Get", mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	err := s.scheduler.UnSchedule(nil, 10000)
	s.Nil(err)
	s.dao.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// the underlying task isn't stopped
	s.dao.On("Get", mock.Anything, mock.Anything).Return(&schedule{
		ID:               1,
		CRON:             "0 * * * * *",
		ExecutionID:      1,
		CallbackFuncName: "callback",
	}, nil)
	s.execMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	s.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
	}, nil)
	err = s.scheduler.UnSchedule(nil, 1)
	s.NotNil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// pass
	s.dao.On("Get", mock.Anything, mock.Anything).Return(&schedule{
		ID:               1,
		CRON:             "0 * * * * *",
		ExecutionID:      1,
		CallbackFuncName: "callback",
	}, nil)
	s.execMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	s.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		ID:     1,
		Status: job.StoppedStatus.String(),
	}, nil)
	s.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	s.execMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err = s.scheduler.UnSchedule(nil, 1)
	s.Nil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestGetSchedule() {
	s.dao.On("Get", mock.Anything, mock.Anything).Return(&schedule{
		ID:          1,
		CRON:        "0 * * * * *",
		ExecutionID: 1,
	}, nil)
	s.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		ID:     1,
		Status: job.SuccessStatus.String(),
	}, nil)
	schedule, err := s.scheduler.GetSchedule(nil, 1)
	s.Require().Nil(err)
	s.Equal("0 * * * * *", schedule.CRON)
	s.Equal(job.SuccessStatus.String(), schedule.Status)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func TestScheduler(t *testing.T) {
	s := &schedulerTestSuite{}
	suite.Run(t, s)
}
