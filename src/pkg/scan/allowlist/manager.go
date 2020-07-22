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

package allowlist

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
)

// Manager defines the interface of CVE allowlist manager, it support both system level and project level allowlists
type Manager interface {
	// CreateEmpty creates empty allowlist for given project
	CreateEmpty(projectID int64) error
	// Set sets the allowlist for given project (create or update)
	Set(projectID int64, list models.CVEAllowlist) error
	// Get gets the allowlist for given project
	Get(projectID int64) (*models.CVEAllowlist, error)
	// SetSys sets system level allowlist
	SetSys(list models.CVEAllowlist) error
	// GetSys gets system level allowlist
	GetSys() (*models.CVEAllowlist, error)
}

type defaultManager struct{}

// CreateEmpty creates empty allowlist for given project
func (d *defaultManager) CreateEmpty(projectID int64) error {
	l := models.CVEAllowlist{
		ProjectID: projectID,
		Items:     []models.CVEAllowlistItem{},
	}
	_, err := dao.CreateCVEAllowlist(l)
	if err != nil {
		logger.Errorf("Failed to create empty CVE allowlist for project: %d, error: %v", projectID, err)
	}
	return err
}

// Set sets the allowlist for given project (create or update)
func (d *defaultManager) Set(projectID int64, list models.CVEAllowlist) error {
	list.ProjectID = projectID
	if err := Validate(list); err != nil {
		return err
	}
	_, err := dao.UpdateCVEAllowlist(list)
	return err
}

// Get gets the allowlist for given project
func (d *defaultManager) Get(projectID int64) (*models.CVEAllowlist, error) {
	wl, err := dao.GetCVEAllowlist(projectID)
	if err != nil {
		return nil, err
	}

	if wl == nil {
		log.Debugf("No CVE allowlist found for project %d, returning empty list.", projectID)
		wl = &models.CVEAllowlist{ProjectID: projectID, Items: []models.CVEAllowlistItem{}}
	} else if wl.Items == nil {
		wl.Items = []models.CVEAllowlistItem{}
	}
	return wl, nil
}

// SetSys sets the system level allowlist
func (d *defaultManager) SetSys(list models.CVEAllowlist) error {
	return d.Set(0, list)
}

// GetSys gets the system level allowlist
func (d *defaultManager) GetSys() (*models.CVEAllowlist, error) {
	return d.Get(0)
}

// NewDefaultManager return a new instance of defaultManager
func NewDefaultManager() Manager {
	return &defaultManager{}
}
