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

package rbac

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	rolePoliciesMap = map[string][]*types.Policy{
		"projectAdmin": {
			{Resource: ResourceSelf, Action: ActionRead},
			{Resource: ResourceSelf, Action: ActionUpdate},
			{Resource: ResourceSelf, Action: ActionDelete},

			{Resource: ResourceMember, Action: ActionCreate},
			{Resource: ResourceMember, Action: ActionRead},
			{Resource: ResourceMember, Action: ActionUpdate},
			{Resource: ResourceMember, Action: ActionDelete},
			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceMetadata, Action: ActionCreate},
			{Resource: ResourceMetadata, Action: ActionRead},
			{Resource: ResourceMetadata, Action: ActionUpdate},
			{Resource: ResourceMetadata, Action: ActionDelete},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionUpdate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceQuota, Action: ActionRead},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionRead},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionDelete},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPull},
			{Resource: ResourceRepository, Action: ActionPush},

			{Resource: ResourceTagRetention, Action: ActionCreate},
			{Resource: ResourceTagRetention, Action: ActionRead},
			{Resource: ResourceTagRetention, Action: ActionUpdate},
			{Resource: ResourceTagRetention, Action: ActionDelete},
			{Resource: ResourceTagRetention, Action: ActionList},
			{Resource: ResourceTagRetention, Action: ActionOperate},

			{Resource: ResourceImmutableTag, Action: ActionCreate},
			{Resource: ResourceImmutableTag, Action: ActionUpdate},
			{Resource: ResourceImmutableTag, Action: ActionDelete},
			{Resource: ResourceImmutableTag, Action: ActionList},

			{Resource: ResourceHelmChart, Action: ActionCreate}, // upload helm chart
			{Resource: ResourceHelmChart, Action: ActionRead},   // download helm chart
			{Resource: ResourceHelmChart, Action: ActionDelete},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate}, // upload helm chart version
			{Resource: ResourceHelmChartVersion, Action: ActionRead},   // read and download helm chart version
			{Resource: ResourceHelmChartVersion, Action: ActionDelete},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},
			{Resource: ResourceConfiguration, Action: ActionUpdate},

			{Resource: ResourceRobot, Action: ActionCreate},
			{Resource: ResourceRobot, Action: ActionRead},
			{Resource: ResourceRobot, Action: ActionUpdate},
			{Resource: ResourceRobot, Action: ActionDelete},
			{Resource: ResourceRobot, Action: ActionList},

			{Resource: ResourceNotificationPolicy, Action: ActionCreate},
			{Resource: ResourceNotificationPolicy, Action: ActionUpdate},
			{Resource: ResourceNotificationPolicy, Action: ActionDelete},
			{Resource: ResourceNotificationPolicy, Action: ActionList},
			{Resource: ResourceNotificationPolicy, Action: ActionRead},

			{Resource: ResourceScan, Action: ActionCreate},
			{Resource: ResourceScan, Action: ActionRead},

			{Resource: ResourceScanner, Action: ActionRead},
			{Resource: ResourceScanner, Action: ActionCreate},

			{Resource: ResourceArtifact, Action: ActionCreate},
			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionDelete},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifactAddition, Action: ActionRead},

			{Resource: ResourceTag, Action: ActionList},
			{Resource: ResourceTag, Action: ActionCreate},
			{Resource: ResourceTag, Action: ActionDelete},

			{Resource: ResourceArtifactLabel, Action: ActionCreate},
			{Resource: ResourceArtifactLabel, Action: ActionDelete},

			{Resource: ResourcePreatPolicy, Action: ActionCreate},
			{Resource: ResourcePreatPolicy, Action: ActionRead},
			{Resource: ResourcePreatPolicy, Action: ActionUpdate},
			{Resource: ResourcePreatPolicy, Action: ActionDelete},
			{Resource: ResourcePreatPolicy, Action: ActionList},
		},

		"maintainer": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionRead},
			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceMetadata, Action: ActionCreate},
			{Resource: ResourceMetadata, Action: ActionRead},
			{Resource: ResourceMetadata, Action: ActionUpdate},
			{Resource: ResourceMetadata, Action: ActionDelete},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceQuota, Action: ActionRead},

			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionUpdate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionRead},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionDelete},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPush},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceTagRetention, Action: ActionCreate},
			{Resource: ResourceTagRetention, Action: ActionRead},
			{Resource: ResourceTagRetention, Action: ActionUpdate},
			{Resource: ResourceTagRetention, Action: ActionDelete},
			{Resource: ResourceTagRetention, Action: ActionList},
			{Resource: ResourceTagRetention, Action: ActionOperate},

			{Resource: ResourceImmutableTag, Action: ActionCreate},
			{Resource: ResourceImmutableTag, Action: ActionUpdate},
			{Resource: ResourceImmutableTag, Action: ActionDelete},
			{Resource: ResourceImmutableTag, Action: ActionList},

			{Resource: ResourceHelmChart, Action: ActionCreate},
			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionDelete},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate},
			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionDelete},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},

			{Resource: ResourceRobot, Action: ActionRead},
			{Resource: ResourceRobot, Action: ActionList},

			{Resource: ResourceNotificationPolicy, Action: ActionList},

			{Resource: ResourceScan, Action: ActionCreate},
			{Resource: ResourceScan, Action: ActionRead},

			{Resource: ResourceScanner, Action: ActionRead},

			{Resource: ResourceArtifact, Action: ActionCreate},
			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionDelete},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifactAddition, Action: ActionRead},

			{Resource: ResourceTag, Action: ActionList},
			{Resource: ResourceTag, Action: ActionCreate},
			{Resource: ResourceTag, Action: ActionDelete},

			{Resource: ResourceArtifactLabel, Action: ActionCreate},
			{Resource: ResourceArtifactLabel, Action: ActionDelete},
		},

		"developer": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionRead},
			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceQuota, Action: ActionRead},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionRead},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPush},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceHelmChart, Action: ActionCreate},
			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate},
			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},

			{Resource: ResourceRobot, Action: ActionRead},
			{Resource: ResourceRobot, Action: ActionList},

			{Resource: ResourceScan, Action: ActionRead},

			{Resource: ResourceScanner, Action: ActionRead},

			{Resource: ResourceArtifact, Action: ActionCreate},
			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifactAddition, Action: ActionRead},

			{Resource: ResourceTag, Action: ActionList},
			{Resource: ResourceTag, Action: ActionCreate},

			{Resource: ResourceArtifactLabel, Action: ActionCreate},
			{Resource: ResourceArtifactLabel, Action: ActionDelete},
		},

		"guest": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionRead},
			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceQuota, Action: ActionRead},

			{Resource: ResourceRepository, Action: ActionRead},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceConfiguration, Action: ActionRead},

			{Resource: ResourceRobot, Action: ActionRead},
			{Resource: ResourceRobot, Action: ActionList},

			{Resource: ResourceScan, Action: ActionRead},

			{Resource: ResourceScanner, Action: ActionRead},

			{Resource: ResourceTag, Action: ActionList},

			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifactAddition, Action: ActionRead},
		},

		"limitedGuest": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceQuota, Action: ActionRead},

			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceConfiguration, Action: ActionRead},

			{Resource: ResourceScan, Action: ActionRead},

			{Resource: ResourceScanner, Action: ActionRead},

			{Resource: ResourceTag, Action: ActionList},

			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifactAddition, Action: ActionRead},
		},
	}
)

// projectRBACRole implement the RBACRole interface
type projectRBACRole struct {
	projectID int64
	roleID    int
}

// GetRoleName returns role name for the visitor role
func (role *projectRBACRole) GetRoleName() string {
	switch role.roleID {
	case common.RoleProjectAdmin:
		return "projectAdmin"
	case common.RoleMaintainer:
		return "maintainer"
	case common.RoleDeveloper:
		return "developer"
	case common.RoleGuest:
		return "guest"
	case common.RoleLimitedGuest:
		return "limitedGuest"
	default:
		return ""
	}
}

// GetPolicies returns policies for the visitor role
func (role *projectRBACRole) GetPolicies() []*types.Policy {
	policies := []*types.Policy{}

	roleName := role.GetRoleName()
	if roleName == "" {
		return policies
	}

	namespace := NewProjectNamespace(role.projectID)
	for _, policy := range rolePoliciesMap[roleName] {
		policies = append(policies, &types.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}
