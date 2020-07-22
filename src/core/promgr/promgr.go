// Copyright 2018 Project Harbor Authors
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

package promgr

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/scan/allowlist"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver"
	"github.com/goharbor/harbor/src/lib/log"
)

// ProjectManager is the project manager which abstracts the operations related
// to projects
type ProjectManager interface {
	Get(projectIDOrName interface{}) (*models.Project, error)
	Create(*models.Project) (int64, error)
	Delete(projectIDOrName interface{}) error
	Update(projectIDOrName interface{}, project *models.Project) error
	List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error)
	IsPublic(projectIDOrName interface{}) (bool, error)
	Exists(projectIDOrName interface{}) (bool, error)
	// get all public project
	GetPublic() ([]*models.Project, error)
	// get all projects that the user is authorized
	GetAuthorized(user *models.User) ([]*models.Project, error)
	// if the project manager uses a metadata manager, return it, otherwise return nil
	GetMetadataManager() metamgr.ProjectMetadataManager
}

type defaultProjectManager struct {
	pmsDriver      pmsdriver.PMSDriver
	metaMgrEnabled bool // if metaMgrEnabled is enabled, metaMgr will be used to CURD metadata
	metaMgr        metamgr.ProjectMetadataManager
	allowlistMgr   allowlist.Manager
}

// NewDefaultProjectManager returns an instance of defaultProjectManager,
// if metaMgrEnabled is true, a project metadata manager will be created
// and used to CURD metadata
func NewDefaultProjectManager(driver pmsdriver.PMSDriver, metaMgrEnabled bool) ProjectManager {
	mgr := &defaultProjectManager{
		pmsDriver:      driver,
		metaMgrEnabled: metaMgrEnabled,
	}
	if metaMgrEnabled {
		mgr.metaMgr = metamgr.NewDefaultProjectMetadataManager()
		mgr.allowlistMgr = allowlist.NewDefaultManager()
	}
	return mgr
}

func (d *defaultProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	project, err := d.pmsDriver.Get(projectIDOrName)
	if err != nil {
		return nil, err
	}

	if project != nil && d.metaMgrEnabled {
		meta, err := d.metaMgr.Get(project.ProjectID)
		if err != nil {
			return nil, err
		}
		if len(project.Metadata) == 0 {
			project.Metadata = make(map[string]string)
		}
		for k, v := range meta {
			project.Metadata[k] = v
		}
		wl, err := d.allowlistMgr.Get(project.ProjectID)
		if err != nil {
			return nil, err
		}
		project.CVEAllowlist = *wl
	}
	return project, nil
}
func (d *defaultProjectManager) Create(project *models.Project) (int64, error) {
	id, err := d.pmsDriver.Create(project)
	if err != nil {
		return 0, err
	}
	if d.metaMgrEnabled {
		d.allowlistMgr.CreateEmpty(id)
		if len(project.Metadata) > 0 {
			if err = d.metaMgr.Add(id, project.Metadata); err != nil {
				log.Errorf("failed to add metadata for project %s: %v", project.Name, err)
			}
		}
	}
	return id, nil
}

func (d *defaultProjectManager) Delete(projectIDOrName interface{}) error {
	project, err := d.Get(projectIDOrName)
	if err != nil {
		return err
	}
	if project == nil {
		return nil
	}
	if project.Metadata != nil && d.metaMgrEnabled {
		if err = d.metaMgr.Delete(project.ProjectID); err != nil {
			return err
		}
	}
	return d.pmsDriver.Delete(project.ProjectID)
}

func (d *defaultProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	pro, err := d.Get(projectIDOrName)
	if err != nil {
		return err
	}
	if pro == nil {
		return fmt.Errorf("project %v not found", projectIDOrName)
	}
	// TODO transaction?
	if d.metaMgrEnabled {
		if err := d.allowlistMgr.Set(pro.ProjectID, project.CVEAllowlist); err != nil {
			return err
		}
		if len(project.Metadata) > 0 {
			metaNeedUpdated := map[string]string{}
			metaNeedCreated := map[string]string{}
			if pro.Metadata == nil {
				pro.Metadata = map[string]string{}
			}
			for key, value := range project.Metadata {
				_, exist := pro.Metadata[key]
				if exist {
					metaNeedUpdated[key] = value
				} else {
					metaNeedCreated[key] = value
				}
			}
			if err = d.metaMgr.Add(pro.ProjectID, metaNeedCreated); err != nil {
				return err
			}
			if err = d.metaMgr.Update(pro.ProjectID, metaNeedUpdated); err != nil {
				return err
			}
		}
	}
	return d.pmsDriver.Update(projectIDOrName, project)
}

func (d *defaultProjectManager) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	// query by public/private property with ProjectMetadataManager first
	if d.metaMgrEnabled && query != nil && query.Public != nil {
		projectIDs, err := d.filterByPublic(*query.Public)
		if err != nil {
			return nil, err
		}

		if len(projectIDs) == 0 {
			return &models.ProjectQueryResult{}, nil
		}

		if query.ProjectIDs == nil {
			query.ProjectIDs = projectIDs
		} else {
			query.ProjectIDs = findInBoth(query.ProjectIDs, projectIDs)
		}
	}

	// query by other properties
	result, err := d.pmsDriver.List(query)
	if err != nil {
		return nil, err
	}

	// populate metadata
	if d.metaMgrEnabled {
		for _, project := range result.Projects {
			meta, err := d.metaMgr.Get(project.ProjectID)
			if err != nil {
				return nil, err
			}
			project.Metadata = meta
		}
	}
	// the allowlist is not populated deliberately
	return result, nil
}

func (d *defaultProjectManager) filterByPublic(public bool) ([]int64, error) {
	metas, err := d.metaMgr.List(models.ProMetaPublic, strconv.FormatBool(public))
	if err != nil {
		return nil, err
	}

	projectIDs := []int64{}
	for _, meta := range metas {
		projectIDs = append(projectIDs, meta.ProjectID)
	}
	return projectIDs, nil
}

func findInBoth(ids1 []int64, ids2 []int64) []int64 {
	m := map[int64]struct{}{}
	for _, id := range ids1 {
		m[id] = struct{}{}
	}

	ids := []int64{}
	for _, id := range ids2 {
		if _, exist := m[id]; exist {
			ids = append(ids, id)
		}
	}

	return ids
}

func (d *defaultProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	project, err := d.Get(projectIDOrName)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, nil
	}
	return project.IsPublic(), nil
}

func (d *defaultProjectManager) Exists(projectIDOrName interface{}) (bool, error) {
	project, err := d.Get(projectIDOrName)
	return project != nil, err
}

func (d *defaultProjectManager) GetPublic() ([]*models.Project, error) {
	value := true
	result, err := d.List(&models.ProjectQueryParam{
		Public: &value,
	})
	if err != nil {
		return nil, err
	}
	return result.Projects, nil
}

func (d *defaultProjectManager) GetAuthorized(user *models.User) ([]*models.Project, error) {
	if user == nil {
		return nil, nil
	}
	result, err := d.List(
		&models.ProjectQueryParam{
			Member: &models.MemberQuery{
				Name:     user.Username,
				GroupIDs: user.GroupIDs,
			},
		})
	if err != nil {
		return nil, err
	}
	return result.Projects, nil
}

func (d *defaultProjectManager) GetMetadataManager() metamgr.ProjectMetadataManager {
	return d.metaMgr
}
