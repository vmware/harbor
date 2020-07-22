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

package route

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
)

// RegisterRoutes for Harbor legacy APIs
func registerLegacyRoutes() {
	version := APIVersion
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/members/?:pmid([0-9]+)", &api.ProjectMemberAPI{})
	beego.Router("/api/"+version+"/projects/", &api.ProjectAPI{}, "head:Head")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)", &api.ProjectAPI{})
	beego.Router("/api/"+version+"/users/:id", &api.UserAPI{}, "get:Get;delete:Delete;put:Put")
	beego.Router("/api/"+version+"/users", &api.UserAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/users/search", &api.UserAPI{}, "get:Search")
	beego.Router("/api/"+version+"/users/:id([0-9]+)/password", &api.UserAPI{}, "put:ChangePassword")
	beego.Router("/api/"+version+"/users/:id/permissions", &api.UserAPI{}, "get:ListUserPermissions")
	beego.Router("/api/"+version+"/users/:id/sysadmin", &api.UserAPI{}, "put:ToggleUserAdminRole")
	beego.Router("/api/"+version+"/users/:id/cli_secret", &api.UserAPI{}, "put:SetCLISecret")
	beego.Router("/api/"+version+"/usergroups/?:ugid([0-9]+)", &api.UserGroupAPI{})
	beego.Router("/api/"+version+"/ldap/ping", &api.LdapAPI{}, "post:Ping")
	beego.Router("/api/"+version+"/ldap/users/search", &api.LdapAPI{}, "get:Search")
	beego.Router("/api/"+version+"/ldap/groups/search", &api.LdapAPI{}, "get:SearchGroup")
	beego.Router("/api/"+version+"/ldap/users/import", &api.LdapAPI{}, "post:ImportUser")
	beego.Router("/api/"+version+"/email/ping", &api.EmailAPI{}, "post:Ping")
	beego.Router("/api/"+version+"/health", &api.HealthAPI{}, "get:CheckHealth")
	beego.Router("/api/"+version+"/ping", &api.SystemInfoAPI{}, "get:Ping")
	beego.Router("/api/"+version+"/search", &api.SearchAPI{})
	beego.Router("/api/"+version+"/projects/", &api.ProjectAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/summary", &api.ProjectAPI{}, "get:Summary")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/_deletable", &api.ProjectAPI{}, "get:Deletable")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/metadatas/?:name", &api.MetadataAPI{}, "get:Get")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/metadatas/", &api.MetadataAPI{}, "post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/robots", &api.RobotAPI{}, "post:Post;get:List")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/robots/:id([0-9]+)", &api.RobotAPI{}, "get:Get;put:Put;delete:Delete")

	beego.Router("/api/"+version+"/quotas", &api.QuotaAPI{}, "get:List")
	beego.Router("/api/"+version+"/quotas/:id([0-9]+)", &api.QuotaAPI{}, "get:Get;put:Put")

	beego.Router("/api/"+version+"/system/gc", &api.GCAPI{}, "get:List")
	beego.Router("/api/"+version+"/system/gc/:id", &api.GCAPI{}, "get:GetGC")
	beego.Router("/api/"+version+"/system/gc/:id([0-9]+)/log", &api.GCAPI{}, "get:GetLog")
	beego.Router("/api/"+version+"/system/gc/schedule", &api.GCAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/"+version+"/system/scanAll/schedule", &api.ScanAllAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/"+version+"/system/CVEAllowlist", &api.SysCVEAllowlistAPI{}, "get:Get;put:Put")
	beego.Router("/api/"+version+"/system/oidc/ping", &api.OIDCAPI{}, "post:Ping")

	beego.Router("/api/"+version+"/replication/adapters", &api.ReplicationAdapterAPI{}, "get:List")
	beego.Router("/api/"+version+"/replication/adapterinfos", &api.ReplicationAdapterAPI{}, "get:ListAdapterInfos")
	beego.Router("/api/"+version+"/replication/executions", &api.ReplicationOperationAPI{}, "get:ListExecutions;post:CreateExecution")
	beego.Router("/api/"+version+"/replication/executions/:id([0-9]+)", &api.ReplicationOperationAPI{}, "get:GetExecution;put:StopExecution")
	beego.Router("/api/"+version+"/replication/executions/:id([0-9]+)/tasks", &api.ReplicationOperationAPI{}, "get:ListTasks")
	beego.Router("/api/"+version+"/replication/executions/:id([0-9]+)/tasks/:tid([0-9]+)/log", &api.ReplicationOperationAPI{}, "get:GetTaskLog")
	beego.Router("/api/"+version+"/replication/policies", &api.ReplicationPolicyAPI{}, "get:List;post:Create")
	beego.Router("/api/"+version+"/replication/policies/:id([0-9]+)", &api.ReplicationPolicyAPI{}, "get:Get;put:Update;delete:Delete")

	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/policies", &api.NotificationPolicyAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/policies/:id([0-9]+)", &api.NotificationPolicyAPI{})
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/lasttrigger", &api.NotificationPolicyAPI{}, "get:ListGroupByEventType")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/events", &api.NotificationPolicyAPI{}, "get:GetSupportedEventTypes")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/jobs/", &api.NotificationJobAPI{}, "get:List")

	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	beego.Router("/api/"+version+"/configurations", &api.ConfigAPI{}, "get:Get;put:Put")
	beego.Router("/api/"+version+"/statistics", &api.StatisticAPI{})
	beego.Router("/api/"+version+"/labels", &api.LabelAPI{}, "post:Post;get:List")
	beego.Router("/api/"+version+"/labels/:id([0-9]+)", &api.LabelAPI{}, "get:Get;put:Put;delete:Delete")

	beego.Router("/api/"+version+"/systeminfo", &api.SystemInfoAPI{}, "get:GetGeneralInfo")
	beego.Router("/api/"+version+"/systeminfo/volumes", &api.SystemInfoAPI{}, "get:GetVolumeInfo")
	beego.Router("/api/"+version+"/systeminfo/getcert", &api.SystemInfoAPI{}, "get:GetCert")

	beego.Router("/api/"+version+"/registries", &api.RegistryAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/registries/:id([0-9]+)", &api.RegistryAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/"+version+"/registries/ping", &api.RegistryAPI{}, "post:Ping")
	// we use "0" as the ID of the local Harbor registry, so don't add "([0-9]+)" in the path
	beego.Router("/api/"+version+"/registries/:id/info", &api.RegistryAPI{}, "get:GetInfo")
	beego.Router("/api/"+version+"/registries/:id/namespace", &api.RegistryAPI{}, "get:GetNamespace")

	beego.Router("/api/"+version+"/retentions/metadatas", &api.RetentionAPI{}, "get:GetMetadatas")
	beego.Router("/api/"+version+"/retentions/:id", &api.RetentionAPI{}, "get:GetRetention")
	beego.Router("/api/"+version+"/retentions", &api.RetentionAPI{}, "post:CreateRetention")
	beego.Router("/api/"+version+"/retentions/:id", &api.RetentionAPI{}, "put:UpdateRetention")
	beego.Router("/api/"+version+"/retentions/:id/executions", &api.RetentionAPI{}, "post:TriggerRetentionExec")
	beego.Router("/api/"+version+"/retentions/:id/executions/:eid", &api.RetentionAPI{}, "patch:OperateRetentionExec")
	beego.Router("/api/"+version+"/retentions/:id/executions", &api.RetentionAPI{}, "get:ListRetentionExecs")
	beego.Router("/api/"+version+"/retentions/:id/executions/:eid/tasks", &api.RetentionAPI{}, "get:ListRetentionExecTasks")
	beego.Router("/api/"+version+"/retentions/:id/executions/:eid/tasks/:tid", &api.RetentionAPI{}, "get:GetRetentionExecTaskLog")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	// APIs for chart repository
	if config.WithChartMuseum() {
		// Labels for chart
		chartLabelAPIType := &api.ChartLabelAPI{}
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")
	}

	// Add routes for plugin scanner management
	scannerAPI := &api.ScannerAPI{}
	beego.Router("/api/"+version+"/scanners", scannerAPI, "post:Create;get:List")
	beego.Router("/api/"+version+"/scanners/:uuid", scannerAPI, "get:Get;delete:Delete;put:Update;patch:SetAsDefault")
	beego.Router("/api/"+version+"/scanners/:uuid/metadata", scannerAPI, "get:Metadata")
	beego.Router("/api/"+version+"/scanners/ping", scannerAPI, "post:Ping")

	// Add routes for project level scanner
	proScannerAPI := &api.ProjectScannerAPI{}
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/scanner", proScannerAPI, "get:GetProjectScanner;put:SetProjectScanner")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/scanner/candidates", proScannerAPI, "get:GetProScannerCandidates")

	// Add routes for scan all metrics
	scanAllAPI := &api.ScanAllAPI{}
	beego.Router("/api/"+version+"/scans/all/metrics", scanAllAPI, "get:GetScanAllMetrics")
	beego.Router("/api/"+version+"/scans/schedule/metrics", scanAllAPI, "get:GetScheduleMetrics")
}
