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

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	// Ctl is a global task controller instance
	Ctl = NewController()
)

// Controller manages the task
type Controller interface {
	// Stop the specified task.
	Stop(ctx context.Context, id int64) (err error)
	// Get the specified task.
	Get(ctx context.Context, id int64) (task *task.Task, err error)
	// List the tasks according to the query.
	List(ctx context.Context, query *q.Query) (tasks []*task.Task, err error)
	// Get the log of the specified task.
	GetLog(ctx context.Context, id int64) (log []byte, err error)
}

// NewController creates an instance of the default task controller.
func NewController() Controller {
	return &controller{
		mgr: task.Mgr,
	}
}

// controller defines the default task controller.
type controller struct {
	mgr task.Manager
}

// Stop the specified task.
func (c *controller) Stop(ctx context.Context, id int64) (err error) {
	return c.mgr.Stop(ctx, id)
}

// Get the specified task.
func (c *controller) Get(ctx context.Context, id int64) (task *task.Task, err error) {
	return c.mgr.Get(ctx, id)
}

// List the tasks according to the query.
func (c *controller) List(ctx context.Context, query *q.Query) (tasks []*task.Task, err error) {
	return c.mgr.List(ctx, query)
}

// Get the log of the specified task.
func (c *controller) GetLog(ctx context.Context, id int64) (log []byte, err error) {
	return c.mgr.GetLog(ctx, id)
}
