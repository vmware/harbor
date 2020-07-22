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

package orm

import (
	"context"
	"errors"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/log"
)

type ormKey struct{}

// FromContext returns orm from context
func FromContext(ctx context.Context) (orm.Ormer, error) {
	o, ok := ctx.Value(ormKey{}).(orm.Ormer)
	if !ok {
		return nil, errors.New("cannot get the ORM from context")
	}
	return o, nil
}

// NewContext returns new context with orm
func NewContext(ctx context.Context, o orm.Ormer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ormKey{}, o)
}

// Context returns a context with an orm
func Context() context.Context {
	return NewContext(context.Background(), orm.NewOrm())
}

// WithTransaction a decorator which make f run in transaction
func WithTransaction(f func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		o, err := FromContext(ctx)
		if err != nil {
			return err
		}

		tx := ormerTx{Ormer: o}
		if err := tx.Begin(); err != nil {
			log.Errorf("begin transaction failed: %v", err)
			return err
		}

		if err := f(ctx); err != nil {
			if e := tx.Rollback(); e != nil {
				log.Errorf("rollback transaction failed: %v", e)
				return e
			}

			return err
		}

		if err := tx.Commit(); err != nil {
			log.Errorf("commit transaction failed: %v", err)
			return err
		}

		return nil
	}
}
