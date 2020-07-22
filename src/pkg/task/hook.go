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
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// NewHookHandler creates a hook handler instance
func NewHookHandler() *HookHandler {
	return &HookHandler{
		taskDAO:      dao.NewTaskDAO(),
		executionDAO: dao.NewExecutionDAO(),
	}
}

// HookHandler handles the job status changing webhook
type HookHandler struct {
	taskDAO      dao.TaskDAO
	executionDAO dao.ExecutionDAO
}

// Handle the job status changing webhook
func (h *HookHandler) Handle(ctx context.Context, taskID int64, sc *job.StatusChange) error {
	// process check in data
	if len(sc.CheckIn) > 0 {
		task, err := h.taskDAO.Get(ctx, taskID)
		if err != nil {
			return err
		}
		execution, err := h.executionDAO.Get(ctx, task.ExecutionID)
		if err != nil {
			return err
		}

		processor, exist := registry[execution.VendorType]
		if !exist {
			return fmt.Errorf("the check in processor for task %d not found", taskID)
		}
		t := &Task{}
		t.From(task)
		return processor(ctx, t, sc)
	}

	// update status
	return h.taskDAO.UpdateStatus(ctx, taskID, sc.Status, sc.Metadata.Revision)
}
