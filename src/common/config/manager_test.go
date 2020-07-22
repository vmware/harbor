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

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

var TestDBConfig = map[string]interface{}{
	"postgresql_host":     "localhost",
	"postgresql_database": "registry",
	"postgresql_password": "root123",
	"postgresql_username": "postgres",
	"postgresql_sslmode":  "disable",
	"email_host":          "127.0.0.1",
	"scan_all_policy":     `{"parameter":{"daily_time":0},"type":"daily"}`,
}

var configManager *CfgManager

func TestMain(m *testing.M) {
	configManager = NewDBCfgManager()
	test.InitDatabaseFromEnv()
	configManager.UpdateConfig(TestDBConfig)
	os.Exit(m.Run())
}

func TestLoadFromDatabase(t *testing.T) {
	configManager.UpdateConfig(TestDBConfig)
	configManager.Load()
	assert.Equal(t, "127.0.0.1", configManager.Get("email_host").GetString())
	assert.Equal(t, `{"parameter":{"daily_time":0},"type":"daily"}`, configManager.Get("scan_all_policy").GetString())
}

func TestLoadUserCfg(t *testing.T) {
	configMap := configManager.GetUserCfgs()
	assert.NotNil(t, configMap["ldap_url"])
	assert.NotNil(t, configMap["ldap_base_dn"])
}

func TestSaveToDatabase(t *testing.T) {
	fmt.Printf("database config %#v\n", configManager.GetDatabaseCfg())
	configManager.Load()
	configManager.Set("read_only", "true")
	configManager.Save()
	configManager.Load()
	assert.Equal(t, true, configManager.Get("read_only").GetBool())
}

func TestUpdateCfg(t *testing.T) {
	testConfig := map[string]interface{}{
		"ldap_url":             "ldaps://ldap.vmware.com",
		"ldap_search_dn":       "cn=admin,dc=example,dc=com",
		"ldap_timeout":         10,
		"ldap_search_password": "admin",
		"ldap_base_dn":         "dc=example,dc=com",
	}
	configManager.Load()
	configManager.UpdateConfig(testConfig)

	assert.Equal(t, "ldaps://ldap.vmware.com", configManager.Get("ldap_url").GetString())
	assert.Equal(t, 10, configManager.Get("ldap_timeout").GetInt())
	assert.Equal(t, "admin", configManager.Get("ldap_search_password").GetPassword())
	assert.Equal(t, "cn=admin,dc=example,dc=com", configManager.Get("ldap_search_dn").GetString())
	assert.Equal(t, "dc=example,dc=com", configManager.Get("ldap_base_dn").GetString())
}

func TestCfgManager_loadDefaultValues(t *testing.T) {
	configManager.loadDefault()
	if configManager.Get("ldap_timeout").GetInt() != 5 {
		t.Errorf("Failed to load ldap_timeout")
	}
}

func TestCfgManger_loadSystemValues(t *testing.T) {
	configManager.loadDefault()
	configManager.loadSystemConfigFromEnv()
	configManager.UpdateConfig(map[string]interface{}{
		"postgresql_host": "127.0.0.1",
	})
	if configManager.Get("postgresql_host").GetString() != "127.0.0.1" {
		t.Errorf("Failed to set system value postgresql_host, expected %v, actual %v", "127.0.0.1", configManager.Get("postgresql_host").GetString())
	}
}
func TestCfgManager_GetDatabaseCfg(t *testing.T) {
	configManager.UpdateConfig(map[string]interface{}{
		"postgresql_host":     "localhost",
		"postgresql_database": "registry",
		"postgresql_password": "root123",
		"postgresql_username": "postgres",
		"postgresql_sslmode":  "disable",
	})
	dbCfg := configManager.GetDatabaseCfg()
	assert.Equal(t, "localhost", dbCfg.PostGreSQL.Host)
	assert.Equal(t, "registry", dbCfg.PostGreSQL.Database)
	assert.Equal(t, "root123", dbCfg.PostGreSQL.Password)
	assert.Equal(t, "postgres", dbCfg.PostGreSQL.Username)
	assert.Equal(t, "disable", dbCfg.PostGreSQL.SSLMode)
}

func TestNewInMemoryManager(t *testing.T) {
	inMemoryManager := NewInMemoryManager()
	inMemoryManager.UpdateConfig(map[string]interface{}{
		"ldap_url":         "ldaps://ldap.vmware.com",
		"ldap_timeout":     5,
		"ldap_verify_cert": true,
	})
	assert.Equal(t, "ldaps://ldap.vmware.com", inMemoryManager.Get("ldap_url").GetString())
	assert.Equal(t, 5, inMemoryManager.Get("ldap_timeout").GetInt())
	assert.Equal(t, true, inMemoryManager.Get("ldap_verify_cert").GetBool())
}

/*
func TestNewRESTCfgManager(t *testing.T) {
	restMgr := NewRESTCfgManager("http://10.161.47.13:8080"+common.CoreConfigPath, "0XtgSGFx1amMDTaH")
	err := restMgr.Load()
	if err != nil {
		t.Errorf("Failed with error %v", err)
	}
	fmt.Printf("db:%v", restMgr.GetDatabaseCfg().Type)
	fmt.Printf("host:%#v\n", restMgr.GetDatabaseCfg().PostGreSQL.Host)
	fmt.Printf("port:%#v\n", restMgr.GetDatabaseCfg().PostGreSQL.Port)

}*/
