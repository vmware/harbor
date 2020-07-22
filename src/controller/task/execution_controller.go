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
	// ExecutionCtl is a global execution controller.
	ExecutionCtl = NewExecutionController()
)

// ExecutionController manages the execution.
type ExecutionController interface {
	// Stop all linked tasks of the specified execution.
	Stop(ctx context.Context, id int64) (err error)
	// Delete the specified execution and its tasks.
	Delete(ctx context.Context, id int64) (err error)
	// Get the specified execution.
	Get(ctx context.Context, id int64) (execution *task.Execution, err error)
	// List executions according to the query.
	List(ctx context.Context, query *q.Query) (executions []*task.Execution, err error)
}

// executionController defines the execution controller.
type executionController struct {
	mgr task.ExecutionManager
}

// NewExecutionController creates an instance of the default execution controller.
func NewExecutionController() ExecutionController {
	return &executionController{
		mgr: task.ExecMgr,
	}
}

// Stop all linked tasks of the specified execution.
func (ec *executionController) Stop(ctx context.Context, id int64) (err error) {
	return ec.mgr.Stop(ctx, id)
}

// Delete the specified execution and its tasks.
func (ec *executionController) Delete(ctx context.Context, id int64) (err error) {
	return ec.mgr.Delete(ctx, id)
}

// Get the specified execution.
func (ec *executionController) Get(ctx context.Context, id int64) (execution *task.Execution, err error) {
	return ec.mgr.Get(ctx, id)
}

// List executions according to the query.
func (ec *executionController) List(ctx context.Context, query *q.Query) (executions []*task.Execution, err error) {
	return ec.mgr.List(ctx, query)
}
