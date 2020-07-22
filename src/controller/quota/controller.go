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
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	redislib "github.com/goharbor/harbor/src/lib/redis"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/gomodule/redigo/redis"

	// quota driver
	_ "github.com/goharbor/harbor/src/controller/quota/driver"
)

var (
	// expire reserved resources when no actions on the key of the reserved resources in redis during 1 hour
	defaultReservedExpiration = time.Hour
)

var (
	// Ctl is a global quota controller instance
	Ctl = NewController()
)

// Controller defines the operations related with quotas
type Controller interface {
	// Count returns the total count of quotas according to the query.
	Count(ctx context.Context, query *q.Query) (int64, error)

	// Create ensure quota for the reference object
	Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error)

	// Delete delete quota by id
	Delete(ctx context.Context, id int64) error

	// Get returns quota by id
	Get(ctx context.Context, id int64, options ...Option) (*quota.Quota, error)

	// GetByRef returns quota by reference object
	GetByRef(ctx context.Context, reference, referenceID string, options ...Option) (*quota.Quota, error)

	// IsEnabled returns true when quota enabled for reference object
	IsEnabled(ctx context.Context, reference, referenceID string) (bool, error)

	// List list quotas
	List(ctx context.Context, query *q.Query, options ...Option) ([]*quota.Quota, error)

	// Refresh refresh quota for the reference object
	Refresh(ctx context.Context, reference, referenceID string, options ...Option) error

	// Request request resources to run f
	// Before run the function, it reserves the resources,
	// then runs f and refresh quota when f success，
	// in the finally it releases the resources which reserved at the beginning.
	Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error

	// Update update quota
	Update(ctx context.Context, q *quota.Quota) error
}

// NewController creates an instance of the default quota controller
func NewController() Controller {
	return &controller{
		reservedExpiration: defaultReservedExpiration,
		quotaMgr:           quota.Mgr,
	}
}

type controller struct {
	reservedExpiration time.Duration

	quotaMgr quota.Manager
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.quotaMgr.Count(ctx, query)
}

func (c *controller) Create(ctx context.Context, reference, referenceID string, hardLimits types.ResourceList, used ...types.ResourceList) (int64, error) {
	return c.quotaMgr.Create(ctx, reference, referenceID, hardLimits, used...)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.quotaMgr.Delete(ctx, id)
}

func (c *controller) Get(ctx context.Context, id int64, options ...Option) (*quota.Quota, error) {
	q, err := c.quotaMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.assembleQuota(ctx, q, newOptions(options...))
}

func (c *controller) GetByRef(ctx context.Context, reference, referenceID string, options ...Option) (*quota.Quota, error) {
	q, err := c.quotaMgr.GetByRef(ctx, reference, referenceID)
	if err != nil {
		return nil, err
	}

	return c.assembleQuota(ctx, q, newOptions(options...))
}

func (c *controller) assembleQuota(ctx context.Context, q *quota.Quota, opts *Options) (*quota.Quota, error) {
	if opts.WithReferenceObject {
		driver, err := Driver(ctx, q.Reference)
		if err != nil {
			return nil, err
		}

		ref, err := driver.Load(ctx, q.ReferenceID)
		if err != nil {
			return nil, err
		}

		q.Ref = ref
	}

	return q, nil
}

func (c *controller) IsEnabled(ctx context.Context, reference, referenceID string) (bool, error) {
	d, err := Driver(ctx, reference)
	if err != nil {
		return false, err
	}

	return d.Enabled(ctx, referenceID)
}

func (c *controller) List(ctx context.Context, query *q.Query, options ...Option) ([]*quota.Quota, error) {
	quotas, err := c.quotaMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	for _, q := range quotas {
		if _, err := c.assembleQuota(ctx, q, opts); err != nil {
			return nil, err
		}
	}

	return quotas, nil
}

func (c *controller) getReservedResources(ctx context.Context, reference, referenceID string) (types.ResourceList, error) {
	conn := redislib.DefaultPool().Get()
	defer conn.Close()

	key := reservedResourcesKey(reference, referenceID)

	str, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return types.NewResourceList(str)
}

func (c *controller) setReservedResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	conn := redislib.DefaultPool().Get()
	defer conn.Close()

	key := reservedResourcesKey(reference, referenceID)

	reply, err := redis.String(conn.Do("SET", key, resources.String(), "EX", int64(c.reservedExpiration/time.Second)))
	if err != nil {
		return err
	}

	if reply != "OK" {
		return fmt.Errorf("bad reply value")
	}

	return nil
}

func (c *controller) reserveResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	reserve := func(ctx context.Context) error {
		q, err := c.quotaMgr.GetByRefForUpdate(ctx, reference, referenceID)
		if err != nil {
			return err
		}

		hardLimits, err := q.GetHard()
		if err != nil {
			return err
		}

		used, err := q.GetUsed()
		if err != nil {
			return err
		}

		reserved, err := c.getReservedResources(ctx, reference, referenceID)
		if err != nil {
			log.G(ctx).Errorf("failed to get reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		newReserved := types.Add(reserved, resources)

		if err := quota.IsSafe(hardLimits, types.Add(used, reserved), types.Add(used, newReserved), false); err != nil {
			return errors.DeniedError(err).WithMessage("Quota exceeded when processing the request of %v", err)
		}

		if err := c.setReservedResources(ctx, reference, referenceID, newReserved); err != nil {
			log.G(ctx).Errorf("failed to set reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		return nil
	}

	return orm.WithTransaction(reserve)(ctx)
}

func (c *controller) unreserveResources(ctx context.Context, reference, referenceID string, resources types.ResourceList) error {
	unreserve := func(ctx context.Context) error {
		if _, err := c.quotaMgr.GetByRefForUpdate(ctx, reference, referenceID); err != nil {
			return err
		}

		reserved, err := c.getReservedResources(ctx, reference, referenceID)
		if err != nil {
			log.G(ctx).Errorf("failed to get reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		newReserved := types.Subtract(reserved, resources)
		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newReserved); len(negativeUsed) > 0 {
			return fmt.Errorf("reserved resources is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		if err := c.setReservedResources(ctx, reference, referenceID, newReserved); err != nil {
			log.G(ctx).Errorf("failed to set reserved resources for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		return nil
	}

	return orm.WithTransaction(unreserve)(ctx)
}

func (c *controller) Refresh(ctx context.Context, reference, referenceID string, options ...Option) error {
	driver, err := Driver(ctx, reference)
	if err != nil {
		return err
	}

	opts := newOptions(options...)

	refresh := func(ctx context.Context) error {
		q, err := c.quotaMgr.GetByRefForUpdate(ctx, reference, referenceID)
		if err != nil {
			return err
		}

		hardLimits, err := q.GetHard()
		if err != nil {
			return err
		}

		used, err := q.GetUsed()
		if err != nil {
			return err
		}

		newUsed, err := driver.CalculateUsage(ctx, referenceID)
		if err != nil {
			log.G(ctx).Errorf("failed to calculate quota usage for %s %s, error: %v", reference, referenceID, err)
			return err
		}

		// ensure that new used is never negative
		if negativeUsed := types.IsNegative(newUsed); len(negativeUsed) > 0 {
			return fmt.Errorf("quota usage is negative for resource(s): %s", quota.PrettyPrintResourceNames(negativeUsed))
		}

		if err := quota.IsSafe(hardLimits, used, newUsed, opts.IgnoreLimitation); err != nil {
			return err
		}

		q.SetUsed(newUsed)
		q.UpdateTime = time.Now()

		return c.quotaMgr.Update(ctx, q)
	}

	return orm.WithTransaction(refresh)(ctx)
}

func (c *controller) Request(ctx context.Context, reference, referenceID string, resources types.ResourceList, f func() error) error {
	if len(resources) == 0 {
		return f()
	}

	if err := c.reserveResources(ctx, reference, referenceID, resources); err != nil {
		return err
	}

	defer func() {
		if err := c.unreserveResources(ctx, reference, referenceID, resources); err != nil {
			// ignore this error because reserved resources will be expired
			// when no actions on the key of the reserved resources in redis during sometimes
			log.G(ctx).Warningf("unreserve resources %s for %s %s failed, error: %v", resources.String(), reference, referenceID, err)
		}
	}()

	if err := f(); err != nil {
		return err
	}

	return c.Refresh(ctx, reference, referenceID)
}

func (c *controller) Update(ctx context.Context, u *quota.Quota) error {
	update := func(ctx context.Context) error {
		q, err := c.quotaMgr.GetByRefForUpdate(ctx, u.Reference, u.ReferenceID)
		if err != nil {
			return err
		}

		if q.Hard != u.Hard {
			if hard, err := u.GetHard(); err == nil {
				q.SetHard(hard)
			}
		}

		if q.Used != u.Used {
			if used, err := u.GetUsed(); err == nil {
				q.SetUsed(used)
			}
		}

		q.UpdateTime = time.Now()
		return c.quotaMgr.Update(ctx, q)
	}

	return orm.WithTransaction(update)(ctx)
}

// Driver returns quota driver for the reference
func Driver(ctx context.Context, reference string) (driver.Driver, error) {
	d, ok := driver.Get(reference)
	if !ok {
		return nil, fmt.Errorf("quota not support for %s", reference)
	}

	return d, nil
}

// Validate validate hard limits
func Validate(ctx context.Context, reference string, hardLimits types.ResourceList) error {
	d, err := Driver(ctx, reference)
	if err != nil {
		return err
	}

	return d.Validate(hardLimits)
}

func reservedResourcesKey(reference, referenceID string) string {
	return fmt.Sprintf("quota:%s:%s:reserved", reference, referenceID)
}
