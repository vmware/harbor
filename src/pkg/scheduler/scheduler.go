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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	cronlib "github.com/robfig/cron"
)

var (
	// Sched is an instance of the default scheduler that can be used globally
	Sched = New()
)

// Schedule describes the detail information about the created schedule
type Schedule struct {
	ID           int64     `json:"id"`
	CRON         string    `json:"cron"`
	Status       string    `json:"status"` // status of the underlying task(jobservice job)
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
	// we can extend this model to include more information(e.g. how many times the schedule already
	// runs; when will the schedule runs next time)
}

// Scheduler provides the capability to run a periodic task, a callback function
// needs to be registered before using the scheduler
type Scheduler interface {
	// Schedule creates a task which calls the specified callback function periodically
	// The callback function needs to be registered first
	// The "params" is passed to the callback function as encoded json string, so the callback
	// function must decode it before using
	Schedule(ctx context.Context, cron string, callbackFuncName string, params interface{}) (int64, error)
	// UnSchedule the created schedule instance
	UnSchedule(ctx context.Context, id int64) error
	// GetSchedule gets the schedule specified by ID
	GetSchedule(ctx context.Context, id int64) (*Schedule, error)
}

// New returns an instance of the default scheduler
func New() Scheduler {
	return &scheduler{
		dao:     &dao{},
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
	}
}

type scheduler struct {
	dao     DAO
	execMgr task.ExecutionManager
	taskMgr task.Manager
}

func (s *scheduler) Schedule(ctx context.Context, cron string, callbackFuncName string, params interface{}) (int64, error) {
	if _, err := cronlib.Parse(cron); err != nil {
		return 0, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("invalid cron %s: %v", cron, err)
	}
	if !callbackFuncExist(callbackFuncName) {
		return 0, fmt.Errorf("callback function %s not found", callbackFuncName)
	}

	execID, err := s.execMgr.Create(ctx, JobNameScheduler, 0, task.ExecutionTriggerManual)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	sched := &schedule{
		CRON:             cron,
		ExecutionID:      execID,
		CallbackFuncName: callbackFuncName,
		CreationTime:     now,
		UpdateTime:       now,
	}
	if params != nil {
		paramsData, err := json.Marshal(params)
		if err != nil {
			return 0, err
		}
		sched.CallbackFuncParam = string(paramsData)
	}

	// create schedule record
	// when status/checkin hook comes, the database record must exist,
	// so the database record must be created first before submitting job
	id, err := s.dao.Create(ctx, sched)
	if err != nil {
		return 0, err
	}

	taskID, err := s.taskMgr.Create(ctx, execID, &task.Job{
		Name: JobNameScheduler,
		Metadata: &job.Metadata{
			JobKind: job.KindPeriodic,
			Cron:    cron,
		},
	})
	if err != nil {
		return 0, err
	}
	// when task manager creating a task, it creates the task database record first and
	// then submits the job to jobservice. If the submitting failed, it doesn't return
	// any error. So we check the task status to make sure the job is submitted to jobservice
	// successfully here
	task, err := s.taskMgr.Get(ctx, taskID)
	if err != nil {
		return 0, err
	}
	if task.Status == job.ErrorStatus.String() {
		return 0, fmt.Errorf("failed to create the schedule: the task status is %s", job.ErrorStatus.String())
	}

	return id, nil
}

func (s *scheduler) UnSchedule(ctx context.Context, id int64) error {
	schedule, err := s.dao.Get(ctx, id)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			log.Warningf("trying to unschedule a non existing schedule %d, skip directly", id)
			return nil
		}
		return err
	}
	if err = s.execMgr.Stop(ctx, schedule.ExecutionID); err != nil {
		return err
	}

	// after the stop called, the execution cannot be stopped immediately,
	// use the for loop to make sure the execution be in final status before deleting it
	for t := 100 * time.Microsecond; t < 5*time.Second; t = t * 2 {
		exec, err := s.execMgr.Get(ctx, schedule.ExecutionID)
		if err != nil {
			return err
		}
		if job.Status(exec.Status).Final() {
			// delete schedule record
			if err = s.dao.Delete(ctx, id); err != nil {
				return err
			}
			// delete execution
			return s.execMgr.Delete(ctx, schedule.ExecutionID)
		}
		time.Sleep(t)
	}

	return fmt.Errorf("failed to unschedule the schedule %d: the execution isn't in final status", id)
}

func (s *scheduler) GetSchedule(ctx context.Context, id int64) (*Schedule, error) {
	schedule, err := s.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	schd := &Schedule{
		ID:           schedule.ID,
		CRON:         schedule.CRON,
		CreationTime: schedule.CreationTime,
		UpdateTime:   schedule.UpdateTime,
	}
	exec, err := s.execMgr.Get(ctx, schedule.ExecutionID)
	if err != nil {
		return nil, err
	}
	schd.Status = exec.Status
	return schd, nil
}

// HandleLegacyHook handles the legacy web hook for scheduler
// We rewrite the implementation of scheduler with task manager mechanism in v2.1,
// this method is used to handle the job status hook for the legacy implementation
// We can remove the method and the hook endpoint after several releases
func HandleLegacyHook(ctx context.Context, scheduleID int64, sc *job.StatusChange) error {
	scheduler := Sched.(*scheduler)
	schedule, err := scheduler.dao.Get(ctx, scheduleID)
	if err != nil {
		return err
	}
	tasks, err := scheduler.taskMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID": schedule.ExecutionID,
		},
	})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("no task references the execution %d", schedule.ExecutionID)
	}
	return task.NewHookHandler().Handle(ctx, tasks[0].ID, sc)
}
