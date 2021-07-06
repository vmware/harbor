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

// Package config provide config for core api and other modules
// Before accessing user settings, need to call Load()
// For system settings, no need to call Load()
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
	comcfg "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver/local"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	defaultKeyPath                     = "/etc/core/key"
	defaultRegistryTokenPrivateKeyPath = "/etc/core/private_key.pem"

	// SessionCookieName is the name of the cookie for session ID
	SessionCookieName = "sid"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	// GlobalProjectMgr is initialized based on the deploy mode
	GlobalProjectMgr promgr.ProjectManager
	keyProvider      comcfg.KeyProvider
	// defined as a var for testing.
	defaultCACertPath = "/etc/core/ca/ca.crt"
	cfgMgr            *comcfg.CfgManager
)

// Init configurations
func Init() {
	// init key provider
	initKeyProvider()

	cfgMgr = comcfg.NewDBCfgManager()

	log.Info("init secret store")
	// init secret store
	initSecretStore()
	log.Info("init project manager")
	// init project manager
	initProjectManager()
}

// InitWithSettings init config with predefined configs, and optionally overwrite the keyprovider
func InitWithSettings(cfgs map[string]interface{}, kp ...comcfg.KeyProvider) {
	Init()
	cfgMgr = comcfg.NewInMemoryManager()
	cfgMgr.UpdateConfig(cfgs)
	if len(kp) > 0 {
		keyProvider = kp[0]
	}
}

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	log.Infof("key path: %s", path)

	keyProvider = comcfg.NewFileKeyProvider(path)
}

func initSecretStore() {
	m := map[string]string{}
	m[JobserviceSecret()] = secret.JobserviceUser
	SecretStore = secret.NewStore(m)
}

func initProjectManager() {
	log.Info("initializing the project manager based on local database...")
	GlobalProjectMgr = promgr.NewDefaultProjectManager(local.NewDriver(), true)
}

// GetCfgManager return the current config manager
func GetCfgManager() *comcfg.CfgManager {
	if cfgMgr == nil {
		return comcfg.NewDBCfgManager()
	}
	return cfgMgr
}

// Load configurations
func Load() error {
	return cfgMgr.Load()
}

// Upload save all system configurations
func Upload(cfg map[string]interface{}) error {
	return cfgMgr.UpdateConfig(cfg)
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (map[string]interface{}, error) {
	sysCfg := cfgMgr.GetAll()
	if len(sysCfg) == 0 {
		return nil, errors.New("can not load system config, the database might be down")
	}
	return sysCfg, nil
}

// AuthMode ...
func AuthMode() (string, error) {
	err := cfgMgr.Load()
	if err != nil {
		log.Errorf("failed to load config, error %v", err)
		return "db_auth", err
	}
	return cfgMgr.Get(common.AUTHMode).GetString(), nil
}

// TokenPrivateKeyPath returns the path to the key for signing token for registry
func TokenPrivateKeyPath() string {
	path := os.Getenv("TOKEN_PRIVATE_KEY_PATH")
	if len(path) == 0 {
		path = defaultRegistryTokenPrivateKeyPath
	}
	return path
}

// LDAPConf returns the setting of ldap server
func LDAPConf() (*models.LdapConf, error) {
	err := cfgMgr.Load()
	if err != nil {
		return nil, err
	}
	return &models.LdapConf{
		LdapURL:               cfgMgr.Get(common.LDAPURL).GetString(),
		LdapSearchDn:          cfgMgr.Get(common.LDAPSearchDN).GetString(),
		LdapSearchPassword:    cfgMgr.Get(common.LDAPSearchPwd).GetString(),
		LdapBaseDn:            cfgMgr.Get(common.LDAPBaseDN).GetString(),
		LdapUID:               cfgMgr.Get(common.LDAPUID).GetString(),
		LdapFilter:            cfgMgr.Get(common.LDAPFilter).GetString(),
		LdapScope:             cfgMgr.Get(common.LDAPScope).GetInt(),
		LdapPageSize:          cfgMgr.Get(common.LDAPPageSize).GetInt(),
		LdapConnectionTimeout: cfgMgr.Get(common.LDAPTimeout).GetInt(),
		LdapVerifyCert:        cfgMgr.Get(common.LDAPVerifyCert).GetBool(),
	}, nil
}

// LDAPGroupConf returns the setting of ldap group search
func LDAPGroupConf() (*models.LdapGroupConf, error) {
	err := cfgMgr.Load()
	if err != nil {
		return nil, err
	}
	return &models.LdapGroupConf{
		LdapGroupBaseDN:              cfgMgr.Get(common.LDAPGroupBaseDN).GetString(),
		LdapGroupFilter:              cfgMgr.Get(common.LDAPGroupSearchFilter).GetString(),
		LdapGroupNameAttribute:       cfgMgr.Get(common.LDAPGroupAttributeName).GetString(),
		LdapGroupSearchScope:         cfgMgr.Get(common.LDAPGroupSearchScope).GetInt(),
		LdapGroupAdminDN:             cfgMgr.Get(common.LDAPGroupAdminDn).GetString(),
		LdapGroupMembershipAttribute: cfgMgr.Get(common.LDAPGroupMembershipAttribute).GetString(),
	}, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() (int, error) {
	return cfgMgr.Get(common.TokenExpiration).GetInt(), nil
}

// RobotTokenDuration returns the token expiration time of robot account (in minute)
func RobotTokenDuration() int {
	return cfgMgr.Get(common.RobotTokenDuration).GetInt()
}

// ExtEndpoint returns the external URL of Harbor: protocol://host:port
func ExtEndpoint() (string, error) {
	return cfgMgr.Get(common.ExtEndpoint).GetString(), nil
}

// ExtURL returns the external URL: host:port
func ExtURL() (string, error) {
	endpoint, err := ExtEndpoint()
	if err != nil {
		log.Errorf("failed to load config, error %v", err)
	}
	l := strings.Split(endpoint, "://")
	if len(l) > 1 {
		return l[1], nil
	}
	return endpoint, nil
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() (string, error) {
	return keyProvider.Get(nil)
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration() (bool, error) {
	return cfgMgr.Get(common.SelfRegistration).GetBool(), nil
}

// RegistryURL ...
func RegistryURL() (string, error) {
	url := os.Getenv("REGISTRY_URL")
	if len(url) == 0 {
		url = "http://registry:5000"
	}
	return url, nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return os.Getenv("JOBSERVICE_URL")
}

// GetCoreURL returns the url of core from env
func GetCoreURL() string {
	return os.Getenv("CORE_URL")
}

// InternalCoreURL returns the local harbor core url
func InternalCoreURL() string {
	return strings.TrimSuffix(cfgMgr.Get(common.CoreURL).GetString(), "/")
}

// LocalCoreURL returns the local harbor core url
func LocalCoreURL() string {
	return cfgMgr.Get(common.CoreLocalURL).GetString()
}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	return InternalCoreURL() + "/service/token"
}

// InternalNotaryEndpoint returns notary server endpoint for internal communication between Harbor containers
// This is currently a conventional value and can be unaccessible when Harbor is not deployed with Notary.
func InternalNotaryEndpoint() string {
	return cfgMgr.Get(common.NotaryURL).GetString()
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	return cfgMgr.Get(common.AdminInitialPassword).GetString(), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() (bool, error) {
	return cfgMgr.Get(common.ProjectCreationRestriction).GetString() == common.ProCrtRestrAdmOnly, nil
}

// Email returns email server settings
func Email() (*models.Email, error) {
	err := cfgMgr.Load()
	if err != nil {
		return nil, err
	}
	return &models.Email{
		Host:     cfgMgr.Get(common.EmailHost).GetString(),
		Port:     cfgMgr.Get(common.EmailPort).GetInt(),
		Username: cfgMgr.Get(common.EmailUsername).GetString(),
		Password: cfgMgr.Get(common.EmailPassword).GetString(),
		SSL:      cfgMgr.Get(common.EmailSSL).GetBool(),
		From:     cfgMgr.Get(common.EmailFrom).GetString(),
		Identity: cfgMgr.Get(common.EmailIdentity).GetString(),
		Insecure: cfgMgr.Get(common.EmailInsecure).GetBool(),
	}, nil
}

// Database returns database settings
func Database() (*models.Database, error) {
	database := &models.Database{}
	database.Type = cfgMgr.Get(common.DatabaseType).GetString()
	postgresql := &models.PostGreSQL{
		Host:         cfgMgr.Get(common.PostGreSQLHOST).GetString(),
		Port:         cfgMgr.Get(common.PostGreSQLPort).GetInt(),
		Username:     cfgMgr.Get(common.PostGreSQLUsername).GetString(),
		Password:     cfgMgr.Get(common.PostGreSQLPassword).GetString(),
		Database:     cfgMgr.Get(common.PostGreSQLDatabase).GetString(),
		SSLMode:      cfgMgr.Get(common.PostGreSQLSSLMode).GetString(),
		MaxIdleConns: cfgMgr.Get(common.PostGreSQLMaxIdleConns).GetInt(),
		MaxOpenConns: cfgMgr.Get(common.PostGreSQLMaxOpenConns).GetInt(),
	}
	database.PostGreSQL = postgresql

	return database, nil
}

// CoreSecret returns a secret to mark harbor-core when communicate with
// other component
func CoreSecret() string {
	return os.Getenv("CORE_SECRET")
}

// RegistryCredential returns the username and password the core uses to access registry
func RegistryCredential() (string, string) {
	return os.Getenv("REGISTRY_CREDENTIAL_USERNAME"), os.Getenv("REGISTRY_CREDENTIAL_PASSWORD")
}

// JobserviceSecret returns a secret to mark Jobservice when communicate with
// other component
// TODO replace it with method of SecretStore
func JobserviceSecret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}

// WithNotary returns a bool value to indicate if Harbor's deployed with Notary
func WithNotary() bool {
	return cfgMgr.Get(common.WithNotary).GetBool()
}

// WithClair returns a bool value to indicate if Harbor's deployed with Clair
func WithClair() bool {
	return cfgMgr.Get(common.WithClair).GetBool()
}

// ClairAdapterEndpoint returns the endpoint of clair adapter instance, by default it's the one deployed within Harbor.
func ClairAdapterEndpoint() string {
	return cfgMgr.Get(common.ClairAdapterURL).GetString()
}

// WithTrivy returns a bool value to indicate if Harbor's deployed with Trivy.
func WithTrivy() bool {
	return cfgMgr.Get(common.WithTrivy).GetBool()
}

// TrivyAdapterURL returns the endpoint URL of a Trivy adapter instance, by default it's the one deployed within Harbor.
func TrivyAdapterURL() string {
	return cfgMgr.Get(common.TrivyAdapterURL).GetString()
}

// UAASettings returns the UAASettings to access UAA service.
func UAASettings() (*models.UAASettings, error) {
	err := cfgMgr.Load()
	if err != nil {
		return nil, err
	}
	us := &models.UAASettings{
		Endpoint:     cfgMgr.Get(common.UAAEndpoint).GetString(),
		ClientID:     cfgMgr.Get(common.UAAClientID).GetString(),
		ClientSecret: cfgMgr.Get(common.UAAClientSecret).GetString(),
		VerifyCert:   cfgMgr.Get(common.UAAVerifyCert).GetBool(),
	}
	return us, nil
}

// ReadOnly returns a bool to indicates if Harbor is in read only mode.
func ReadOnly() bool {
	return cfgMgr.Get(common.ReadOnly).GetBool()
}

// WithChartMuseum returns a bool to indicate if chartmuseum is deployed with Harbor.
func WithChartMuseum() bool {
	return cfgMgr.Get(common.WithChartMuseum).GetBool()
}

// GetChartMuseumEndpoint returns the endpoint of the chartmuseum service
// otherwise an non nil error is returned
func GetChartMuseumEndpoint() (string, error) {
	chartEndpoint := strings.TrimSpace(cfgMgr.Get(common.ChartRepoURL).GetString())
	if len(chartEndpoint) == 0 {
		return "", errors.New("empty chartmuseum endpoint")
	}
	return chartEndpoint, nil
}

// GetRedisOfRegURL returns the URL of Redis used by registry
func GetRedisOfRegURL() string {
	return os.Getenv("_REDIS_URL_REG")
}

// GetPortalURL returns the URL of portal
func GetPortalURL() string {
	url := os.Getenv("PORTAL_URL")
	if len(url) == 0 {
		return common.DefaultPortalURL
	}
	return url
}

// GetRegistryCtlURL returns the URL of registryctl
func GetRegistryCtlURL() string {
	url := os.Getenv("REGISTRY_CONTROLLER_URL")
	if len(url) == 0 {
		return common.DefaultRegistryCtlURL
	}
	return url
}

// HTTPAuthProxySetting returns the setting of HTTP Auth proxy.  the settings are only meaningful when the auth_mode is
// set to http_auth
func HTTPAuthProxySetting() (*models.HTTPAuthProxy, error) {
	if err := cfgMgr.Load(); err != nil {
		return nil, err
	}
	return &models.HTTPAuthProxy{
		Endpoint:            cfgMgr.Get(common.HTTPAuthProxyEndpoint).GetString(),
		TokenReviewEndpoint: cfgMgr.Get(common.HTTPAuthProxyTokenReviewEndpoint).GetString(),
		VerifyCert:          cfgMgr.Get(common.HTTPAuthProxyVerifyCert).GetBool(),
		SkipSearch:          cfgMgr.Get(common.HTTPAuthProxySkipSearch).GetBool(),
		ServerCertificate:   cfgMgr.Get(common.HTTPAuthProxyServerCertificate).GetString(),
	}, nil
}

// OIDCSetting returns the setting of OIDC provider, currently there's only one OIDC provider allowed for Harbor and it's
// only effective when auth_mode is set to oidc_auth
func OIDCSetting() (*models.OIDCSetting, error) {
	if err := cfgMgr.Load(); err != nil {
		return nil, err
	}
	scopeStr := cfgMgr.Get(common.OIDCScope).GetString()
	extEndpoint := strings.TrimSuffix(cfgMgr.Get(common.ExtEndpoint).GetString(), "/")
	scope := []string{}
	for _, s := range strings.Split(scopeStr, ",") {
		scope = append(scope, strings.TrimSpace(s))
	}

	return &models.OIDCSetting{
		Name:         cfgMgr.Get(common.OIDCName).GetString(),
		Endpoint:     cfgMgr.Get(common.OIDCEndpoint).GetString(),
		VerifyCert:   cfgMgr.Get(common.OIDCVerifyCert).GetBool(),
		AutoOnboard:  cfgMgr.Get(common.OIDCAutoOnboard).GetBool(),
		ClientID:     cfgMgr.Get(common.OIDCCLientID).GetString(),
		ClientSecret: cfgMgr.Get(common.OIDCClientSecret).GetString(),
		GroupsClaim:  cfgMgr.Get(common.OIDCGroupsClaim).GetString(),
		RedirectURL:  extEndpoint + common.OIDCCallbackPath,
		Scope:        scope,
		UserClaim:    cfgMgr.Get(common.OIDCUserClaim).GetString(),
	}, nil
}

// NotificationEnable returns a bool to indicates if notification enabled in harbor
func NotificationEnable() bool {
	return cfgMgr.Get(common.NotificationEnable).GetBool()
}

// QuotaPerProjectEnable returns a bool to indicates if quota per project enabled in harbor
func QuotaPerProjectEnable() bool {
	return cfgMgr.Get(common.QuotaPerProjectEnable).GetBool()
}

// QuotaSetting returns the setting of quota.
func QuotaSetting() (*models.QuotaSetting, error) {
	if err := cfgMgr.Load(); err != nil {
		return nil, err
	}
	return &models.QuotaSetting{
		StoragePerProject: cfgMgr.Get(common.StoragePerProject).GetInt64(),
	}, nil
}

// GetPermittedRegistryTypesForProxyCache returns the permitted registry types for proxy cache
func GetPermittedRegistryTypesForProxyCache() []string {
	types := os.Getenv("PERMITTED_REGISTRY_TYPES_FOR_PROXY_CACHE")
	if len(types) == 0 {
		return []string{}
	}
	return strings.Split(types, ",")
}

// GetGCTimeWindow returns the reserve time window of blob.
func GetGCTimeWindow() int64 {
	// the env is for testing/debugging. For production, Do NOT set it.
	if env, exist := os.LookupEnv("GC_TIME_WINDOW_HOURS"); exist {
		timeWindow, err := strconv.ParseInt(env, 10, 64)
		if err == nil {
			return timeWindow
		}
	}
	return common.DefaultGCTimeWindowHours
}
