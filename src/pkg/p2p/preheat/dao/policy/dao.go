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

package policy

import (
	"context"

	beego_orm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
)

// DAO is the data access object for policy.
type DAO interface {
	// Count returns the total count of policies according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// Create the policy schema
	Create(ctx context.Context, schema *policy.Schema) (id int64, err error)
	// Update the policy schema, Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, schema *policy.Schema, props ...string) (err error)
	// Get the policy schema by id
	Get(ctx context.Context, id int64) (schema *policy.Schema, err error)
	// Get the policy schema by name
	GetByName(ctx context.Context, projectID int64, name string) (schema *policy.Schema, err error)
	// Delete the policy schema by id
	Delete(ctx context.Context, id int64) (err error)
	// List policy schemas by query
	List(ctx context.Context, query *q.Query) (schemas []*policy.Schema, err error)
}

// New returns an instance of the default DAO.
func New() DAO {
	return &dao{}
}

type dao struct{}

// Count returns the total count of policies according to the query
func (d *dao) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}

	qs, err := orm.QuerySetter(ctx, &policy.Schema{}, query)
	if err != nil {
		return 0, err
	}

	return qs.Count()
}

// Create a policy schema.
func (d *dao) Create(ctx context.Context, schema *policy.Schema) (id int64, err error) {
	var ormer beego_orm.Ormer
	ormer, err = orm.FromContext(ctx)
	if err != nil {
		return
	}

	id, err = ormer.Insert(schema)
	if err != nil {
		if e := orm.AsConflictError(err, "policy %s already exists", schema.Name); e != nil {
			err = e
		}
		return
	}

	return
}

// Update a policy schema.
func (d *dao) Update(ctx context.Context, schema *policy.Schema, props ...string) (err error) {
	var ormer beego_orm.Ormer
	ormer, err = orm.FromContext(ctx)
	if err != nil {
		return err
	}

	id, err := ormer.Update(schema, props...)
	if err != nil {
		return err
	}

	if id == 0 {
		return errors.NotFoundError(nil).WithMessage("policy %d not found", schema.ID)
	}

	return nil
}

// Get a policy schema by id.
func (d *dao) Get(ctx context.Context, id int64) (schema *policy.Schema, err error) {
	var ormer beego_orm.Ormer
	ormer, err = orm.FromContext(ctx)
	if err != nil {
		return
	}

	schema = &policy.Schema{ID: id}
	if err = ormer.Read(schema); err != nil {
		if e := orm.AsNotFoundError(err, "policy %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}

	return schema, nil
}

// GetByName gets a policy schema by name.
func (d *dao) GetByName(ctx context.Context, projectID int64, name string) (schema *policy.Schema, err error) {
	var ormer beego_orm.Ormer
	ormer, err = orm.FromContext(ctx)
	if err != nil {
		return
	}

	schema = &policy.Schema{Name: name, ProjectID: projectID}
	if err = ormer.Read(schema, "Name", "ProjectID"); err != nil {
		if e := orm.AsNotFoundError(err, "policy %s not found", name); e != nil {
			err = e
		}
		return nil, err
	}

	return schema, nil
}

// Delete a policy schema by id.
func (d *dao) Delete(ctx context.Context, id int64) (err error) {
	var ormer beego_orm.Ormer
	ormer, err = orm.FromContext(ctx)
	if err != nil {
		return
	}

	n, err := ormer.Delete(&policy.Schema{
		ID: id,
	})
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("policy %d not found", id)
	}

	return nil
}

// List policies by query.
func (d *dao) List(ctx context.Context, query *q.Query) (schemas []*policy.Schema, err error) {
	var qs beego_orm.QuerySeter
	qs, err = orm.QuerySetter(ctx, &policy.Schema{}, query)
	if err != nil {
		return
	}

	qs = qs.OrderBy("UpdatedTime", "ID")
	if _, err = qs.All(&schemas); err != nil {
		return
	}

	return schemas, nil
}
