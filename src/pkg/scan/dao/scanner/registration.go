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

package scanner

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	liborm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(new(Registration))
}

// AddRegistration adds a new registration
func AddRegistration(r *Registration) (int64, error) {
	o := dao.GetOrmer()

	id, err := o.Insert(r)
	if err != nil {
		return 0, liborm.WrapConflictError(err, "registration name or url already exists")
	}

	return id, nil
}

// GetRegistration gets the specified registration
func GetRegistration(UUID string) (*Registration, error) {
	e := &Registration{}

	o := dao.GetOrmer()
	qs := o.QueryTable(new(Registration))

	if err := qs.Filter("uuid", UUID).One(e); err != nil {
		if err == orm.ErrNoRows {
			// Not existing case
			return nil, nil
		}
		return nil, err
	}

	return e, nil
}

// UpdateRegistration update the specified registration
func UpdateRegistration(r *Registration, cols ...string) error {
	o := dao.GetOrmer()
	count, err := o.Update(r, cols...)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s is updated", r.UUID)
	}

	return nil
}

// DeleteRegistration deletes the registration with the specified UUID
func DeleteRegistration(UUID string) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Registration))

	// delete with query way
	count, err := qt.Filter("uuid", UUID).Delete()

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s is deleted", UUID)
	}

	return nil
}

// ListRegistrations lists all the existing registrations
func ListRegistrations(query *q.Query) ([]*Registration, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Registration))

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				if strings.HasPrefix(k, "ex_") {
					kk := strings.TrimPrefix(k, "ex_")
					qt = qt.Filter(kk, v)
					continue
				}

				qt = qt.Filter(fmt.Sprintf("%s__icontains", k), v)
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			qt = qt.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
		}
	}

	// Order the list
	qt = qt.OrderBy("-is_default", "-create_time")

	l := make([]*Registration, 0)
	_, err := qt.All(&l)

	return l, err
}

// SetDefaultRegistration sets the specified registration as default one
func SetDefaultRegistration(UUID string) error {
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		return err
	}

	var count int64
	qt := o.QueryTable(new(Registration))
	count, err = qt.Filter("uuid", UUID).
		Filter("disabled", false).
		Update(orm.Params{
			"is_default": true,
		})
	if err == nil && count == 0 {
		err = errors.Errorf("set default for %s failed", UUID)
	}

	if err == nil {
		qt2 := o.QueryTable(new(Registration))
		_, err = qt2.Exclude("uuid__exact", UUID).
			Filter("is_default", true).
			Update(orm.Params{
				"is_default": false,
			})
	}

	if err != nil {
		if e := o.Rollback(); e != nil {
			err = errors.Wrap(e, err.Error())
		}
	} else {
		err = o.Commit()
	}

	return err
}

// GetDefaultRegistration gets the default registration
func GetDefaultRegistration() (*Registration, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Registration))

	e := &Registration{}
	if err := qt.Filter("is_default", true).One(e); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return e, nil
}
