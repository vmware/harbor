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
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	pro "github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	errutil "github.com/goharbor/harbor/src/common/utils/error"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	evt "github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/replication"
)

type deletableResp struct {
	Deletable bool   `json:"deletable"`
	Message   string `json:"message"`
}

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseController
	project *models.Project
}

const projectNameMaxLen int = 255
const projectNameMinLen int = 1
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`
const defaultDaysToRetention = 7

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	p.BaseController.Prepare()
	if len(p.GetStringFromPath(":id")) != 0 {
		id, err := p.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid project ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			p.SendBadRequestError(errors.New(text))
			return
		}

		project, err := p.ProjectMgr.Get(id)
		if err != nil {
			p.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
			return
		}

		if project == nil {
			p.handleProjectNotFound(id)
			return
		}

		p.project = project
	}
}

func (p *ProjectAPI) requireAccess(action rbac.Action, subresource ...rbac.Resource) bool {
	if len(subresource) == 0 {
		subresource = append(subresource, rbac.ResourceSelf)
	}

	return p.RequireProjectAccess(p.project.ProjectID, action, subresource...)
}

// Post ...
func (p *ProjectAPI) Post() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}
	onlyAdmin, err := config.OnlyAdminCreateProject()
	if err != nil {
		log.Errorf("failed to determine whether only admin can create projects: %v", err)
		p.SendInternalServerError(fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
		return
	}

	if onlyAdmin && !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can create project")
		p.SendForbiddenError(errors.New("Only system admin can create project"))
		return
	}
	var pro *models.ProjectRequest
	if err := p.DecodeJSONReq(&pro); err != nil {
		p.SendBadRequestError(err)
		return
	}

	err = validateProjectReq(pro)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.SendBadRequestError(fmt.Errorf("invalid request: %v", err))
		return
	}

	// trying to create a proxy cache project
	if pro.RegistryID > 0 {
		// only system admin can create the proxy cache project
		if !p.SecurityCtx.IsSysAdmin() {
			p.SendForbiddenError(errors.New("Only system admin can create proxy cache project"))
			return
		}
		registry, err := replication.RegistryMgr.Get(pro.RegistryID)
		if err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to get the registry %d: %v", pro.RegistryID, err))
			return
		}
		if registry == nil {
			p.SendNotFoundError(fmt.Errorf("registry %d not found", pro.RegistryID))
			return
		}
		permitted := false
		for _, t := range config.GetPermittedRegistryTypesForProxyCache() {
			if string(registry.Type) == t {
				permitted = true
				break
			}
		}
		if !permitted {
			p.SendBadRequestError(fmt.Errorf("unsupported registry type %s", string(registry.Type)))
			return
		}
	}

	var hardLimits types.ResourceList
	if config.QuotaPerProjectEnable() {
		setting, err := config.QuotaSetting()
		if err != nil {
			log.Errorf("failed to get quota setting: %v", err)
			p.SendInternalServerError(fmt.Errorf("failed to get quota setting: %v", err))
			return
		}

		if !p.SecurityCtx.IsSysAdmin() {
			pro.StorageLimit = &setting.StoragePerProject
		}

		hardLimits, err = projectQuotaHardLimits(p.Ctx.Request.Context(), pro, setting)
		if err != nil {
			log.Errorf("Invalid project request, error: %v", err)
			p.SendBadRequestError(fmt.Errorf("invalid request: %v", err))
			return
		}
	}

	exist, err := p.ProjectMgr.Exists(pro.Name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			pro.Name), err)
		return
	}
	if exist {
		p.SendConflictError(errors.New("conflict project"))
		return
	}

	if pro.Metadata == nil {
		pro.Metadata = map[string]string{}
	}
	// accept the "public" property to make replication work well with old versions(<=1.2.0)
	if pro.Public != nil && len(pro.Metadata[models.ProMetaPublic]) == 0 {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(*pro.Public == 1)
	}

	// populate public metadata as false if it isn't set
	if _, ok := pro.Metadata[models.ProMetaPublic]; !ok {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(false)
	}
	// populate

	owner := p.SecurityCtx.GetUsername()
	// set the owner as the system admin when the API being called by replication
	// it's a solution to workaround the restriction of project creation API:
	// only normal users can create projects
	if p.SecurityCtx.IsSolutionUser() {
		user, err := dao.GetUser(models.User{
			UserID: 1,
		})
		if err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to get the user 1: %v", err))
			return
		}
		owner = user.Username
	}
	projectID, err := p.ProjectMgr.Create(&models.Project{
		Name:       pro.Name,
		OwnerName:  owner,
		Metadata:   pro.Metadata,
		RegistryID: pro.RegistryID,
	})
	if err != nil {
		if err == errutil.ErrDupProject {
			log.Debugf("conflict %s", pro.Name)
			p.SendConflictError(fmt.Errorf("conflict %s", pro.Name))
		} else {
			p.ParseAndHandleError("failed to add project", err)
		}
		return
	}

	if config.QuotaPerProjectEnable() {
		ctx := p.Ctx.Request.Context()
		referenceID := quota.ReferenceID(projectID)
		if _, err := quota.Ctl.Create(ctx, quota.ProjectReference, referenceID, hardLimits); err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to create quota for project: %v", err))
			return
		}
	}

	// create a default retention policy for proxy project
	if pro.RegistryID > 0 {
		if err := p.addRetentionPolicyForProxy(projectID); err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to add tag retention policy for project: %v", err))
			return
		}
	}

	// fire event
	evt.BuildAndPublish(&metadata.CreateProjectEventMetadata{
		ProjectID: projectID,
		Project:   pro.Name,
		Operator:  owner,
	})

	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

func (p *ProjectAPI) addRetentionPolicyForProxy(projID int64) error {
	plc := policy.WithNDaysSinceLastPull(projID, defaultDaysToRetention)
	retID, err := retentionController.CreateRetention(plc)
	if err != nil {
		return err
	}
	if err := p.ProjectMgr.GetMetadataManager().Add(projID, map[string]string{"retention_id": strconv.FormatInt(retID, 10)}); err != nil {
		return err
	}
	return nil
}

// Head ...
func (p *ProjectAPI) Head() {

	if !p.SecurityCtx.IsAuthenticated() {
		p.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	name := p.GetString("project_name")
	if len(name) == 0 {
		p.SendBadRequestError(errors.New("project_name is needed"))
		return
	}

	project, err := p.ProjectMgr.Get(name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to get project %s", name), err)
		return
	}

	if project == nil {
		p.SendNotFoundError(fmt.Errorf("project %s not found", name))
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	if !p.requireAccess(rbac.ActionRead) {
		return
	}

	err := p.populateProperties(p.project)
	if err != nil {
		log.Errorf("populate project properties failed with : %+v", err)
	}

	p.Data["json"] = p.project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}
	if !result.Deletable {
		p.SendPreconditionFailedError(errors.New(result.Message))
		return
	}

	if err = p.ProjectMgr.Delete(p.project.ProjectID); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to delete project %d", p.project.ProjectID), err)
		return
	}

	ctx := p.Ctx.Request.Context()
	referenceID := quota.ReferenceID(p.project.ProjectID)
	q, err := quota.Ctl.GetByRef(ctx, quota.ProjectReference, referenceID)
	if err != nil {
		log.Warningf("failed to get quota for project %s, error: %v", p.project.Name, err)
	} else {
		if err := quota.Ctl.Delete(ctx, q.ID); err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to delete quota for project: %v", err))
			return
		}
	}

	// fire event
	evt.BuildAndPublish(&metadata.DeleteProjectEventMetadata{
		ProjectID: p.project.ProjectID,
		Project:   p.project.Name,
		Operator:  p.SecurityCtx.GetUsername(),
	})
}

// Deletable ...
func (p *ProjectAPI) Deletable() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}

	p.Data["json"] = result
	p.ServeJSON()
}

func (p *ProjectAPI) deletable(projectID int64) (*deletableResp, error) {
	count, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
	})
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains repositories, can not be deleted",
		}, nil
	}

	// Check helm charts number
	if config.WithChartMuseum() {
		charts, err := chartController.ListCharts(p.project.Name)
		if err != nil {
			return nil, err
		}

		if len(charts) > 0 {
			return &deletableResp{
				Deletable: false,
				Message:   "the project contains helm charts, can not be deleted",
			}, nil
		}
	}

	return &deletableResp{
		Deletable: true,
	}, nil
}

// List ...
func (p *ProjectAPI) List() {
	// query strings
	page, size, err := p.GetPaginationParams()
	if err != nil {
		p.SendBadRequestError(err)
		return
	}
	query := &models.ProjectQueryParam{
		Name:  p.GetString("name"),
		Owner: p.GetString("owner"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	public := p.GetString("public")
	if len(public) > 0 {
		pub, err := strconv.ParseBool(public)
		if err != nil {
			p.SendBadRequestError(fmt.Errorf("invalid public: %s", public))
			return
		}
		query.Public = &pub
	}

	var projects []*models.Project
	if !p.SecurityCtx.IsAuthenticated() {
		// not login, only get public projects
		pros, err := p.ProjectMgr.GetPublic()
		if err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
			return
		}
		projects = []*models.Project{}
		projects = append(projects, pros...)
	} else {
		if !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
			projects = []*models.Project{}
			// login, but not system admin or solution user, get public projects and
			// projects that the user is member of
			pros, err := p.ProjectMgr.GetPublic()
			if err != nil {
				p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
				return
			}
			projects = append(projects, pros...)
			if sc, ok := p.SecurityCtx.(*local.SecurityContext); ok {
				mps, err := p.ProjectMgr.GetAuthorized(sc.User())
				if err != nil {
					p.SendInternalServerError(fmt.Errorf("failed to list authorized projects: %v", err))
					return
				}
				projects = append(projects, mps...)
			}
		}
	}
	// Query projects by user group

	if projects != nil {
		projectIDs := []int64{}
		for _, project := range projects {
			projectIDs = append(projectIDs, project.ProjectID)
		}
		query.ProjectIDs = projectIDs
	}

	result, err := p.ProjectMgr.List(query)
	if err != nil {
		p.ParseAndHandleError("failed to list projects", err)
		return
	}

	for _, project := range result.Projects {
		err = p.populateProperties(project)
		if err != nil {
			log.Errorf("populate project properties failed %v", err)
		}
	}
	p.SetPaginationHeader(result.Total, page, size)
	p.Data["json"] = result.Projects
	p.ServeJSON()
}

func (p *ProjectAPI) populateProperties(project *models.Project) error {
	// Transform the severity to severity of CVSS v3.0 Ratings
	if severity, ok := project.GetMetadata(models.ProMetaSeverity); ok {
		project.SetMetadata(models.ProMetaSeverity, strings.ToLower(vuln.ParseSeverityVersion3(severity).String()))
	}

	if sc, ok := p.SecurityCtx.(*local.SecurityContext); ok {
		roles, err := pro.ListRoles(sc.User(), project.ProjectID)
		if err != nil {
			return err
		}
		project.RoleList = roles
		project.Role = highestRole(roles)
	}

	total, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{project.ProjectID},
	})
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("get repo count of project %d failed", project.ProjectID))
		return err
	}

	project.RepoCount = total

	// Populate chart count property
	if config.WithChartMuseum() {
		count, err := chartController.GetCountOfCharts([]string{project.Name})
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("get chart count of project %d failed", project.ProjectID))
			return err
		}

		project.ChartCount = count
	}
	return nil
}

// Put ...
func (p *ProjectAPI) Put() {
	if !p.requireAccess(rbac.ActionUpdate) {
		return
	}

	var req *models.ProjectRequest
	if err := p.DecodeJSONReq(&req); err != nil {
		p.SendBadRequestError(err)
		return
	}

	if err := p.ProjectMgr.Update(p.project.ProjectID,
		&models.Project{
			Metadata:     req.Metadata,
			CVEAllowlist: req.CVEAllowlist,
		}); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to update project %d",
			p.project.ProjectID), err)
		return
	}
}

// Summary returns the summary of the project
func (p *ProjectAPI) Summary() {
	if !p.requireAccess(rbac.ActionRead) {
		return
	}

	if err := p.populateProperties(p.project); err != nil {
		log.Warningf("populate project properties failed with : %+v", err)
	}

	summary := &models.ProjectSummary{
		RepoCount:  p.project.RepoCount,
		ChartCount: p.project.ChartCount,
	}

	var fetchSummaries []func(context.Context, int64, *models.ProjectSummary)

	if hasPerm, _ := p.HasProjectPermission(p.project.ProjectID, rbac.ActionRead, rbac.ResourceQuota); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectQuotaSummary)
	}

	if hasPerm, _ := p.HasProjectPermission(p.project.ProjectID, rbac.ActionList, rbac.ResourceMember); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectMemberSummary)
	}

	ctx := p.Ctx.Request.Context()

	var wg sync.WaitGroup
	for _, fn := range fetchSummaries {
		fn := fn

		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(ctx, p.project.ProjectID, summary)
		}()
	}
	wg.Wait()

	p.Data["json"] = summary
	p.ServeJSON()
}

// TODO move this to pa ckage models
func validateProjectReq(req *models.ProjectRequest) error {
	pn := req.Name
	if utils.IsIllegalLength(pn, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name %s is illegal in length. (greater than %d or less than %d)", pn, projectNameMaxLen, projectNameMinLen)
	}
	validProjectName := regexp.MustCompile(`^` + restrictedNameChars + `$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("project name is not in lower case or contains illegal characters")
	}

	metas, err := validateProjectMetadata(req.Metadata)
	if err != nil {
		return err
	}

	req.Metadata = metas
	return nil
}

func projectQuotaHardLimits(ctx context.Context, req *models.ProjectRequest, setting *models.QuotaSetting) (types.ResourceList, error) {
	hardLimits := types.ResourceList{}

	if req.StorageLimit != nil {
		hardLimits[types.ResourceStorage] = *req.StorageLimit
	} else {
		hardLimits[types.ResourceStorage] = setting.StoragePerProject
	}

	if err := quota.Validate(ctx, quota.ProjectReference, hardLimits); err != nil {
		return nil, err
	}

	return hardLimits, nil
}

func getProjectQuotaSummary(ctx context.Context, projectID int64, summary *models.ProjectSummary) {
	if !config.QuotaPerProjectEnable() {
		log.Debug("Quota per project disabled")
		return
	}

	q, err := quota.Ctl.GetByRef(ctx, quota.ProjectReference, quota.ReferenceID(projectID))
	if err != nil {
		log.Debugf("failed to get quota for project: %d", projectID)
		return
	}

	summary.Quota = &models.QuotaSummary{}
	summary.Quota.Hard, _ = types.NewResourceList(q.Hard)
	summary.Quota.Used, _ = types.NewResourceList(q.Used)
}

func getProjectMemberSummary(ctx context.Context, projectID int64, summary *models.ProjectSummary) {
	var wg sync.WaitGroup

	for _, e := range []struct {
		role  int
		count *int64
	}{
		{common.RoleProjectAdmin, &summary.ProjectAdminCount},
		{common.RoleMaintainer, &summary.MaintainerCount},
		{common.RoleDeveloper, &summary.DeveloperCount},
		{common.RoleGuest, &summary.GuestCount},
		{common.RoleLimitedGuest, &summary.LimitedGuestCount},
	} {
		wg.Add(1)
		go func(role int, count *int64) {
			defer wg.Done()

			total, err := pro.GetTotalOfProjectMembers(projectID, role)
			if err != nil {
				log.Debugf("failed to get total of project members of role %d", role)
				return
			}

			*count = total
		}(e.role, e.count)
	}

	wg.Wait()
}

// Returns the highest role in the role list.
// This func should be removed once we deprecate the "current_user_role_id" in project API
// A user can have multiple roles and they may not have a strict ranking relationship
func highestRole(roles []int) int {
	if roles == nil {
		return 0
	}
	rolePower := map[int]int{
		common.RoleProjectAdmin: 50,
		common.RoleMaintainer:   40,
		common.RoleDeveloper:    30,
		common.RoleGuest:        20,
		common.RoleLimitedGuest: 10,
	}
	var highest, highestPower int
	for _, role := range roles {
		if p, ok := rolePower[role]; ok && p > highestPower {
			highest = role
			highestPower = p
		}
	}
	return highest
}
