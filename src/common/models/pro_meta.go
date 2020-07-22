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

package models

import (
	"time"
)

// keys of project metadata and severity values
const (
	ProMetaPublic               = "public"
	ProMetaEnableContentTrust   = "enable_content_trust"
	ProMetaPreventVul           = "prevent_vul" // prevent vulnerable images from being pulled
	ProMetaSeverity             = "severity"
	ProMetaAutoScan             = "auto_scan"
	ProMetaReuseSysCVEAllowlist = "reuse_sys_cve_allowlist"
)

// ProjectMetadata holds the metadata of a project.
type ProjectMetadata struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Name         string    `orm:"column(name)" json:"name"`
	Value        string    `orm:"column(value)" json:"value"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	Deleted      bool      `orm:"column(deleted)" json:"deleted"`
}
