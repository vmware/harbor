package v2token

import (
	"context"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"strings"

	registry_token "github.com/distribution/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// tokenSecurityCtx is used for check permission of an internal signed token.
// The intention for this guy is only for support CLI push/pull.  It should not be used in other scenario without careful review
// Each request should have a different instance of tokenSecurityCtx
type tokenSecurityCtx struct {
	logger    *log.Logger
	name      string
	accessMap map[string]map[types.Action]struct{}
	ctl       project.Controller
}

func (t *tokenSecurityCtx) Name() string {
	return "v2token"
}

func (t *tokenSecurityCtx) IsAuthenticated() bool {
	return len(t.name) > 0
}

func (t *tokenSecurityCtx) GetUsername() string {
	return t.name
}

func (t *tokenSecurityCtx) IsSysAdmin() bool {
	return false
}

func (t *tokenSecurityCtx) IsSolutionUser() bool {
	return false
}

func (t *tokenSecurityCtx) GetMyProjects() ([]*models.Project, error) {
	return []*models.Project{}, nil
}

func (t *tokenSecurityCtx) GetProjectRoles(projectIDOrName interface{}) []int {
	return []int{}
}

func (t *tokenSecurityCtx) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	if !strings.HasSuffix(resource.String(), rbac.ResourceRepository.String()) {
		return false
	}
	ns, ok := rbac_project.NamespaceParse(resource)
	if !ok {
		t.logger.Warningf("Failed to get namespace from resource: %s", resource)
		return false
	}
	pid, ok := ns.Identity().(int64)
	if !ok {
		t.logger.Warningf("Failed to get project id from namespace: %s", ns)
		return false
	}
	p, err := t.ctl.Get(ctx, pid)
	if err != nil {
		t.logger.Warningf("Failed to get project, id: %d, error: %v", pid, err)
		return false
	}
	actions, ok := t.accessMap[p.Name]
	if !ok {
		return false
	}
	_, hasAction := actions[action]
	return hasAction
}

// New creates instance of token security context based on access list and name
func New(ctx context.Context, name string, access []*registry_token.ResourceActions) security.Context {
	logger := log.G(ctx)
	m := make(map[string]map[types.Action]struct{})
	for _, ac := range access {
		if ac.Type != "repository" {
			logger.Debugf("dropped unsupported type '%s' in token", ac.Type)
			continue
		}
		l := strings.Split(ac.Name, "/")
		if len(l) < 1 {
			logger.Debugf("Unable to get project name from resource %s, drop the access", ac.Name)
			continue
		}
		actionMap := make(map[types.Action]struct{})
		for _, a := range ac.Actions {
			switch a {
			case "pull":
				actionMap[rbac.ActionPull] = struct{}{}
			case "push":
				actionMap[rbac.ActionPush] = struct{}{}
			case "delete":
				actionMap[rbac.ActionDelete] = struct{}{}
			case "scanner-pull":
				actionMap[rbac.ActionScannerPull] = struct{}{}
			case "*":
				actionMap[rbac.ActionPull] = struct{}{}
				actionMap[rbac.ActionPush] = struct{}{}
				actionMap[rbac.ActionDelete] = struct{}{}
			}
		}
		m[l[0]] = actionMap
	}

	return &tokenSecurityCtx{
		logger:    logger,
		name:      name,
		accessMap: m,
		ctl:       project.Ctl,
	}
}
