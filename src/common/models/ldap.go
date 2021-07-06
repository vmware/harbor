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

// LdapConf holds information about ldap configuration
type LdapConf struct {
	LdapURL               string `json:"ldap_url"`
	LdapSearchDn          string `json:"ldap_search_dn"`
	LdapSearchPassword    string `json:"ldap_search_password"`
	LdapBaseDn            string `json:"ldap_base_dn"`
	LdapFilter            string `json:"ldap_filter"`
	LdapUID               string `json:"ldap_uid"`
	LdapScope             int    `json:"ldap_scope"`
	LdapPageSize          int    `json:"ldap_page_size"`
	LdapConnectionTimeout int    `json:"ldap_connection_timeout"`
	LdapVerifyCert        bool   `json:"ldap_verify_cert"`
}

// LdapGroupConf holds information about ldap group
type LdapGroupConf struct {
	LdapGroupBaseDN              string `json:"ldap_group_base_dn,omitempty"`
	LdapGroupFilter              string `json:"ldap_group_filter,omitempty"`
	LdapGroupNameAttribute       string `json:"ldap_group_name_attribute,omitempty"`
	LdapGroupSearchScope         int    `json:"ldap_group_search_scope"`
	LdapGroupAdminDN             string `json:"ldap_group_admin_dn,omitempty"`
	LdapGroupMembershipAttribute string `json:"ldap_group_membership_attribute,omitempty"`
}

// LdapUser ...
type LdapUser struct {
	Username    string   `json:"ldap_username"`
	Email       string   `json:"ldap_email"`
	Realname    string   `json:"ldap_realname"`
	DN          string   `json:"-"`
	GroupDNList []string `json:"ldap_groupdn"`
}

// LdapImportUser ...
type LdapImportUser struct {
	LdapUIDList []string `json:"ldap_uid_list"`
}

// LdapFailedImportUser ...
type LdapFailedImportUser struct {
	UID   string `json:"uid"`
	Error string `json:"err_msg"`
}

// LdapGroup ...
type LdapGroup struct {
	GroupName string `json:"group_name,omitempty"`
	GroupDN   string `json:"ldap_group_dn,omitempty"`
}
