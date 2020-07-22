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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/scheduler"
)

const (
	yamlFileContentType = "application/x-yaml"
	userSessionKey      = "user"
)

// the managers/controllers used globally
var (
	projectMgr          project.Manager
	retentionScheduler  scheduler.Scheduler
	retentionMgr        retention.Manager
	retentionLauncher   retention.Launcher
	retentionController retention.APIController
)

// BaseController ...
type BaseController struct {
	api.BaseAPI
	// SecurityCtx is the security context used to authN &authZ
	SecurityCtx security.Context
	// ProjectMgr is the project manager which abstracts the operations
	// related to projects
	ProjectMgr promgr.ProjectManager
}

// Prepare inits security context and project manager from request
// context
func (b *BaseController) Prepare() {
	ctx, ok := security.FromContext(b.Ctx.Request.Context())
	if !ok {
		log.Errorf("failed to get security context")
		b.SendInternalServerError(errors.New(""))
		return
	}
	b.SecurityCtx = ctx
	b.ProjectMgr = config.GlobalProjectMgr
}

// RequireAuthenticated returns true when the request is authenticated
// otherwise send Unauthorized response and returns false
func (b *BaseController) RequireAuthenticated() bool {
	if !b.SecurityCtx.IsAuthenticated() {
		b.SendError(errors.UnauthorizedError(errors.New("Unauthorized")))
		return false
	}
	return true
}

// HasProjectPermission returns true when the request has action permission on project subresource
func (b *BaseController) HasProjectPermission(projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) (bool, error) {
	_, _, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		return false, err
	}

	project, err := b.ProjectMgr.Get(projectIDOrName)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, errors.NotFoundError(fmt.Errorf("project %v not found", projectIDOrName))
	}

	resource := rbac.NewProjectNamespace(project.ProjectID).Resource(subresource...)
	if !b.SecurityCtx.Can(action, resource) {
		return false, nil
	}

	return true, nil
}

// RequireProjectAccess returns true when the request has action access on project subresource
// otherwise send UnAuthorized or Forbidden response and returns false
func (b *BaseController) RequireProjectAccess(projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) bool {
	hasPermission, err := b.HasProjectPermission(projectIDOrName, action, subresource...)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			b.handleProjectNotFound(projectIDOrName)
		} else {
			b.SendError(err)
		}
		return false
	}

	if !hasPermission {
		b.SendPermissionError()
		return false
	}

	return true
}

// This should be called when a project is not found, if the caller is a system admin it returns 404.
// If it's regular user, it will render permission error
func (b *BaseController) handleProjectNotFound(projectIDOrName interface{}) {
	if b.SecurityCtx.IsSysAdmin() {
		b.SendNotFoundError(fmt.Errorf("project %v not found", projectIDOrName))
	} else {
		b.SendPermissionError()
	}
}

// SendPermissionError is a shortcut for sending different http error based on authentication status.
func (b *BaseController) SendPermissionError() {
	if !b.SecurityCtx.IsAuthenticated() {
		b.SendUnAuthorizedError(errors.New("UnAuthorized"))
	} else {
		b.SendForbiddenError(errors.New(b.SecurityCtx.GetUsername()))
	}

}

// WriteJSONData writes the JSON data to the client.
func (b *BaseController) WriteJSONData(object interface{}) {
	b.Data["json"] = object
	b.ServeJSON()
}

// WriteYamlData writes the yaml data to the client.
func (b *BaseController) WriteYamlData(object interface{}) {
	yData, err := yaml.Marshal(object)
	if err != nil {
		b.SendInternalServerError(err)
		return
	}

	w := b.Ctx.ResponseWriter
	w.Header().Set("Content-Type", yamlFileContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(yData)
}

// PopulateUserSession generates a new session ID and fill the user model in parm to the session
func (b *BaseController) PopulateUserSession(u models.User) {
	b.SessionRegenerateID()
	b.SetSession(userSessionKey, u)
}

// Init related objects/configurations for the API controllers
func Init() error {
	registerHealthCheckers()

	// init chart controller
	if err := initChartController(); err != nil {
		return err
	}

	// init project manager
	initProjectManager()

	retentionMgr = retention.NewManager()

	retentionLauncher = retention.NewLauncher(projectMgr, repository.Mgr, retentionMgr)

	retentionController = retention.NewAPIController(retentionMgr, projectMgr, repository.Mgr, scheduler.Sched, retentionLauncher)

	callbackFun := func(p interface{}) error {
		str, ok := p.(string)
		if !ok {
			return fmt.Errorf("the type of param %v isn't string", p)
		}
		param := &retention.TriggerParam{}
		if err := json.Unmarshal([]byte(str), param); err != nil {
			return fmt.Errorf("failed to unmarshal the param: %v", err)
		}
		_, err := retentionController.TriggerRetentionExec(param.PolicyID, param.Trigger, false)
		return err
	}
	err := scheduler.RegisterCallbackFunc(retention.SchedulerCallback, callbackFun)

	return err
}

func initChartController() error {
	// If chart repository is not enabled then directly return
	if !config.WithChartMuseum() {
		return nil
	}

	chartCtl, err := initializeChartController()
	if err != nil {
		return err
	}

	chartController = chartCtl
	return nil
}

func initProjectManager() {
	projectMgr = project.Mgr
}
