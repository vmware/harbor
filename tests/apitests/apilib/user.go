/*
 * Harbor API
 *
 * These APIs provide services for manipulating Harbor project.
 *
 * OpenAPI spec version: 0.3.0
 *
 * Generated by: https://github.com/swagger-api/swagger-codegen.git
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package apilib

type User struct {

	// The ID of the user.
	UserId int `json:"user_id,omitempty"`

	Username string `json:"username,omitempty"`

	Email string `json:"email,omitempty"`

	Password string `json:"password,omitempty"`

	Realname string `json:"realname,omitempty"`

	Comment string `json:"comment,omitempty"`

	Deleted bool `json:"deleted,omitempty"`

	RoleName string `json:"role_name,omitempty"`

	RoleId int32 `json:"role_id,omitempty"`

	HasAdminRole bool `json:"has_admin_role,omitempty"`

	ResetUuid string `json:"reset_uuid,omitempty"`

	Salt string `json:"Salt,omitempty"`

	CreationTime string `json:"creation_time,omitempty"`

	UpdateTime string `json:"update_time,omitempty"`
}

// Permission the permission type
type Permission struct {
	Resource string `json:"resource,omitempty"`
	Action   string `json:"action,omitempty"`
}
