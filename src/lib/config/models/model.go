//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package models

import (
	"github.com/astaxie/beego/orm"
)

// Email ...
type Email struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
	Identity string `json:"identity"`
	From     string `json:"from"`
	Insecure bool   `json:"insecure"`
}

// HTTPAuthProxy wraps the settings for HTTP auth proxy
type HTTPAuthProxy struct {
	Endpoint            string   `json:"endpoint"`
	TokenReviewEndpoint string   `json:"tokenreivew_endpoint"`
	AdminGroups         []string `json:"admin_groups"`
	AdminUsernames      []string `json:"admin_usernames"`
	VerifyCert          bool     `json:"verify_cert"`
	SkipSearch          bool     `json:"skip_search"`
	ServerCertificate   string   `json:"server_certificate"`
}

// OIDCSetting wraps the settings for OIDC auth endpoint
type OIDCSetting struct {
	Name               string            `json:"name"`
	Endpoint           string            `json:"endpoint"`
	VerifyCert         bool              `json:"verify_cert"`
	AutoOnboard        bool              `json:"auto_onboard"`
	ClientID           string            `json:"client_id"`
	ClientSecret       string            `json:"client_secret"`
	GroupsClaim        string            `json:"groups_claim"`
	AdminGroup         string            `json:"admin_group"`
	RedirectURL        string            `json:"redirect_url"`
	Scope              []string          `json:"scope"`
	UserClaim          string            `json:"user_claim"`
	EmailClaim         string            `json:"email_claim"`
	ExtraRedirectParms map[string]string `json:"extra_redirect_parms"`
}

// QuotaSetting wraps the settings for Quota
type QuotaSetting struct {
	StoragePerProject int64 `json:"storage_per_project"`
}

func init() {
	orm.RegisterModel(new(ConfigEntry))
}

// ConfigEntry ...
type ConfigEntry struct {
	ID    int64  `orm:"pk;auto;column(id)" json:"-"`
	Key   string `orm:"column(k)" json:"k"`
	Value string `orm:"column(v)" json:"v"`
}

// TableName ...
func (ce *ConfigEntry) TableName() string {
	return "properties"
}

// Value ...
type Value struct {
	Val      interface{} `json:"value"`
	Editable bool        `json:"editable"`
}

// LdapConf holds information about ldap configuration
type LdapConf struct {
	URL               string `json:"ldap_url"`
	SearchDn          string `json:"ldap_search_dn"`
	SearchPassword    string `json:"ldap_search_password"`
	BaseDn            string `json:"ldap_base_dn"`
	Filter            string `json:"ldap_filter"`
	UID               string `json:"ldap_uid"`
	Scope             int    `json:"ldap_scope"`
	ConnectionTimeout int    `json:"ldap_connection_timeout"`
	VerifyCert        bool   `json:"ldap_verify_cert"`
}

// GroupConf holds information about ldap group
type GroupConf struct {
	BaseDN              string `json:"ldap_group_base_dn,omitempty"`
	Filter              string `json:"ldap_group_filter,omitempty"`
	NameAttribute       string `json:"ldap_group_name_attribute,omitempty"`
	SearchScope         int    `json:"ldap_group_search_scope"`
	AdminDN             string `json:"ldap_group_admin_dn,omitempty"`
	MembershipAttribute string `json:"ldap_group_membership_attribute,omitempty"`
}
