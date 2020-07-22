package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/common/rbac"
	preheatCtl "github.com/goharbor/harbor/src/controller/p2p/preheat"
	projectCtl "github.com/goharbor/harbor/src/controller/project"
	taskCtl "github.com/goharbor/harbor/src/controller/task"
	"github.com/goharbor/harbor/src/jobservice/job"
	liberrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	instanceModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/preheat"
)

func newPreheatAPI() *preheatAPI {
	return &preheatAPI{
		preheatCtl:   preheatCtl.Ctl,
		projectCtl:   projectCtl.Ctl,
		enforcer:     preheatCtl.Enf,
		executionCtl: taskCtl.ExecutionCtl,
		taskCtl:      taskCtl.Ctl,
	}
}

var _ restapi.PreheatAPI = (*preheatAPI)(nil)

// nameRegex is the regex for name validation.
const nameRegex = "^[A-Za-z0-9]+(?:[._-][A-Za-z0-9]+)*$"

type preheatAPI struct {
	BaseAPI
	preheatCtl   preheatCtl.Controller
	projectCtl   projectCtl.Controller
	enforcer     preheatCtl.Enforcer
	executionCtl taskCtl.ExecutionController
	taskCtl      taskCtl.Controller
}

func (api *preheatAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (api *preheatAPI) CreateInstance(ctx context.Context, params operation.CreateInstanceParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	instance, err := convertParamInstanceToModelInstance(params.Instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	_, err = api.preheatCtl.CreateInstance(ctx, instance)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewCreateInstanceCreated()
}

func (api *preheatAPI) DeleteInstance(ctx context.Context, params operation.DeleteInstanceParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	instance, err := api.preheatCtl.GetInstanceByName(ctx, params.PreheatInstanceName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.DeleteInstance(ctx, instance.ID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewDeleteInstanceOK()
}

func (api *preheatAPI) GetInstance(ctx context.Context, params operation.GetInstanceParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	var payload *models.Instance
	instance, err := api.preheatCtl.GetInstanceByName(ctx, params.PreheatInstanceName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	payload, err = convertInstanceToPayload(instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewGetInstanceOK().WithPayload(payload)
}

// ListInstances is List p2p instances
func (api *preheatAPI) ListInstances(ctx context.Context, params operation.ListInstancesParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	var payload []*models.Instance

	query, err := api.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}

	total, err := api.preheatCtl.CountInstance(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	instances, err := api.preheatCtl.ListInstance(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	for _, instance := range instances {
		ins, err := convertInstanceToPayload(instance)
		if err != nil {
			return api.SendError(ctx, err)
		}
		payload = append(payload, ins)
	}
	return operation.NewListInstancesOK().
		WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (api *preheatAPI) ListProviders(ctx context.Context, params operation.ListProvidersParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	var providers, err = preheatCtl.Ctl.GetAvailableProviders()
	if err != nil {
		return operation.NewListProvidersInternalServerError()
	}
	var payload = convertProvidersToFrontend(providers)

	return operation.NewListProvidersOK().WithPayload(payload)
}

// UpdateInstance is Update instance
func (api *preheatAPI) UpdateInstance(ctx context.Context, params operation.UpdateInstanceParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	instance, err := convertParamInstanceToModelInstance(params.Instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.UpdateInstance(ctx, instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewUpdateInstanceOK()
}

func convertProvidersToFrontend(backend []*provider.Metadata) (frontend []*models.Metadata) {
	frontend = make([]*models.Metadata, 0)
	for _, provider := range backend {
		frontend = append(frontend, &models.Metadata{
			ID:          provider.ID,
			Icon:        provider.Icon,
			Name:        provider.Name,
			Source:      provider.Source,
			Version:     provider.Version,
			Maintainers: provider.Maintainers,
		})
	}
	return
}

// GetPolicy is Get a preheat policy
func (api *preheatAPI) GetPolicy(ctx context.Context, params operation.GetPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payload *models.PreheatPolicy
	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	// get provider
	provider, err := api.preheatCtl.GetInstance(ctx, policy.ProviderID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	payload, err = convertPolicyToPayload(policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	payload.ProviderName = provider.Name

	return operation.NewGetPolicyOK().WithPayload(payload)
}

// CreatePolicy is Create a preheat policy under a project
func (api *preheatAPI) CreatePolicy(ctx context.Context, params operation.CreatePolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := convertParamPolicyToModelPolicy(params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}

	_, err = api.preheatCtl.CreatePolicy(ctx, policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewCreatePolicyCreated()
}

// UpdatePolicy is Update preheat policy
func (api *preheatAPI) UpdatePolicy(ctx context.Context, params operation.UpdatePolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionUpdate, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := convertParamPolicyToModelPolicy(params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.UpdatePolicy(ctx, policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewUpdatePolicyOK()
}

// DeletePolicy is Delete a preheat policy
func (api *preheatAPI) DeletePolicy(ctx context.Context, params operation.DeletePolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.DeletePolicy(ctx, policy.ID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewDeletePolicyOK()
}

// ListPolicies is List preheat policies
func (api *preheatAPI) ListPolicies(ctx context.Context, params operation.ListPoliciesParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	query, err := api.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}

	if query != nil {
		query.Keywords["project_id"] = project.ProjectID
	}

	total, err := api.preheatCtl.CountPolicy(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policies, err := api.preheatCtl.ListPolicies(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payload []*models.PreheatPolicy
	for _, policy := range policies {
		// get provider
		provider, err := api.preheatCtl.GetInstance(ctx, policy.ProviderID)
		if err != nil {
			return api.SendError(ctx, err)
		}

		p, err := convertPolicyToPayload(policy)
		if err != nil {
			return api.SendError(ctx, err)
		}
		p.ProviderName = provider.Name
		payload = append(payload, p)
	}
	return operation.NewListPoliciesOK().WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

// ManualPreheat is manual preheat
func (api *preheatAPI) ManualPreheat(ctx context.Context, params operation.ManualPreheatParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	_, err = api.enforcer.EnforcePolicy(ctx, policy.ID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	// TODO: build execution URL
	var location = ""

	return operation.NewManualPreheatCreated().WithLocation(location)
}

func (api *preheatAPI) PingInstances(ctx context.Context, params operation.PingInstancesParams) middleware.Responder {
	if err := api.RequireSysAdmin(ctx); err != nil {
		return api.SendError(ctx, err)
	}

	var instance *instanceModel.Instance
	var err error

	if params.Instance.ID > 0 {
		// by ID
		instance, err = api.preheatCtl.GetInstance(ctx, params.Instance.ID)
		if liberrors.IsNotFoundErr(err) {
			return operation.NewPingInstancesNotFound()
		}
		if err != nil {
			return api.SendError(ctx, err)
		}
	} else {
		// by endpoint URL
		if params.Instance.Endpoint == "" {
			return operation.NewPingInstancesBadRequest()
		}

		instance, err = convertParamInstanceToModelInstance(params.Instance)
		if err != nil {
			return api.SendError(ctx, err)
		}
	}

	err = api.preheatCtl.CheckHealth(ctx, instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewPingInstancesOK()
}

// convertPolicyToPayload converts model policy to swagger model
func convertPolicyToPayload(policy *policy.Schema) (*models.PreheatPolicy, error) {
	if policy == nil {
		return nil, errors.New("policy can not be nil")
	}

	return &models.PreheatPolicy{
		CreationTime: strfmt.DateTime(policy.CreatedAt),
		Description:  policy.Description,
		Enabled:      policy.Enabled,
		Filters:      policy.FiltersStr,
		ID:           policy.ID,
		Name:         policy.Name,
		ProjectID:    policy.ProjectID,
		ProviderID:   policy.ProviderID,
		Trigger:      policy.TriggerStr,
		UpdateTime:   strfmt.DateTime(policy.UpdatedTime),
	}, nil
}

// convertParamPolicyToPolicy converts params policy to pkg model policy
func convertParamPolicyToModelPolicy(model *models.PreheatPolicy) (*policy.Schema, error) {
	if model == nil {
		return nil, errors.New("policy can not be nil")
	}

	valid, err := regexp.MatchString(nameRegex, model.Name)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("name %s is invalid", model.Name)
	}

	return &policy.Schema{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		ProjectID:   model.ProjectID,
		ProviderID:  model.ProviderID,
		FiltersStr:  model.Filters,
		TriggerStr:  model.Trigger,
		Enabled:     model.Enabled,
		CreatedAt:   time.Time(model.CreationTime),
		UpdatedTime: time.Time(model.UpdateTime),
	}, nil
}

func convertInstanceToPayload(model *instanceModel.Instance) (*models.Instance, error) {
	if model == nil {
		return nil, errors.New("instance can not be nil")
	}

	var authInfo = map[string]string{}
	var err = json.Unmarshal([]byte(model.AuthData), &authInfo)
	if err != nil {
		return nil, err
	}
	return &models.Instance{
		AuthInfo:       authInfo,
		AuthMode:       model.AuthMode,
		Default:        model.Default,
		Description:    model.Description,
		Enabled:        model.Enabled,
		Endpoint:       model.Endpoint,
		ID:             model.ID,
		Insecure:       model.Insecure,
		Name:           model.Name,
		SetupTimestamp: model.SetupTimestamp,
		Status:         "Unknown",
		Vendor:         model.Vendor,
	}, nil
}

func convertParamInstanceToModelInstance(model *models.Instance) (*instanceModel.Instance, error) {
	if model == nil {
		return nil, errors.New("instance can not be nil")
	}

	valid, err := regexp.MatchString(nameRegex, model.Name)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("name %s is invalid", model.Name)
	}

	authData, err := json.Marshal(model.AuthInfo)
	if err != nil {
		return nil, err
	}

	return &instanceModel.Instance{
		AuthData:       string(authData),
		AuthMode:       model.AuthMode,
		Default:        model.Default,
		Description:    model.Description,
		Enabled:        model.Enabled,
		Endpoint:       model.Endpoint,
		ID:             model.ID,
		Insecure:       model.Insecure,
		Name:           model.Name,
		SetupTimestamp: model.SetupTimestamp,
		Status:         model.Status,
		Vendor:         model.Vendor,
	}, nil
}

// convertExecutionToPayload converts model execution to swagger model.
func convertExecutionToPayload(model *task.Execution) (*models.Execution, error) {
	if model == nil {
		return nil, errors.New("execution can not be nil")
	}

	execution := &models.Execution{
		EndTime:       model.EndTime.String(),
		ExtraAttrs:    model.ExtraAttrs,
		ID:            model.ID,
		StartTime:     model.StartTime.String(),
		Status:        model.Status,
		StatusMessage: model.StatusMessage,
		Trigger:       model.Trigger,
		VendorID:      model.VendorID,
		VendorType:    model.VendorType,
	}
	if model.Metrics != nil {
		execution.Metrics = &models.Metrics{
			ErrorTaskCount:     model.Metrics.ErrorTaskCount,
			PendingTaskCount:   model.Metrics.PendingTaskCount,
			RunningTaskCount:   model.Metrics.RunningTaskCount,
			ScheduledTaskCount: model.Metrics.ScheduledTaskCount,
			StoppedTaskCount:   model.Metrics.StoppedTaskCount,
			SuccessTaskCount:   model.Metrics.SuccessTaskCount,
			TaskCount:          model.Metrics.TaskCount,
		}
	}

	return execution, nil
}

// GetExecution gets an execution.
func (api *preheatAPI) GetExecution(ctx context.Context, params operation.GetExecutionParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	execution, err := api.executionCtl.Get(ctx, params.ExecutionID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	payload, err := convertExecutionToPayload(execution)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewGetExecutionOK().WithPayload(payload)
}

// ListExecutions lists executions.
func (api *preheatAPI) ListExecutions(ctx context.Context, params operation.ListExecutionsParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	query, err := api.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}

	if query != nil {
		query.Keywords["vendor_type"] = job.P2PPreheat
		query.Keywords["vendor_id"] = policy.ID
	}

	executions, err := api.executionCtl.List(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payloads []*models.Execution
	for _, exec := range executions {
		p, err := convertExecutionToPayload(exec)
		if err != nil {
			return api.SendError(ctx, err)
		}
		payloads = append(payloads, p)
	}

	return operation.NewListExecutionsOK().WithPayload(payloads)
}

// StopExecution stops execution.
func (api *preheatAPI) StopExecution(ctx context.Context, params operation.StopExecutionParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionUpdate, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	if params.Execution.Status == "Stopped" {
		err := api.executionCtl.Stop(ctx, params.ExecutionID)
		if err != nil {
			return api.SendError(ctx, err)
		}

		return operation.NewStopExecutionOK()
	}

	return api.SendError(ctx, fmt.Errorf("param status invalid: %#v", params.Execution))
}

// convertTaskToPayload converts task to swagger model.
func convertTaskToPayload(model *task.Task) (*models.Task, error) {
	if model == nil {
		return nil, errors.New("task model can not be nil")
	}

	return &models.Task{
		CreationTime:  model.CreationTime.String(),
		EndTime:       model.EndTime.String(),
		ExecutionID:   model.ExecutionID,
		ExtraAttrs:    model.ExtraAttrs,
		ID:            model.ID,
		RunCount:      int64(model.RunCount),
		StartTime:     model.StartTime.String(),
		Status:        model.Status,
		StatusMessage: model.StatusMessage,
		UpdateTime:    model.UpdateTime.String(),
	}, nil
}

// ListTasks lists tasks.
func (api *preheatAPI) ListTasks(ctx context.Context, params operation.ListTasksParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	query := &q.Query{
		Keywords: map[string]interface{}{
			"execution_id": params.ExecutionID,
		},
	}

	tasks, err := api.taskCtl.List(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payloads []*models.Task
	for _, task := range tasks {
		p, err := convertTaskToPayload(task)
		if err != nil {
			return api.SendError(ctx, err)
		}
		payloads = append(payloads, p)
	}

	return operation.NewListTasksOK().WithPayload(payloads)
}

// GetLog gets log.
func (api *preheatAPI) GetLog(ctx context.Context, params operation.GetLogParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	l, err := api.taskCtl.GetLog(ctx, params.TaskID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewGetLogOK().WithPayload(string(l))
}

// ListProvidersUnderProject is Get all providers at project level
func (api *preheatAPI) ListProvidersUnderProject(ctx context.Context, params operation.ListProvidersUnderProjectParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourcePreatPolicy); err != nil {
		return api.SendError(ctx, err)
	}

	instances, err := api.preheatCtl.ListInstance(ctx, &q.Query{})
	if err != nil {
		return api.SendError(ctx, err)
	}

	var providers []*models.ProviderUnderProject
	for _, instance := range instances {
		providers = append(providers, &models.ProviderUnderProject{
			ID:       instance.ID,
			Provider: fmt.Sprintf("%s %s-%s", instance.Vendor, instance.Name, instance.Endpoint),
		})
	}

	return operation.NewListProvidersUnderProjectOK().WithPayload(providers)
}
