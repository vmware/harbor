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

package provider

import (
	"encoding/json"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// PreheatingImageTypeImage defines the 'image' type of preheating images
	PreheatingImageTypeImage = "image"
	// PreheatingStatusPending means the preheating is waiting for starting
	PreheatingStatusPending = "PENDING"
	// PreheatingStatusRunning means the preheating is ongoing
	PreheatingStatusRunning = "RUNNING"
	// PreheatingStatusSuccess means the preheating is success
	PreheatingStatusSuccess = "SUCCESS"
	// PreheatingStatusFail means the preheating is failed
	PreheatingStatusFail = "FAIL"
)

func init() {
	orm.RegisterModel(&Instance{})
}

// Instance defines the properties of the preheating provider instance.
type Instance struct {
	ID          int64  `orm:"pk;auto;column(id)" json:"id"`
	Name        string `orm:"column(name)" json:"name"`
	Description string `orm:"column(description)" json:"description"`
	Vendor      string `orm:"column(vendor)" json:"vendor"`
	Endpoint    string `orm:"column(endpoint)" json:"endpoint"`
	AuthMode    string `orm:"column(auth_mode)" json:"auth_mode"`
	// The auth credential data if exists
	AuthInfo map[string]string `orm:"-" json:"auth_info,omitempty"`
	// Data format for "AuthInfo"
	AuthData string `orm:"column(auth_data)" json:"-"`
	// Default 'Unknown', use separate API for client to retrieve
	Status         string `orm:"-" json:"status"`
	Enabled        bool   `orm:"column(enabled)" json:"enabled"`
	Default        bool   `orm:"column(is_default)" json:"default"`
	Insecure       bool   `orm:"column(insecure)" json:"insecure"`
	SetupTimestamp int64  `orm:"column(setup_timestamp)" json:"setup_timestamp"`
}

// FromJSON build instance from the given data.
func (ins *Instance) FromJSON(data string) error {
	if len(data) == 0 {
		return errors.New("empty JSON data")
	}

	if err := json.Unmarshal([]byte(data), ins); err != nil {
		return errors.Wrap(err, "construct preheat instance error")
	}

	return nil
}

// ToJSON encodes the instance to JSON data.
func (ins *Instance) ToJSON() (string, error) {
	data, err := json.Marshal(ins)
	if err != nil {
		return "", errors.Wrap(err, "encode preheat instance error")
	}

	return string(data), nil
}

// TableName ...
func (ins *Instance) TableName() string {
	return "p2p_preheat_instance"
}
