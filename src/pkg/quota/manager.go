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

package quota

import (
	"context"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/quota/dao"
	"github.com/goharbor/harbor/src/pkg/quota/models"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// Quota alias `models.Quota` to make it natural to use the Manager
type Quota = models.Quota

// Manager interface provide the management functions for quotas
type Manager interface {
	// Create create quota for the reference object
	Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, usages ...types.ResourceList) (int64, error)

	// Count returns the total count of quotas according to the query.
	Count(ctx context.Context, query *q.Query) (int64, error)

	// Delete delete quota by id
	Delete(ctx context.Context, id int64) error

	// Get returns quota by id
	Get(ctx context.Context, id int64) (*Quota, error)

	// GetByRef returns quota by reference object
	GetByRef(ctx context.Context, reference, referenceID string) (*Quota, error)

	// GetByRefForUpdate returns quota by reference and reference id for update
	GetByRefForUpdate(ctx context.Context, reference, referenceID string) (*Quota, error)

	// Update update quota
	Update(ctx context.Context, quota *Quota) error

	// List list quotas
	List(ctx context.Context, query *q.Query) ([]*Quota, error)
}

var (
	// Mgr default quota manager
	Mgr = NewManager()
)

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, usages ...types.ResourceList) (id int64, err error) {
	h := func(ctx context.Context) error {
		var used types.ResourceList
		if len(usages) > 0 {
			used = usages[0]
		} else {
			used = types.Zero(hardLimits)
		}

		id, err = m.dao.Create(ctx, reference, referenceID, hardLimits, used)

		return err
	}

	err = orm.WithTransaction(h)(ctx)

	return id, err
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	h := func(ctx context.Context) error {
		return m.dao.Delete(ctx, id)
	}

	return orm.WithTransaction(h)(ctx)
}

func (m *manager) Get(ctx context.Context, id int64) (*Quota, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) GetByRef(ctx context.Context, reference, referenceID string) (*Quota, error) {
	return m.dao.GetByRef(ctx, reference, referenceID)
}

func (m *manager) GetByRefForUpdate(ctx context.Context, reference, referenceID string) (*Quota, error) {
	return m.dao.GetByRefForUpdate(ctx, reference, referenceID)
}

func (m *manager) Update(ctx context.Context, q *Quota) error {
	h := func(ctx context.Context) error {
		return m.dao.Update(ctx, q)
	}

	return orm.WithTransaction(h)(ctx)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*Quota, error) {
	return m.dao.List(ctx, query)
}

// NewManager returns quota manager
func NewManager() Manager {
	return &manager{dao: dao.New()}
}
