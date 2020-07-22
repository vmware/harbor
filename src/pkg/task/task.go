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
	"fmt"
	"time"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// Mgr is a global task manager instance
	Mgr = NewManager()
)

// Manager manages tasks.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type Manager interface {
	// Create submits the job to jobservice and creates a corresponding task record.
	// An execution must be created first and the task will be linked to it.
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, executionID int64, job *Job, extraAttrs ...map[string]interface{}) (id int64, err error)
	// Stop the specified task
	Stop(ctx context.Context, id int64) (err error)
	// Get the specified task
	Get(ctx context.Context, id int64) (task *Task, err error)
	// List the tasks according to the query
	List(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// Get the log of the specified task
	GetLog(ctx context.Context, id int64) (log []byte, err error)
}

// NewManager creates an instance of the default task manager
func NewManager() Manager {
	return &manager{
		dao:      dao.NewTaskDAO(),
		jsClient: cjob.GlobalClient,
		coreURL:  config.GetCoreURL(),
	}
}

type manager struct {
	dao      dao.TaskDAO
	jsClient cjob.Client
	coreURL  string
}

func (m *manager) Create(ctx context.Context, executionID int64, jb *Job, extraAttrs ...map[string]interface{}) (int64, error) {
	// create task record in database
	id, err := m.createTaskRecord(ctx, executionID, extraAttrs...)
	if err != nil {
		return 0, err
	}
	log.Debugf("the database record for task %d created", id)

	// submit job to jobservice
	jobID, err := m.submitJob(ctx, id, jb)
	if err != nil {
		// failed to submit job to jobservice, update the status of task to error
		err = fmt.Errorf("failed to submit job to jobservice: %v", err)
		log.Error(err)
		now := time.Now()
		err = m.dao.Update(ctx, &dao.Task{
			ID:            id,
			Status:        job.ErrorStatus.String(),
			StatusCode:    job.ErrorStatus.Code(),
			StatusMessage: err.Error(),
			UpdateTime:    now,
			EndTime:       now,
		}, "Status", "StatusCode", "StatusMessage", "UpdateTime", "EndTime")
		if err != nil {
			log.Errorf("failed to update task %d: %v", id, err)
		}
		return id, nil
	}

	log.Debugf("the task %d is submitted to jobservice, the job ID is %s", id, jobID)

	// populate the job ID for the task
	if err = m.dao.Update(ctx, &dao.Task{
		ID:    id,
		JobID: jobID,
	}, "JobID"); err != nil {
		log.Errorf("failed to populate the job ID for the task %d: %v", id, err)
	}

	return id, nil
}

func (m *manager) createTaskRecord(ctx context.Context, executionID int64, extraAttrs ...map[string]interface{}) (int64, error) {
	extras := map[string]interface{}{}
	if len(extraAttrs) > 0 && extraAttrs[0] != nil {
		extras = extraAttrs[0]
	}
	data, err := json.Marshal(extras)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	return m.dao.Create(ctx, &dao.Task{
		ExecutionID:  executionID,
		Status:       job.PendingStatus.String(),
		StatusCode:   job.PendingStatus.Code(),
		ExtraAttrs:   string(data),
		CreationTime: now,
		UpdateTime:   now,
	})
}

func (m *manager) submitJob(ctx context.Context, id int64, jb *Job) (string, error) {
	jobData := &models.JobData{
		Name:       jb.Name,
		StatusHook: fmt.Sprintf("%s/service/notifications/tasks/%d", m.coreURL, id),
	}
	if jb.Parameters != nil {
		jobData.Parameters = models.Parameters(jb.Parameters)
	}
	if jb.Metadata != nil {
		jobData.Metadata = &models.JobMetadata{
			JobKind:       jb.Metadata.JobKind,
			ScheduleDelay: jb.Metadata.ScheduleDelay,
			Cron:          jb.Metadata.Cron,
			IsUnique:      jb.Metadata.IsUnique,
		}
	}

	return m.jsClient.SubmitJob(jobData)
}

func (m *manager) Stop(ctx context.Context, id int64) error {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return err
	}

	// when a task is in final status, if it's a periodic or retrying job it will
	// run again in the near future, so we must operate the stop action to these final
	// status jobs as well
	if err = m.jsClient.PostAction(task.JobID, string(job.StopCommand)); err != nil {
		// job not found, update it's status to stop directly
		if err == cjob.ErrJobNotFound {
			now := time.Now()
			err = m.dao.Update(ctx, &dao.Task{
				ID:         task.ID,
				Status:     job.StoppedStatus.String(),
				StatusCode: job.StoppedStatus.Code(),
				UpdateTime: now,
				EndTime:    now,
			}, "Status", "StatusCode", "UpdateTime", "EndTime")
			if err != nil {
				return err
			}
			log.Debugf("got job not found error for task %d, update it's status to stop directly", task.ID)
			return nil
		}
		return err
	}
	log.Debugf("the stop request for task %d is sent", id)
	return nil
}

func (m *manager) Get(ctx context.Context, id int64) (*Task, error) {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	t := &Task{}
	t.From(task)
	return t, nil
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*Task, error) {
	tasks, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var ts []*Task
	for _, task := range tasks {
		t := &Task{}
		t.From(task)
		ts = append(ts, t)
	}
	return ts, nil
}

func (m *manager) GetLog(ctx context.Context, id int64) ([]byte, error) {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return m.jsClient.GetJobLog(task.JobID)
}
