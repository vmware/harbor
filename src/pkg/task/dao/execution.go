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

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// ExecutionDAO is the data access object interface for execution
type ExecutionDAO interface {
	// Count returns the total count of executions according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List the executions according to the query
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// Create an execution
	Create(ctx context.Context, execution *Execution) (id int64, err error)
	// Update the specified execution. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, execution *Execution, props ...string) (err error)
	// Delete the specified execution
	Delete(ctx context.Context, id int64) (err error)
}

// NewExecutionDAO returns an instance of ExecutionDAO
func NewExecutionDAO() ExecutionDAO {
	return &executionDAO{}
}

type executionDAO struct{}

func (e *executionDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := orm.QuerySetter(ctx, &Execution{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (e *executionDAO) List(ctx context.Context, query *q.Query) ([]*Execution, error) {
	executions := []*Execution{}
	qs, err := orm.QuerySetter(ctx, &Execution{}, query)
	if err != nil {
		return nil, err
	}
	qs = qs.OrderBy("-StartTime")
	if _, err = qs.All(&executions); err != nil {
		return nil, err
	}
	return executions, nil
}

func (e *executionDAO) Get(ctx context.Context, id int64) (*Execution, error) {
	execution := &Execution{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(execution); err != nil {
		if e := orm.AsNotFoundError(err, "execution %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return execution, nil
}

func (e *executionDAO) Create(ctx context.Context, execution *Execution) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return ormer.Insert(execution)
}

func (e *executionDAO) Update(ctx context.Context, execution *Execution, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(execution, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("execution %d not found", execution.ID)
	}
	return nil
}

func (e *executionDAO) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Execution{
		ID: id,
	})
	if err != nil {
		if e := orm.AsForeignKeyError(err,
			"the execution %d is referenced by other resources", id); e != nil {
			err = e
		}
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("execution %d not found", id)
	}
	return nil
}
