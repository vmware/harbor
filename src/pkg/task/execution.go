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
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// ExecMgr is a global execution manager instance
	ExecMgr = NewExecutionManager()
)

// ExecutionManager manages executions.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type ExecutionManager interface {
	// Create an execution. The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
		extraAttrs ...map[string]interface{}) (id int64, err error)
	// MarkDone marks the status of the specified execution as success.
	// It must be called to update the execution status if the created execution contains no tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkDone(ctx context.Context, id int64, message string) (err error)
	// MarkError marks the status of the specified execution as error.
	// It must be called to update the execution status when failed to create tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkError(ctx context.Context, id int64, message string) (err error)
	// Stop all linked tasks of the specified execution
	Stop(ctx context.Context, id int64) (err error)
	// Delete the specified execution and its tasks
	Delete(ctx context.Context, id int64) (err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// List executions according to the query
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
}

// NewExecutionManager return an instance of the default execution manager
func NewExecutionManager() ExecutionManager {
	return &executionManager{
		executionDAO: dao.NewExecutionDAO(),
		taskMgr:      Mgr,
		taskDAO:      dao.NewTaskDAO(),
	}
}

type executionManager struct {
	executionDAO dao.ExecutionDAO
	taskMgr      Manager
	taskDAO      dao.TaskDAO
}

func (e *executionManager) Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
	extraAttrs ...map[string]interface{}) (int64, error) {
	extras := map[string]interface{}{}
	if len(extraAttrs) > 0 && extraAttrs[0] != nil {
		extras = extraAttrs[0]
	}
	data, err := json.Marshal(extras)
	if err != nil {
		return 0, err
	}

	execution := &dao.Execution{
		VendorType: vendorType,
		VendorID:   vendorID,
		Trigger:    trigger,
		ExtraAttrs: string(data),
		StartTime:  time.Now(),
	}
	return e.executionDAO.Create(ctx, execution)
}

func (e *executionManager) MarkDone(ctx context.Context, id int64, message string) error {
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.SuccessStatus.String(),
		StatusMessage: message,
		EndTime:       time.Now(),
	}, "Status", "StatusMessage", "EndTime")
}

func (e *executionManager) MarkError(ctx context.Context, id int64, message string) error {
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.ErrorStatus.String(),
		StatusMessage: message,
		EndTime:       time.Now(),
	}, "Status", "StatusMessage", "EndTime")
}

func (e *executionManager) Stop(ctx context.Context, id int64) error {
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err = e.taskMgr.Stop(ctx, task.ID); err != nil {
			log.Errorf("failed to stop task %d: %v", task.ID, err)
			continue
		}
	}
	return nil
}

func (e *executionManager) Delete(ctx context.Context, id int64) error {
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if !job.Status(task.Status).Final() {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessage("the execution %d has tasks that aren't in final status, stop the tasks first", id)
		}
		if err = e.taskDAO.Delete(ctx, task.ID); err != nil {
			return err
		}
	}

	return e.executionDAO.Delete(ctx, id)
}

func (e *executionManager) Get(ctx context.Context, id int64) (*Execution, error) {
	execution, err := e.executionDAO.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return e.populateExecution(ctx, execution), nil
}

func (e *executionManager) List(ctx context.Context, query *q.Query) ([]*Execution, error) {
	executions, err := e.executionDAO.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var execs []*Execution
	for _, execution := range executions {
		execs = append(execs, e.populateExecution(ctx, execution))
	}
	return execs, nil
}

func (e *executionManager) populateExecution(ctx context.Context, execution *dao.Execution) *Execution {
	exec := &Execution{
		ID:            execution.ID,
		VendorType:    execution.VendorType,
		VendorID:      execution.VendorID,
		Status:        execution.Status,
		StatusMessage: execution.StatusMessage,
		Metrics:       nil,
		Trigger:       execution.Trigger,
		StartTime:     execution.StartTime,
		EndTime:       execution.EndTime,
	}

	if len(execution.ExtraAttrs) > 0 {
		extras := map[string]interface{}{}
		if err := json.Unmarshal([]byte(execution.ExtraAttrs), &extras); err != nil {
			log.Errorf("failed to unmarshal the extra attributes of execution %d: %v", execution.ID, err)
		} else {
			exec.ExtraAttrs = extras
		}
	}

	// if the status isn't null which means the status is set manually, return directly
	if len(exec.Status) > 0 {
		return exec
	}

	// populate task metrics
	e.populateExecutionMetrics(ctx, exec)
	// populate status
	e.populateExecutionStatus(exec)
	// populate the end time
	e.populateExecutionEndTime(ctx, exec)

	return exec
}

func (e *executionManager) populateExecutionMetrics(ctx context.Context, execution *Execution) {
	scs, err := e.taskDAO.ListStatusCount(ctx, execution.ID)
	if err != nil {
		log.Errorf("failed to list status count of execution %d: %v", execution.ID, err)
		return
	}
	if len(scs) == 0 {
		return
	}

	metrics := &Metrics{}
	for _, sc := range scs {
		switch sc.Status {
		case job.SuccessStatus.String():
			metrics.SuccessTaskCount = sc.Count
		case job.ErrorStatus.String():
			metrics.ErrorTaskCount = sc.Count
		case job.PendingStatus.String():
			metrics.PendingTaskCount = sc.Count
		case job.RunningStatus.String():
			metrics.RunningTaskCount = sc.Count
		case job.ScheduledStatus.String():
			metrics.ScheduledTaskCount = sc.Count
		case job.StoppedStatus.String():
			metrics.StoppedTaskCount = sc.Count
		default:
			log.Errorf("unknown task status: %s", sc.Status)
		}
	}
	metrics.TaskCount = metrics.SuccessTaskCount + metrics.ErrorTaskCount +
		metrics.PendingTaskCount + metrics.RunningTaskCount +
		metrics.ScheduledTaskCount + metrics.StoppedTaskCount
	execution.Metrics = metrics
}

func (e *executionManager) populateExecutionStatus(execution *Execution) {
	metrics := execution.Metrics
	if metrics == nil {
		execution.Status = job.RunningStatus.String()
		return
	}
	if metrics.PendingTaskCount > 0 || metrics.RunningTaskCount > 0 || metrics.ScheduledTaskCount > 0 {
		execution.Status = job.RunningStatus.String()
		return
	}
	if metrics.ErrorTaskCount > 0 {
		execution.Status = job.ErrorStatus.String()
		return
	}
	if metrics.StoppedTaskCount > 0 {
		execution.Status = job.StoppedStatus.String()
		return
	}
	if metrics.SuccessTaskCount > 0 {
		execution.Status = job.SuccessStatus.String()
		return
	}
}

func (e *executionManager) populateExecutionEndTime(ctx context.Context, execution *Execution) {
	if !job.Status(execution.Status).Final() {
		return
	}
	endTime, err := e.taskDAO.GetMaxEndTime(ctx, execution.ID)
	if err != nil {
		log.Errorf("failed to get the max end time of the execution %d: %v", execution.ID, err)
		return
	}
	execution.EndTime = endTime
}
