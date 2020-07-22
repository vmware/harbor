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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// MetadataAPI ...
type MetadataAPI struct {
	BaseController
	metaMgr metamgr.ProjectMetadataManager
	project *models.Project
	name    string
}

// Prepare ...
func (m *MetadataAPI) Prepare() {
	m.BaseController.Prepare()

	m.metaMgr = m.ProjectMgr.GetMetadataManager()

	// the project manager doesn't use a project metadata manager
	if m.metaMgr == nil {
		log.Debug("the project manager doesn't use a project metadata manager")
		m.RenderError(http.StatusMethodNotAllowed, "")
		return
	}

	id, err := m.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", id)
		}
		m.SendBadRequestError(errors.New(text))
		return
	}

	project, err := m.ProjectMgr.Get(id)
	if err != nil {
		m.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
		return
	}

	if project == nil {
		m.handleProjectNotFound(id)
		return
	}

	m.project = project

	name := m.GetStringFromPath(":name")
	if len(name) > 0 {
		m.name = name
		metas, err := m.metaMgr.Get(project.ProjectID, name)
		if err != nil {
			m.SendInternalServerError(fmt.Errorf("failed to get metadata of project %d: %v", project.ProjectID, err))
			return
		}
		if len(metas) == 0 {
			m.SendNotFoundError(fmt.Errorf("metadata %s of project %d not found", name, project.ProjectID))
			return
		}
	}
}

func (m *MetadataAPI) requireAccess(action rbac.Action) bool {
	return m.RequireProjectAccess(m.project.ProjectID, action, rbac.ResourceMetadata)
}

// Get ...
func (m *MetadataAPI) Get() {
	if !m.requireAccess(rbac.ActionRead) {
		return
	}

	var metas map[string]string
	var err error
	if len(m.name) > 0 {
		metas, err = m.metaMgr.Get(m.project.ProjectID, m.name)
	} else {
		metas, err = m.metaMgr.Get(m.project.ProjectID)
	}

	if err != nil {
		m.SendInternalServerError(fmt.Errorf("failed to get metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
	m.Data["json"] = metas
	m.ServeJSON()
}

// Post ...
func (m *MetadataAPI) Post() {
	if !m.requireAccess(rbac.ActionCreate) {
		return
	}

	var metas map[string]string
	if err := m.DecodeJSONReq(&metas); err != nil {
		m.SendBadRequestError(err)
		return
	}

	ms, err := validateProjectMetadata(metas)
	if err != nil {
		m.SendBadRequestError(err)
		return
	}

	if len(ms) != 1 {
		m.SendBadRequestError(errors.New("invalid request: has no valid key/value pairs or has more than one valid key/value pairs"))
		return
	}

	keys := reflect.ValueOf(ms).MapKeys()
	mts, err := m.metaMgr.Get(m.project.ProjectID, keys[0].String())
	if err != nil {
		m.SendInternalServerError(fmt.Errorf("failed to get metadata for project %d: %v", m.project.ProjectID, err))
		return
	}

	if len(mts) != 0 {
		m.SendConflictError(errors.New("conflict metadata"))
		return
	}

	if err := m.metaMgr.Add(m.project.ProjectID, ms); err != nil {
		m.SendInternalServerError(fmt.Errorf("failed to create metadata for project %d: %v", m.project.ProjectID, err))
		return
	}

	m.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
}

// Put ...
func (m *MetadataAPI) Put() {
	if !m.requireAccess(rbac.ActionUpdate) {
		return
	}

	var metas map[string]string
	if err := m.DecodeJSONReq(&metas); err != nil {
		m.SendBadRequestError(err)
		return
	}

	meta, exist := metas[m.name]
	if !exist {
		m.SendBadRequestError(fmt.Errorf("must contains key %s", m.name))
		return
	}

	ms, err := validateProjectMetadata(map[string]string{
		m.name: meta,
	})
	if err != nil {
		m.SendBadRequestError(err)
		return
	}

	if err := m.metaMgr.Update(m.project.ProjectID, map[string]string{
		m.name: ms[m.name],
	}); err != nil {
		m.SendInternalServerError(fmt.Errorf("failed to update metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
}

// Delete ...
func (m *MetadataAPI) Delete() {
	if !m.requireAccess(rbac.ActionDelete) {
		return
	}

	if err := m.metaMgr.Delete(m.project.ProjectID, m.name); err != nil {
		m.SendInternalServerError(fmt.Errorf("failed to delete metadata %s of project %d: %v", m.name, m.project.ProjectID, err))
		return
	}
}

// validate metas and return a new map which contains the valid key/value pairs only
func validateProjectMetadata(metas map[string]string) (map[string]string, error) {
	if len(metas) == 0 {
		return nil, nil
	}

	boolMetas := []string{
		models.ProMetaPublic,
		models.ProMetaEnableContentTrust,
		models.ProMetaPreventVul,
		models.ProMetaAutoScan}

	for _, boolMeta := range boolMetas {
		value, exist := metas[boolMeta]
		if exist {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s to bool: %v", value, err)
			}
			metas[boolMeta] = strconv.FormatBool(b)
		}
	}

	value, exist := metas[models.ProMetaSeverity]
	if exist {
		severity := vuln.ParseSeverityVersion3(strings.ToLower(value))
		if severity == vuln.Unknown {
			return nil, fmt.Errorf("invalid severity %s", value)
		}

		metas[models.ProMetaSeverity] = strings.ToLower(severity.String())
	}

	return metas, nil
}
