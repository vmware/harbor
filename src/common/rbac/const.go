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

// const action variables
const (
	ActionAll = Action("*") // action match any other actions

	ActionPull = Action("pull") // pull repository tag
	ActionPush = Action("push") // push repository tag

	// create, read, update, delete, list actions compatible with restful api methods
	ActionCreate = Action("create")
	ActionRead   = Action("read")
	ActionUpdate = Action("update")
	ActionDelete = Action("delete")
	ActionList   = Action("list")

	ActionOperate     = Action("operate")
	ActionScannerPull = Action("scanner-pull") // for robot account created by scanner to pull image, bypass the policy check
)

// const resource variables
const (
	ResourceAll                   = Resource("*")             // resource match any other resources
	ResourceConfiguration         = Resource("configuration") // project configuration compatible for portal only
	ResourceHelmChart             = Resource("helm-chart")
	ResourceHelmChartVersion      = Resource("helm-chart-version")
	ResourceHelmChartVersionLabel = Resource("helm-chart-version-label")
	ResourceLabel                 = Resource("label")
	ResourceLog                   = Resource("log")
	ResourceMember                = Resource("member")
	ResourceMetadata              = Resource("metadata")
	ResourceQuota                 = Resource("quota")
	ResourceRepository            = Resource("repository")
	ResourceTagRetention          = Resource("tag-retention")
	ResourceImmutableTag          = Resource("immutable-tag")
	ResourceRobot                 = Resource("robot")
	ResourceNotificationPolicy    = Resource("notification-policy")
	ResourceScan                  = Resource("scan")
	ResourceScanner               = Resource("scanner")
	ResourceArtifact              = Resource("artifact")
	ResourceTag                   = Resource("tag")
	ResourceArtifactAddition      = Resource("artifact-addition")
	ResourceArtifactLabel         = Resource("artifact-label")
	ResourcePreatPolicy           = Resource("preheat-policy")
	ResourceSelf                  = Resource("") // subresource for self
)
