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
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type hookHandlerTestSuite struct {
	suite.Suite
	handler *HookHandler
	execDAO *mockExecutionDAO
	taskDAO *mockTaskDAO
}

func (h *hookHandlerTestSuite) SetupTest() {
	h.execDAO = &mockExecutionDAO{}
	h.taskDAO = &mockTaskDAO{}
	h.handler = &HookHandler{
		taskDAO:      h.taskDAO,
		executionDAO: h.execDAO,
	}
}

func (h *hookHandlerTestSuite) TestHandle() {
	// handle check in data
	registry["test"] = func(ctx context.Context, task *Task, change *job.StatusChange) (err error) { return nil }
	h.taskDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Task{
		ID:          1,
		ExecutionID: 1,
	}, nil)
	h.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:         1,
		VendorType: "test",
	}, nil)
	sc := &job.StatusChange{
		CheckIn: "data",
	}
	err := h.handler.Handle(nil, 1, sc)
	h.Require().Nil(err)
	h.taskDAO.AssertExpectations(h.T())
	h.execDAO.AssertExpectations(h.T())

	// reset mock
	h.SetupTest()

	// handle status changing
	h.taskDAO.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sc = &job.StatusChange{
		Status: job.SuccessStatus.String(),
		Metadata: &job.StatsInfo{
			Revision: time.Now().Unix(),
		},
	}
	err = h.handler.Handle(nil, 1, sc)
	h.Require().Nil(err)
	h.taskDAO.AssertExpectations(h.T())
}

func TestHookHandlerTestSuite(t *testing.T) {
	suite.Run(t, &hookHandlerTestSuite{})
}
