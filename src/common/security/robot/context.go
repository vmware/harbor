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

package robot

import (
	"context"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/controller/robot"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	robot         *model.Robot
	isSystemLevel bool
	ctl           project.Controller
	policies      []*types.Policy
	evaluator     evaluator.Evaluator
	once          sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(robot *model.Robot, isSystemLevel bool, policy []*types.Policy) *SecurityContext {
	return &SecurityContext{
		ctl:           project.Ctl,
		robot:         robot,
		policies:      policy,
		isSystemLevel: isSystemLevel,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "robot"
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.robot != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.robot.Name
}

// IsSysAdmin robot cannot be a system admin
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser robot cannot be a system admin
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// IsSystemLevel returns if a robot is system or project level
// true if robot is system level otherwise false
func (s *SecurityContext) IsSystemLevel() bool {
	return s.isSystemLevel
}

// Can returns whether the robot can do action on resource
func (s *SecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		if s.isSystemLevel {
			var proPolicies []*types.Policy
			var sysPolicies []*types.Policy
			var evaluators evaluator.Evaluators
			for _, p := range s.policies {
				if strings.HasPrefix(p.Resource.String(), robot.SCOPESYSTEM) {
					sysPolicies = append(sysPolicies, p)
				} else if strings.HasPrefix(p.Resource.String(), robot.SCOPEPROJECT) {
					proPolicies = append(proPolicies, p)
				}
			}
			if len(sysPolicies) != 0 {
				evaluators = evaluators.Add(system.NewEvaluator(s.GetUsername(), sysPolicies))
			} else if len(proPolicies) != 0 {
				evaluators = evaluators.Add(rbac_project.NewEvaluator(s.ctl, rbac_project.NewBuilderForPolicies(s.GetUsername(), proPolicies)))
			}
			s.evaluator = evaluators

		} else {
			s.evaluator = rbac_project.NewEvaluator(s.ctl, rbac_project.NewBuilderForPolicies(s.GetUsername(), s.policies, filterRobotPolicies))
		}
	})

	return s.evaluator != nil && s.evaluator.HasPermission(ctx, resource, action)
}

func filterRobotPolicies(p *models.Project, policies []*types.Policy) []*types.Policy {
	namespace := rbac_project.NewNamespace(p.ProjectID)

	var results []*types.Policy
	for _, policy := range policies {
		if types.ResourceAllowedInNamespace(policy.Resource, namespace) {
			results = append(results, policy)
			// give the PUSH action a pull access
			if policy.Action == rbac.ActionPush {
				results = append(results, &types.Policy{Resource: policy.Resource, Action: rbac.ActionPull})
			}
		}
	}
	return results
}
