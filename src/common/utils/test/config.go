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

package test

import (
	"github.com/goharbor/harbor/src/common"
)

var defaultConfig = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.DBAuth,
	common.DatabaseType:               "postgresql",
	common.PostGreSQLHOST:             "127.0.0.1",
	common.PostGreSQLPort:             5432,
	common.PostGreSQLUsername:         "postgres",
	common.PostGreSQLPassword:         "root123",
	common.PostGreSQLDatabase:         "registry",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1",
	common.LDAPSearchDN:               "uid=searchuser,ou=people,dc=mydomain,dc=com",
	common.LDAPSearchPwd:              "password",
	common.LDAPBaseDN:                 "ou=people,dc=mydomain,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.LDAPGroupBaseDN:            "dc=example,dc=com",
	common.LDAPGroupSearchFilter:      "objectClass=groupOfNames",
	common.LDAPGroupSearchScope:       2,
	common.LDAPGroupAttributeName:     "cn",
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailInsecure:              false,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.AdminInitialPassword:       "password",
	common.WithNotary:                 false,
	common.WithClair:                  false,
	common.UAAClientID:                "testid",
	common.UAAClientSecret:            "testsecret",
	common.UAAEndpoint:                "10.192.168.5",
	common.UAAVerifyCert:              false,
	common.CoreURL:                    "http://myui:8888/",
	common.JobServiceURL:              "http://myjob:8888/",
	common.ReadOnly:                   false,
	common.NotaryURL:                  "http://notary-server:4443",
}

// GetDefaultConfigMap returns the defailt config map for easier modification.
func GetDefaultConfigMap() map[string]interface{} {
	return defaultConfig
}
