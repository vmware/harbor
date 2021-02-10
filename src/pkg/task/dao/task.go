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

package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/lib/pq"
)

// TaskDAO is the data access object interface for task
type TaskDAO interface {
	// Count returns the total count of tasks according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List the tasks according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	List(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// Get the specified task
	Get(ctx context.Context, id int64) (task *Task, err error)
	// Create a task
	Create(ctx context.Context, task *Task) (id int64, err error)
	// Update the specified task. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, task *Task, props ...string) (err error)
	// UpdateStatus updates the status of task
	UpdateStatus(ctx context.Context, id int64, status string, statusRevision int64) (err error)
	// Delete the specified task
	Delete(ctx context.Context, id int64) (err error)
	// ListStatusCount lists the status count for the tasks reference the specified execution
	ListStatusCount(ctx context.Context, executionID int64) (statusCounts []*StatusCount, err error)
	// GetMaxEndTime gets the max end time for the tasks references the specified execution
	GetMaxEndTime(ctx context.Context, executionID int64) (endTime time.Time, err error)
}

// NewTaskDAO returns an instance of TaskDAO
func NewTaskDAO() TaskDAO {
	return &taskDAO{}
}

type taskDAO struct{}

func (t *taskDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := t.querySetter(ctx, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (t *taskDAO) List(ctx context.Context, query *q.Query) ([]*Task, error) {
	tasks := []*Task{}
	qs, err := t.querySetter(ctx, query)
	if err != nil {
		return nil, err
	}
	qs = qs.OrderBy("-StartTime")
	if _, err = qs.All(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (t *taskDAO) Get(ctx context.Context, id int64) (*Task, error) {
	task := &Task{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(task); err != nil {
		if e := orm.AsNotFoundError(err, "task %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return task, nil
}

func (t *taskDAO) Create(ctx context.Context, task *Task) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(task)
	if err != nil {
		if e := orm.AsForeignKeyError(err,
			"the task tries to reference a non existing execution %d", task.ExecutionID); e != nil {
			err = e
		}
		return 0, err
	}
	return id, nil
}

func (t *taskDAO) Update(ctx context.Context, task *Task, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(task, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("task %d not found", task.ID)
	}
	return nil
}

func (t *taskDAO) UpdateStatus(ctx context.Context, id int64, status string, statusRevision int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// status revision is the unix timestamp of job starting time, it's changing means a retrying of the job
	startTime := time.Unix(statusRevision, 0)
	// update run count and start time when status revision changes
	sql := `update task set run_count = run_count +1, start_time = ? 
				where id = ? and status_revision < ?`
	if _, err = ormer.Raw(sql, startTime, id, statusRevision).Exec(); err != nil {
		return err
	}

	jobStatus := job.Status(status)
	statusCode := jobStatus.Code()
	var endTime time.Time
	now := time.Now()
	// when the task is in final status, update the end time
	// when the task re-runs again, the end time should be cleared, so set the end time
	// to null if the task isn't in final status
	if jobStatus.Final() {
		endTime = now
	}
	// use raw sql rather than the ORM as the sql generated by ORM isn't a "single" statement
	// which means the operation isn't atomic, this will cause issues when running in concurrency
	sql = `update task set status = ?, status_code = ?, status_revision = ?, update_time = ?, end_time = ? 
		where id = ? and (status_revision = ? and status_code < ? or status_revision < ?) `
	_, err = ormer.Raw(sql, status, statusCode, statusRevision, now, endTime,
		id, statusRevision, statusCode, statusRevision).Exec()
	return err
}

func (t *taskDAO) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Task{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("task %d not found", id)
	}
	return nil
}

func (t *taskDAO) ListStatusCount(ctx context.Context, executionID int64) ([]*StatusCount, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	statusCounts := []*StatusCount{}
	_, err = ormer.Raw("select status, count(*) as count from task where execution_id=? group by status", executionID).
		QueryRows(&statusCounts)
	if err != nil {
		return nil, err
	}
	return statusCounts, nil
}

func (t *taskDAO) GetMaxEndTime(ctx context.Context, executionID int64) (time.Time, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return time.Time{}, err
	}
	var endTime time.Time
	err = ormer.Raw("select max(end_time) from task where execution_id = ?", executionID).
		QueryRow(&endTime)
	return endTime, nil
}

func (t *taskDAO) querySetter(ctx context.Context, query *q.Query) (orm.QuerySeter, error) {
	qs, err := orm.QuerySetter(ctx, &Task{}, query)
	if err != nil {
		return nil, err
	}

	// append the filter for "extra attrs"
	if query != nil && len(query.Keywords) > 0 {
		var (
			key       string
			keyPrefix string
			value     interface{}
		)
		for key, value = range query.Keywords {
			if strings.HasPrefix(key, "ExtraAttrs.") {
				keyPrefix = "ExtraAttrs."
				break
			}
			if strings.HasPrefix(key, "extra_attrs.") {
				keyPrefix = "extra_attrs."
				break
			}
		}
		if len(keyPrefix) == 0 {
			return qs, nil
		}
		// avoid the sql injection
		key = pq.QuoteLiteral(key)
		if _, ok := value.(string); ok {
			value = value.(string)
		}
		qs = qs.FilterRaw("id", fmt.Sprintf("in (select id from task where extra_attrs->>%s=%v)", key, value))
	}

	return qs, nil
}
