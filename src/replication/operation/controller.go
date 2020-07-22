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

package operation

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/job"
	hjob "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/execution"
	"github.com/goharbor/harbor/src/replication/operation/flow"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
)

// Controller handles the replication-related operations: start,
// stop, query, etc.
type Controller interface {
	// trigger is used to specify what this replication is triggered by
	StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error)
	StopReplication(int64) error
	ListExecutions(...*models.ExecutionQuery) (int64, []*models.Execution, error)
	GetExecution(int64) (*models.Execution, error)
	ListTasks(...*models.TaskQuery) (int64, []*models.Task, error)
	GetTask(int64) (*models.Task, error)
	UpdateTaskStatus(id int64, status string, statusRevision int64, statusCondition ...string) error
	GetTaskLog(int64) ([]byte, error)
}

const (
	maxReplicators = 1024
)

// NewController returns a controller implementation
func NewController(js job.Client) Controller {
	ctl := &controller{
		replicators:  make(chan struct{}, maxReplicators),
		executionMgr: execution.NewDefaultManager(),
		scheduler:    scheduler.NewScheduler(js),
		flowCtl:      flow.NewController(),
	}
	for i := 0; i < maxReplicators; i++ {
		ctl.replicators <- struct{}{}
	}
	return ctl
}

type controller struct {
	replicators  chan struct{}
	flowCtl      flow.Controller
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

func (c *controller) StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error) {
	if !policy.Enabled {
		return 0, fmt.Errorf("the policy %d is disabled", policy.ID)
	}
	if len(trigger) == 0 {
		trigger = model.TriggerTypeManual
	}
	id, err := createExecution(c.executionMgr, policy.ID, trigger)
	if err != nil {
		return 0, err
	}
	// control the count of concurrent replication requests
	log.Debugf("waiting for the available replicator ...")
	<-c.replicators
	log.Debugf("got an available replicator, starting the replication ...")
	go func() {
		defer func() {
			c.replicators <- struct{}{}
		}()
		flow := c.createFlow(id, policy, resource)
		if n, err := c.flowCtl.Start(flow); err != nil {
			// only update the execution when got error.
			// if got no error, it will be updated automatically
			// when listing the execution records
			if e := c.executionMgr.Update(&models.Execution{
				ID:         id,
				Status:     models.ExecutionStatusFailed,
				StatusText: err.Error(),
				Total:      n,
				Failed:     n,
			}, "Status", "StatusText", "Total", "Failed"); e != nil {
				log.Errorf("failed to update the execution %d: %v", id, e)
			}
			log.Errorf("the execution %d failed: %v", id, err)
		}
	}()
	return id, nil
}

// create different replication flows according to the input parameters
func (c *controller) createFlow(executionID int64, policy *model.Policy, resource *model.Resource) flow.Flow {
	// replicate the deletion operation, so create a deletion flow
	if resource != nil && resource.Deleted {
		return flow.NewDeletionFlow(c.executionMgr, c.scheduler, executionID, policy, resource)
	}
	resources := []*model.Resource{}
	if resource != nil {
		resources = append(resources, resource)
	}
	return flow.NewCopyFlow(c.executionMgr, c.scheduler, executionID, policy, resources...)
}

func (c *controller) StopReplication(executionID int64) error {
	_, tasks, err := c.ListTasks(&models.TaskQuery{
		ExecutionID: executionID,
	})
	if err != nil {
		return err
	}
	// no tasks, just set its status to "stopped"
	if len(tasks) == 0 {
		execution, err := c.executionMgr.Get(executionID)
		if err != nil {
			return err
		}
		if execution == nil {
			return fmt.Errorf("the execution %d not found", executionID)
		}
		if execution.Status != models.ExecutionStatusInProgress {
			log.Debugf("the execution %d isn't in progress, no need to stop", executionID)
			return nil
		}
		if err = c.executionMgr.Update(&models.Execution{
			ID:      executionID,
			Status:  models.ExecutionStatusStopped,
			EndTime: time.Now(),
		}, models.ExecutionPropsName.Status, models.ExecutionPropsName.EndTime); err != nil {
			return err
		}
		log.Debugf("the status of execution %d is set to stopped", executionID)
	}
	// got tasks, stopping the tasks one by one
	for _, task := range tasks {
		if isTaskInFinalStatus(task) {
			log.Debugf("the task %d(job ID: %s) is in final status, its status is %s, skip", task.ID, task.JobID, task.Status)
			continue
		}
		if err = c.scheduler.Stop(task.JobID); err != nil {
			isStatusBehindError, ok := err.(*job.StatusBehindError)
			if ok {
				status := isStatusBehindError.Status()
				switch hjob.Status(status) {
				case hjob.ErrorStatus:
					status = models.TaskStatusFailed
				case hjob.SuccessStatus:
					status = models.TaskStatusSucceed
				}
				e := c.executionMgr.UpdateTask(&models.Task{
					ID:     task.ID,
					Status: status,
				}, "Status")
				if e != nil {
					log.Errorf("failed to update the status the task %d(job ID: %s): %v", task.ID, task.JobID, e)
				} else {
					log.Debugf("got status behind error for task %d, update it's status to %s directly", task.ID, status)
				}
				continue
			}
			if err == job.ErrJobNotFound {
				e := c.executionMgr.UpdateTask(&models.Task{
					ID:     task.ID,
					Status: models.ExecutionStatusStopped,
				}, "Status")
				if e != nil {
					log.Errorf("failed to update the status the task %d(job ID: %s): %v", task.ID, task.JobID, e)
				} else {
					log.Debugf("got job not found error for task %d, update it's status to %s directly", task.ID, models.ExecutionStatusStopped)
				}
				continue
			}
			log.Errorf("failed to stop the task %d(job ID: %s): %v", task.ID, task.JobID, err)
			continue
		}
		log.Debugf("the stop request for task %d(job ID: %s) sent", task.ID, task.JobID)
	}
	return nil
}

func isTaskInFinalStatus(task *models.Task) bool {
	if task == nil {
		return false
	}
	switch task.Status {
	case models.TaskStatusSucceed,
		models.TaskStatusStopped,
		models.TaskStatusFailed:
		return true
	}
	return false
}

func (c *controller) ListExecutions(query ...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return c.executionMgr.List(query...)
}
func (c *controller) GetExecution(executionID int64) (*models.Execution, error) {
	return c.executionMgr.Get(executionID)
}
func (c *controller) ListTasks(query ...*models.TaskQuery) (int64, []*models.Task, error) {
	return c.executionMgr.ListTasks(query...)
}
func (c *controller) GetTask(id int64) (*models.Task, error) {
	return c.executionMgr.GetTask(id)
}
func (c *controller) UpdateTaskStatus(id int64, status string, statusRevision int64, statusCondition ...string) error {
	return c.executionMgr.UpdateTaskStatus(id, status, statusRevision, statusCondition...)
}
func (c *controller) GetTaskLog(taskID int64) ([]byte, error) {
	return c.executionMgr.GetTaskLog(taskID)
}

// create the execution record in database
func createExecution(mgr execution.Manager, policyID int64, trigger model.TriggerType) (int64, error) {
	id, err := mgr.Create(&models.Execution{
		PolicyID:  policyID,
		Trigger:   trigger,
		Status:    models.ExecutionStatusInProgress,
		StartTime: time.Now(),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create the execution record for replication based on policy %d: %v", policyID, err)
	}
	log.Debugf("an execution record for replication based on the policy %d created: %d", policyID, id)
	return id, nil
}
