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

package metadata

import "github.com/goharbor/harbor/src/common"

// Item - Configure item include default value, type, env name
type Item struct {
	// The Scope of this configuration item: eg: SystemScope, UserScope
	Scope string `json:"scope,omitempty"`
	// email, ldapbasic, ldapgroup, uaa settings, used to retieve configure items by group
	Group string `json:"group,omitempty"`
	// environment key to retrieves this value when initialize, for example: POSTGRESQL_HOST, only used for system settings, for user settings no EnvKey
	EnvKey string `json:"environment_key,omitempty"`
	// The default string value for this key
	DefaultValue string `json:"default_value,omitempty"`
	// The key for current configure settings in database or rest api
	Name string `json:"name,omitempty"`
	// It can be &IntType{}, &StringType{}, &BoolType{}, &PasswordType{}, &MapType{} etc, any type interface implementation
	ItemType Type
	// Editable means it can updated by configure api, For system configure, the editable is always false, for user configure, it may depends
	Editable bool `json:"editable,omitempty"`
	// Description - Describle the usage of the configure item
	Description string
}

// Constant for configure item
const (
	// Scope
	UserScope   = "user"
	SystemScope = "system"
	// Group
	LdapBasicGroup = "ldapbasic"
	LdapGroupGroup = "ldapgroup"
	EmailGroup     = "email"
	UAAGroup       = "uaa"
	HTTPAuthGroup  = "http_auth"
	OIDCGroup      = "oidc"
	DatabaseGroup  = "database"
	QuotaGroup     = "quota"
	// Put all config items do not belong a existing group into basic
	BasicGroup = "basic"
	TrivyGroup = "trivy"
)

var (
	// ConfigList - All configure items used in harbor
	// Steps to onboard a new setting
	// 1. Add configure item in metadatalist.go
	// 2. Get/Set config settings by CfgManager
	// 3. CfgManager.Load()/CfgManager.Save() to load/save from configure storage.
	ConfigList = []Item{

		{Name: common.AdminInitialPassword, Scope: SystemScope, Group: BasicGroup, EnvKey: "HARBOR_ADMIN_PASSWORD", DefaultValue: "", ItemType: &PasswordType{}, Editable: true},
		{Name: common.AUTHMode, Scope: UserScope, Group: BasicGroup, EnvKey: "AUTH_MODE", DefaultValue: "db_auth", ItemType: &AuthModeType{}, Editable: false, Description: `The auth mode of current system, such as "db_auth", "ldap_auth", "oidc_auth"`},
		{Name: common.ChartRepoURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "CHART_REPOSITORY_URL", DefaultValue: "http://chartmuseum:9999", ItemType: &StringType{}, Editable: false},

		{Name: common.TrivyAdapterURL, Scope: SystemScope, Group: TrivyGroup, EnvKey: "TRIVY_ADAPTER_URL", DefaultValue: "http://trivy-adapter:8080", ItemType: &StringType{}, Editable: false},

		{Name: common.CoreURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "CORE_URL", DefaultValue: "http://core:8080", ItemType: &StringType{}, Editable: false},
		{Name: common.CoreLocalURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "CORE_LOCAL_URL", DefaultValue: "http://127.0.0.1:8080", ItemType: &StringType{}, Editable: false},
		{Name: common.DatabaseType, Scope: SystemScope, Group: BasicGroup, EnvKey: "DATABASE_TYPE", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},

		{Name: common.EmailFrom, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_FROM", DefaultValue: "admin <sample_admin@mydomain.com>", ItemType: &StringType{}, Editable: false, Description: `The sender name for Email notification.`},
		{Name: common.EmailHost, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_HOST", DefaultValue: "smtp.mydomain.com", ItemType: &StringType{}, Editable: false, Description: `The hostname of SMTP server that sends Email notification.`},
		{Name: common.EmailIdentity, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_IDENTITY", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `By default it's empty so the email_username is picked`},
		{Name: common.EmailInsecure, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_INSECURE", DefaultValue: "false", ItemType: &BoolType{}, Editable: false, Description: `Whether or not the certificate will be verified when Harbor tries to access the email server.`},
		{Name: common.EmailPassword, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false, Description: `Email password`},
		{Name: common.EmailPort, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PORT", DefaultValue: "25", ItemType: &PortType{}, Editable: false, Description: `The port of SMTP server`},
		{Name: common.EmailSSL, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_SSL", DefaultValue: "false", ItemType: &BoolType{}, Editable: false, Description: `When it''s set to true the system will access Email server via TLS by default.  If it''s set to false, it still will handle "STARTTLS" from server side.`},
		{Name: common.EmailUsername, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_USR", DefaultValue: "sample_admin@mydomain.com", ItemType: &StringType{}, Editable: false, Description: `The username for authenticate against SMTP server`},

		{Name: common.ExtEndpoint, Scope: SystemScope, Group: BasicGroup, EnvKey: "EXT_ENDPOINT", DefaultValue: "https://host01.com", ItemType: &StringType{}, Editable: false},
		{Name: common.JobServiceURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "JOBSERVICE_URL", DefaultValue: "http://jobservice:8080", ItemType: &StringType{}, Editable: false},

		{Name: common.LDAPBaseDN, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_BASE_DN", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false, Description: `The Base DN for LDAP binding.`},
		{Name: common.LDAPFilter, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The filter for LDAP search`},
		{Name: common.LDAPGroupBaseDN, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_BASE_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The base DN to search LDAP group.`},
		{Name: common.LDAPGroupAdminDn, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_ADMIN_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `Specify the ldap group which have the same privilege with Harbor admin`},
		{Name: common.LDAPGroupAttributeName, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_GID", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The attribute which is used as identity of the LDAP group, default is cn.'`},
		{Name: common.LDAPGroupSearchFilter, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The filter to search the ldap group`},
		{Name: common.LDAPGroupSearchScope, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false, Description: `The scope to search ldap group. ''0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE''`},
		{Name: common.LDAPScope, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false, Description: `The scope to search ldap users,'0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE'`},
		{Name: common.LDAPSearchDN, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The DN of the user to do the search.`},
		{Name: common.LDAPSearchPwd, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false, Description: `The password of the ldap search dn`},
		{Name: common.LDAPTimeout, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_TIMEOUT", DefaultValue: "5", ItemType: &IntType{}, Editable: false, Description: `Timeout in seconds for connection to LDAP server`},
		{Name: common.LDAPUID, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_UID", DefaultValue: "cn", ItemType: &NonEmptyStringType{}, Editable: false, Description: `The attribute which is used as identity for the LDAP binding, such as "CN" or "SAMAccountname"`},
		{Name: common.LDAPURL, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_URL", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false, Description: `The URL of LDAP server`},
		{Name: common.LDAPVerifyCert, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_VERIFY_CERT", DefaultValue: "true", ItemType: &BoolType{}, Editable: false, Description: `Whether verify your OIDC server certificate, disable it if your OIDC server is hosted via self-hosted certificate.`},
		{Name: common.LDAPGroupMembershipAttribute, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_GROUP_MEMBERSHIP_ATTRIBUTE", DefaultValue: "memberof", ItemType: &StringType{}, Editable: true, Description: `The user attribute to identify the group membership`},

		{Name: common.MaxJobWorkers, Scope: SystemScope, Group: BasicGroup, EnvKey: "MAX_JOB_WORKERS", DefaultValue: "10", ItemType: &IntType{}, Editable: false},
		{Name: common.NotaryURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "NOTARY_URL", DefaultValue: "http://notary-server:4443", ItemType: &StringType{}, Editable: false},
		{Name: common.ScanAllPolicy, Scope: UserScope, Group: BasicGroup, EnvKey: "", DefaultValue: "", ItemType: &MapType{}, Editable: false, Description: `The policy to scan images`},

		{Name: common.PostGreSQLDatabase, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_DATABASE", DefaultValue: "registry", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLHOST, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_HOST", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLPassword, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PASSWORD", DefaultValue: "root123", ItemType: &PasswordType{}, Editable: false},
		{Name: common.PostGreSQLPort, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PORT", DefaultValue: "5432", ItemType: &PortType{}, Editable: false},
		{Name: common.PostGreSQLSSLMode, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_SSLMODE", DefaultValue: "disable", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLUsername, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_USERNAME", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLMaxIdleConns, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_MAX_IDLE_CONNS", DefaultValue: "2", ItemType: &IntType{}, Editable: false},
		{Name: common.PostGreSQLMaxOpenConns, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_MAX_OPEN_CONNS", DefaultValue: "0", ItemType: &IntType{}, Editable: false},

		{Name: common.ProjectCreationRestriction, Scope: UserScope, Group: BasicGroup, EnvKey: "PROJECT_CREATION_RESTRICTION", DefaultValue: common.ProCrtRestrEveryone, ItemType: &ProjectCreationRestrictionType{}, Editable: false, Description: `Indicate who can create projects, it could be ''adminonly'' or ''everyone''.`},
		{Name: common.ReadOnly, Scope: UserScope, Group: BasicGroup, EnvKey: "READ_ONLY", DefaultValue: "false", ItemType: &BoolType{}, Editable: false, Description: `The flag to indicate whether Harbor is in readonly mode.`},

		{Name: common.RegistryStorageProviderName, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_STORAGE_PROVIDER_NAME", DefaultValue: "filesystem", ItemType: &StringType{}, Editable: false},
		{Name: common.RegistryURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_URL", DefaultValue: "http://registry:5000", ItemType: &StringType{}, Editable: false},
		{Name: common.RegistryControllerURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_CONTROLLER_URL", DefaultValue: "http://registryctl:8080", ItemType: &StringType{}, Editable: false},
		{Name: common.SelfRegistration, Scope: UserScope, Group: BasicGroup, EnvKey: "SELF_REGISTRATION", DefaultValue: "false", ItemType: &BoolType{}, Editable: false, Description: `Whether the Harbor instance supports self-registration.  If it''s set to false, admin need to add user to the instance.`},
		{Name: common.TokenExpiration, Scope: UserScope, Group: BasicGroup, EnvKey: "TOKEN_EXPIRATION", DefaultValue: "30", ItemType: &IntType{}, Editable: false, Description: `The expiration time of the token for internal Registry, in minutes.`},
		{Name: common.TokenServiceURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "TOKEN_SERVICE_URL", DefaultValue: "", ItemType: &StringType{}, Editable: false},

		{Name: common.UAAClientID, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTID", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The client id of UAA`},
		{Name: common.UAAClientSecret, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTSECRET", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The client secret of the UAA`},
		{Name: common.UAAEndpoint, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_ENDPOINT", DefaultValue: "", ItemType: &StringType{}, Editable: false, Description: `The endpoint of the UAA`},
		{Name: common.UAAVerifyCert, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_VERIFY_CERT", DefaultValue: "false", ItemType: &BoolType{}, Editable: false, Description: `Verify the certificate in UAA server`},

		{Name: common.HTTPAuthProxyEndpoint, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}, Description: `The endpoint of the HTTP auth`},
		{Name: common.HTTPAuthProxyTokenReviewEndpoint, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}, Description: `The token review endpoint`},
		{Name: common.HTTPAuthProxyAdminGroups, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}, Description: `The group which has the harbor admin privileges`},
		{Name: common.HTTPAuthProxyAdminUsernames, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}, Description: `The username which has the harbor admin privileges`},
		{Name: common.HTTPAuthProxyVerifyCert, Scope: UserScope, Group: HTTPAuthGroup, DefaultValue: "true", ItemType: &BoolType{}, Description: `Verify the HTTP auth provider's certificate`},
		{Name: common.HTTPAuthProxySkipSearch, Scope: UserScope, Group: HTTPAuthGroup, DefaultValue: "false", ItemType: &BoolType{}, Description: `Search user before onboard`},
		{Name: common.HTTPAuthProxyServerCertificate, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}, Description: `The certificate of the HTTP auth provider`},

		{Name: common.OIDCName, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The OIDC provider name`},
		{Name: common.OIDCEndpoint, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The endpoint of the OIDC provider`},
		{Name: common.OIDCCLientID, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The client ID of the OIDC provider`},
		{Name: common.OIDCClientSecret, Scope: UserScope, Group: OIDCGroup, ItemType: &PasswordType{}, Description: `The OIDC provider secret`},
		{Name: common.OIDCGroupsClaim, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The attribute claims the group name`},
		{Name: common.OIDCAdminGroup, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The OIDC group which has the harbor admin privileges`},
		{Name: common.OIDCScope, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The scope of the OIDC provider`},
		{Name: common.OIDCUserClaim, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}, Description: `The attribute claims the username`},
		{Name: common.OIDCVerifyCert, Scope: UserScope, Group: OIDCGroup, DefaultValue: "true", ItemType: &BoolType{}, Description: `Verify the OIDC provider's certificate'`},
		{Name: common.OIDCAutoOnboard, Scope: UserScope, Group: OIDCGroup, DefaultValue: "false", ItemType: &BoolType{}, Description: `Auto onboard the OIDC user`},
		{Name: common.OIDCExtraRedirectParms, Scope: UserScope, Group: OIDCGroup, DefaultValue: "{}", ItemType: &StringToStringMapType{}, Description: `Extra parameters to add when redirect request to OIDC provider`},

		{Name: common.WithChartMuseum, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_CHARTMUSEUM", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: common.WithTrivy, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_TRIVY", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: common.WithNotary, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_NOTARY", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		// the unit of expiration is days
		{Name: common.RobotTokenDuration, Scope: UserScope, Group: BasicGroup, EnvKey: "ROBOT_TOKEN_DURATION", DefaultValue: "30", ItemType: &IntType{}, Editable: true, Description: `The robot account token duration in days`},
		{Name: common.RobotNamePrefix, Scope: UserScope, Group: BasicGroup, EnvKey: "ROBOT_NAME_PREFIX", DefaultValue: "robot$", ItemType: &StringType{}, Editable: true, Description: `The rebot account name prefix`},
		{Name: common.NotificationEnable, Scope: UserScope, Group: BasicGroup, EnvKey: "NOTIFICATION_ENABLE", DefaultValue: "true", ItemType: &BoolType{}, Editable: true, Description: `Enable notification`},
		{Name: common.MetricEnable, Scope: SystemScope, Group: BasicGroup, EnvKey: "METRIC_ENABLE", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: common.MetricPort, Scope: SystemScope, Group: BasicGroup, EnvKey: "METRIC_PORT", DefaultValue: "9090", ItemType: &IntType{}, Editable: true},
		{Name: common.MetricPath, Scope: SystemScope, Group: BasicGroup, EnvKey: "METRIC_PATH", DefaultValue: "/metrics", ItemType: &StringType{}, Editable: true},

		{Name: common.QuotaPerProjectEnable, Scope: UserScope, Group: QuotaGroup, EnvKey: "QUOTA_PER_PROJECT_ENABLE", DefaultValue: "true", ItemType: &BoolType{}, Editable: true, Description: `Enable quota per project`},
		{Name: common.StoragePerProject, Scope: UserScope, Group: QuotaGroup, EnvKey: "STORAGE_PER_PROJECT", DefaultValue: "-1", ItemType: &QuotaType{}, Editable: true, Description: `The storage quota per project`},
	}
)
