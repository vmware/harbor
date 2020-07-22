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

	model "github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type controllerTestSuite struct {
	suite.Suite
	ctl *controller
	mgr *task.FakeManager
}

// TestControllerTestSuite tests controller.
func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}

// SetupTest setups the testing env.
func (c *controllerTestSuite) SetupTest() {
	c.mgr = &task.FakeManager{}
	c.ctl = &controller{mgr: c.mgr}
}

// TestStop tests stop.
func (c *controllerTestSuite) TestStop() {
	c.mgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err := c.ctl.Stop(nil, 1)
	c.NoError(err)
}

// TestGet tests get.
func (c *controllerTestSuite) TestGet() {
	c.mgr.On("Get", mock.Anything, int64(1)).Return(&model.Task{ID: 1}, nil)
	t, err := c.ctl.Get(nil, 1)
	c.NoError(err)
	c.Equal(int64(1), t.ID)
}

// TestList tests list.
func (c *controllerTestSuite) TestList() {
	c.mgr.On("List", mock.Anything, mock.Anything).Return([]*model.Task{
		{ID: 1}, {ID: 2},
	}, nil)
	ts, err := c.ctl.List(nil, nil)
	c.NoError(err)
	c.Len(ts, 2)
}

// TestGetLog tests get log.
func (c *controllerTestSuite) TestGetLog() {
	c.mgr.On("GetLog", mock.Anything, mock.Anything).Return([]byte("logs"), nil)
	l, err := c.ctl.GetLog(nil, 1)
	c.NoError(err)
	c.Equal([]byte("logs"), l)
}
