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

package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// Quota quota model for manager
type Quota struct {
	ID           int64            `orm:"pk;auto;column(id)" json:"id"`
	Ref          driver.RefObject `orm:"-" json:"ref"`
	Reference    string           `orm:"column(reference)" json:"-"`
	ReferenceID  string           `orm:"column(reference_id)" json:"-"`
	Hard         string           `orm:"column(hard);type(jsonb)" json:"-"`
	Used         string           `orm:"column(used);type(jsonb)" json:"-"`
	CreationTime time.Time        `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time        `orm:"column(update_time);auto_now" json:"update_time"`

	HardChanged bool `orm:"-" json:"-"`
	UsedChanged bool `orm:"-" json:"-"`
}

// MarshalJSON ...
func (q *Quota) MarshalJSON() ([]byte, error) {
	hard, err := types.NewResourceList(q.Hard)
	if err != nil {
		return nil, err
	}

	used, err := types.NewResourceList(q.Used)
	if err != nil {
		return nil, err
	}

	type Alias Quota
	return json.Marshal(&struct {
		*Alias
		Hard types.ResourceList `json:"hard"`
		Used types.ResourceList `json:"used"`
	}{
		Alias: (*Alias)(q),
		Hard:  hard,
		Used:  used,
	})
}

// GetHard returns quota hard
func (q *Quota) GetHard() (types.ResourceList, error) {
	return types.NewResourceList(q.Hard)
}

// SetHard set hard value of the quota
func (q *Quota) SetHard(hardLimits types.ResourceList) *Quota {
	q.HardChanged = true
	q.Hard = hardLimits.String()

	return q
}

// GetUsed returns quota used
func (q *Quota) GetUsed() (types.ResourceList, error) {
	return types.NewResourceList(q.Used)
}

// SetUsed set used value of the quota
func (q *Quota) SetUsed(used types.ResourceList) *Quota {
	q.UsedChanged = true
	q.Used = used.String()

	return q
}

// GetWarningResources returns resource names which exceeded the warning percent
func (q *Quota) GetWarningResources(warningPercent int) ([]types.ResourceName, error) {
	if warningPercent < 0 || warningPercent > 100 {
		return nil, fmt.Errorf("bad warningPercent")
	}

	hardLimits, err := q.GetHard()
	if err != nil {
		return nil, err
	}

	usage, err := q.GetUsed()
	if err != nil {
		return nil, err
	}

	var resources []types.ResourceName
	for resource, used := range usage {
		limited, ok := hardLimits[resource]
		if !ok {
			return nil, fmt.Errorf("resource %s not found in hard limits", resource)
		}

		if limited == types.UNLIMITED {
			continue
		}

		// used / limited >= warningPercent / 100
		if used*100 >= limited*int64(warningPercent) {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}
