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

package retention

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/orm"
	"time"

	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
)

// go:generate mockery -name APIController -case snake

// APIController to handle the requests related with retention
type APIController interface {
	GetRetention(id int64) (*policy.Metadata, error)

	CreateRetention(p *policy.Metadata) (int64, error)

	UpdateRetention(p *policy.Metadata) error

	DeleteRetention(id int64) error

	TriggerRetentionExec(policyID int64, trigger string, dryRun bool) (int64, error)

	OperateRetentionExec(eid int64, action string) error

	GetRetentionExec(eid int64) (*Execution, error)

	ListRetentionExecs(policyID int64, query *q.Query) ([]*Execution, error)

	GetTotalOfRetentionExecs(policyID int64) (int64, error)

	ListRetentionExecTasks(executionID int64, query *q.Query) ([]*Task, error)

	GetTotalOfRetentionExecTasks(executionID int64) (int64, error)

	GetRetentionExecTaskLog(taskID int64) ([]byte, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager        Manager
	launcher       Launcher
	projectManager project.Manager
	repositoryMgr  repository.Manager
	scheduler      scheduler.Scheduler
}

const (
	// SchedulerCallback ...
	SchedulerCallback = "RetentionCallback"
)

// TriggerParam ...
type TriggerParam struct {
	PolicyID int64
	Trigger  string
}

// GetRetention Get Retention
func (r *DefaultAPIController) GetRetention(id int64) (*policy.Metadata, error) {
	return r.manager.GetPolicy(id)
}

// CreateRetention Create Retention
func (r *DefaultAPIController) CreateRetention(p *policy.Metadata) (int64, error) {
	if p.Trigger.Kind == policy.TriggerKindSchedule {
		cron, ok := p.Trigger.Settings[policy.TriggerSettingsCron]
		if ok && len(cron.(string)) > 0 {
			jobid, err := r.scheduler.Schedule(orm.Context(), cron.(string), SchedulerCallback, TriggerParam{
				PolicyID: p.ID,
				Trigger:  ExecutionTriggerSchedule,
			})
			if err != nil {
				return 0, err
			}
			if p.Trigger.References == nil {
				p.Trigger.References = map[string]interface{}{}
			}
			p.Trigger.References[policy.TriggerReferencesJobid] = jobid
		}
	}
	id, err := r.manager.CreatePolicy(p)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateRetention Update Retention
func (r *DefaultAPIController) UpdateRetention(p *policy.Metadata) error {
	p0, err := r.manager.GetPolicy(p.ID)
	if err != nil {
		return err
	}
	needUn := false
	needSch := false

	if p0.Trigger.Kind != p.Trigger.Kind {
		if p0.Trigger.Kind == policy.TriggerKindSchedule {
			needUn = true
		}

		if p.Trigger.Kind == policy.TriggerKindSchedule {
			needSch = true
		}
	} else {
		switch p.Trigger.Kind {
		case policy.TriggerKindSchedule:
			if p0.Trigger.Settings["cron"] != p.Trigger.Settings["cron"] {
				// unschedule old
				if len(p0.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
					needUn = true
				}
				// schedule new
				if len(p.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
					// valid cron
					needSch = true
				}
			}
		case "":

		default:
			return fmt.Errorf("not support Trigger %s", p.Trigger.Kind)
		}
	}
	if needUn {
		err = r.scheduler.UnSchedule(orm.Context(), p0.Trigger.References[policy.TriggerReferencesJobid].(int64))
		if err != nil {
			return err
		}
	}
	if needSch {
		jobid, err := r.scheduler.Schedule(orm.Context(), p.Trigger.Settings[policy.TriggerSettingsCron].(string), SchedulerCallback, TriggerParam{
			PolicyID: p.ID,
			Trigger:  ExecutionTriggerSchedule,
		})
		if err != nil {
			return err
		}
		p.Trigger.References[policy.TriggerReferencesJobid] = jobid
	}

	return r.manager.UpdatePolicy(p)
}

// DeleteRetention Delete Retention
func (r *DefaultAPIController) DeleteRetention(id int64) error {
	p, err := r.manager.GetPolicy(id)
	if err != nil {
		return err
	}
	if p.Trigger.Kind == policy.TriggerKindSchedule && len(p.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
		err = r.scheduler.UnSchedule(orm.Context(), p.Trigger.References[policy.TriggerReferencesJobid].(int64))
		if err != nil {
			return err
		}
	}

	return r.manager.DeletePolicyAndExec(id)
}

// TriggerRetentionExec Trigger Retention Execution
func (r *DefaultAPIController) TriggerRetentionExec(policyID int64, trigger string, dryRun bool) (int64, error) {
	p, err := r.manager.GetPolicy(policyID)
	if err != nil {
		return 0, err
	}

	exec := &Execution{
		PolicyID:  policyID,
		StartTime: time.Now().Truncate(time.Second),
		Trigger:   trigger,
		DryRun:    dryRun,
	}
	id, err := r.manager.CreateExecution(exec)
	if _, err = r.launcher.Launch(p, id, dryRun); err != nil {
		// clean execution if launch failed
		_ = r.manager.DeleteExecution(id)
		return 0, err
	}
	return id, err

}

// OperateRetentionExec Operate Retention Execution
func (r *DefaultAPIController) OperateRetentionExec(eid int64, action string) error {
	e, err := r.manager.GetExecution(eid)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("execution %d not found", eid)
	}
	switch action {
	case "stop":
		return r.launcher.Stop(eid)
	default:
		return fmt.Errorf("not support action %s", action)
	}
}

// GetRetentionExec Get Retention Execution
func (r *DefaultAPIController) GetRetentionExec(executionID int64) (*Execution, error) {
	return r.manager.GetExecution(executionID)
}

// ListRetentionExecs List Retention Executions
func (r *DefaultAPIController) ListRetentionExecs(policyID int64, query *q.Query) ([]*Execution, error) {
	return r.manager.ListExecutions(policyID, query)
}

// GetTotalOfRetentionExecs Count Retention Executions
func (r *DefaultAPIController) GetTotalOfRetentionExecs(policyID int64) (int64, error) {
	return r.manager.GetTotalOfRetentionExecs(policyID)
}

// ListRetentionExecTasks List Retention Execution Histories
func (r *DefaultAPIController) ListRetentionExecTasks(executionID int64, query *q.Query) ([]*Task, error) {
	q1 := &q.TaskQuery{
		ExecutionID: executionID,
	}
	if query != nil {
		q1.PageSize = query.PageSize
		q1.PageNumber = query.PageNumber
	}
	return r.manager.ListTasks(q1)
}

// GetTotalOfRetentionExecTasks Count Retention Execution Histories
func (r *DefaultAPIController) GetTotalOfRetentionExecTasks(executionID int64) (int64, error) {
	return r.manager.GetTotalOfTasks(executionID)
}

// GetRetentionExecTaskLog Get Retention Execution Task Log
func (r *DefaultAPIController) GetRetentionExecTaskLog(taskID int64) ([]byte, error) {
	return r.manager.GetTaskLog(taskID)
}

// NewAPIController ...
func NewAPIController(retentionMgr Manager, projectManager project.Manager, repositoryMgr repository.Manager, scheduler scheduler.Scheduler, retentionLauncher Launcher) APIController {
	return &DefaultAPIController{
		manager:        retentionMgr,
		launcher:       retentionLauncher,
		projectManager: projectManager,
		repositoryMgr:  repositoryMgr,
		scheduler:      scheduler,
	}
}
