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

package config

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"strings"
)

// It contains all user related configurations, each of user related settings requires a context provided

// GetSystemCfg returns the all configurations
func GetSystemCfg(ctx context.Context) (map[string]interface{}, error) {
	sysCfg := defaultMgr().GetAll(ctx)
	if len(sysCfg) == 0 {
		return nil, errors.New("can not load system config, the database might be down")
	}
	return sysCfg, nil
}

// AuthMode ...
func AuthMode(ctx context.Context) (string, error) {
	mgr := defaultMgr()
	err := mgr.Load(ctx)
	if err != nil {
		log.Errorf("failed to load config, error %v", err)
		return "db_auth", err
	}
	return mgr.Get(ctx, common.AUTHMode).GetString(), nil
}

// LDAPConf returns the setting of ldap server
func LDAPConf(ctx context.Context) (*cfgModels.LdapConf, error) {
	mgr := defaultMgr()
	err := mgr.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &cfgModels.LdapConf{
		URL:               mgr.Get(ctx, common.LDAPURL).GetString(),
		SearchDn:          mgr.Get(ctx, common.LDAPSearchDN).GetString(),
		SearchPassword:    mgr.Get(ctx, common.LDAPSearchPwd).GetString(),
		BaseDn:            mgr.Get(ctx, common.LDAPBaseDN).GetString(),
		UID:               mgr.Get(ctx, common.LDAPUID).GetString(),
		Filter:            mgr.Get(ctx, common.LDAPFilter).GetString(),
		Scope:             mgr.Get(ctx, common.LDAPScope).GetInt(),
		ConnectionTimeout: mgr.Get(ctx, common.LDAPTimeout).GetInt(),
		VerifyCert:        mgr.Get(ctx, common.LDAPVerifyCert).GetBool(),
	}, nil
}

// LDAPGroupConf returns the setting of ldap group search
func LDAPGroupConf(ctx context.Context) (*cfgModels.GroupConf, error) {
	mgr := defaultMgr()
	err := mgr.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &cfgModels.GroupConf{
		BaseDN:              mgr.Get(ctx, common.LDAPGroupBaseDN).GetString(),
		Filter:              mgr.Get(ctx, common.LDAPGroupSearchFilter).GetString(),
		NameAttribute:       mgr.Get(ctx, common.LDAPGroupAttributeName).GetString(),
		SearchScope:         mgr.Get(ctx, common.LDAPGroupSearchScope).GetInt(),
		AdminDN:             mgr.Get(ctx, common.LDAPGroupAdminDn).GetString(),
		MembershipAttribute: mgr.Get(ctx, common.LDAPGroupMembershipAttribute).GetString(),
	}, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration(ctx context.Context) (int, error) {
	return defaultMgr().Get(ctx, common.TokenExpiration).GetInt(), nil
}

// RobotTokenDuration returns the token expiration time of robot account (in minute)
func RobotTokenDuration(ctx context.Context) int {
	return defaultMgr().Get(ctx, common.RobotTokenDuration).GetInt()
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration(ctx context.Context) (bool, error) {
	return defaultMgr().Get(ctx, common.SelfRegistration).GetBool(), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject(ctx context.Context) (bool, error) {
	err := defaultMgr().Load(ctx)
	if err != nil {
		return true, err
	}
	return defaultMgr().Get(ctx, common.ProjectCreationRestriction).GetString() == common.ProCrtRestrAdmOnly, nil
}

// Email returns email server settings
func Email(ctx context.Context) (*cfgModels.Email, error) {
	mgr := defaultMgr()
	err := mgr.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &cfgModels.Email{
		Host:     mgr.Get(ctx, common.EmailHost).GetString(),
		Port:     mgr.Get(ctx, common.EmailPort).GetInt(),
		Username: mgr.Get(ctx, common.EmailUsername).GetString(),
		Password: mgr.Get(ctx, common.EmailPassword).GetString(),
		SSL:      mgr.Get(ctx, common.EmailSSL).GetBool(),
		From:     mgr.Get(ctx, common.EmailFrom).GetString(),
		Identity: mgr.Get(ctx, common.EmailIdentity).GetString(),
		Insecure: mgr.Get(ctx, common.EmailInsecure).GetBool(),
	}, nil
}

// UAASettings returns the UAASettings to access UAA service.
func UAASettings(ctx context.Context) (*models.UAASettings, error) {
	mgr := defaultMgr()
	err := mgr.Load(ctx)
	if err != nil {
		return nil, err
	}
	us := &models.UAASettings{
		Endpoint:     mgr.Get(ctx, common.UAAEndpoint).GetString(),
		ClientID:     mgr.Get(ctx, common.UAAClientID).GetString(),
		ClientSecret: mgr.Get(ctx, common.UAAClientSecret).GetString(),
		VerifyCert:   mgr.Get(ctx, common.UAAVerifyCert).GetBool(),
	}
	return us, nil
}

// ReadOnly returns a bool to indicates if Harbor is in read only mode.
func ReadOnly(ctx context.Context) bool {
	return defaultMgr().Get(ctx, common.ReadOnly).GetBool()
}

// HTTPAuthProxySetting returns the setting of HTTP Auth proxy.  the settings are only meaningful when the auth_mode is
// set to http_auth
func HTTPAuthProxySetting(ctx context.Context) (*cfgModels.HTTPAuthProxy, error) {
	mgr := defaultMgr()
	if err := mgr.Load(ctx); err != nil {
		return nil, err
	}
	return &cfgModels.HTTPAuthProxy{
		Endpoint:            mgr.Get(ctx, common.HTTPAuthProxyEndpoint).GetString(),
		TokenReviewEndpoint: mgr.Get(ctx, common.HTTPAuthProxyTokenReviewEndpoint).GetString(),
		AdminGroups:         SplitAndTrim(mgr.Get(ctx, common.HTTPAuthProxyAdminGroups).GetString(), ","),
		AdminUsernames:      SplitAndTrim(mgr.Get(ctx, common.HTTPAuthProxyAdminUsernames).GetString(), ","),
		VerifyCert:          mgr.Get(ctx, common.HTTPAuthProxyVerifyCert).GetBool(),
		SkipSearch:          mgr.Get(ctx, common.HTTPAuthProxySkipSearch).GetBool(),
		ServerCertificate:   mgr.Get(ctx, common.HTTPAuthProxyServerCertificate).GetString(),
	}, nil
}

// OIDCSetting returns the setting of OIDC provider, currently there's only one OIDC provider allowed for Harbor and it's
// only effective when auth_mode is set to oidc_auth
func OIDCSetting(ctx context.Context) (*cfgModels.OIDCSetting, error) {
	mgr := defaultMgr()
	if err := mgr.Load(ctx); err != nil {
		return nil, err
	}
	scopeStr := mgr.Get(ctx, common.OIDCScope).GetString()
	extEndpoint := strings.TrimSuffix(mgr.Get(nil, common.ExtEndpoint).GetString(), "/")
	scope := SplitAndTrim(scopeStr, ",")
	return &cfgModels.OIDCSetting{
		Name:               mgr.Get(ctx, common.OIDCName).GetString(),
		Endpoint:           mgr.Get(ctx, common.OIDCEndpoint).GetString(),
		VerifyCert:         mgr.Get(ctx, common.OIDCVerifyCert).GetBool(),
		AutoOnboard:        mgr.Get(ctx, common.OIDCAutoOnboard).GetBool(),
		ClientID:           mgr.Get(ctx, common.OIDCCLientID).GetString(),
		ClientSecret:       mgr.Get(ctx, common.OIDCClientSecret).GetString(),
		GroupsClaim:        mgr.Get(ctx, common.OIDCGroupsClaim).GetString(),
		AdminGroup:         mgr.Get(ctx, common.OIDCAdminGroup).GetString(),
		RedirectURL:        extEndpoint + common.OIDCCallbackPath,
		Scope:              scope,
		UserClaim:          mgr.Get(ctx, common.OIDCUserClaim).GetString(),
		EmailClaim:         mgr.Get(ctx, common.OIDCEmailClaim).GetString(),
		ExtraRedirectParms: mgr.Get(ctx, common.OIDCExtraRedirectParms).GetStringToStringMap(),
	}, nil
}

// NotificationEnable returns a bool to indicates if notification enabled in harbor
func NotificationEnable(ctx context.Context) bool {
	return defaultMgr().Get(ctx, common.NotificationEnable).GetBool()
}

// QuotaPerProjectEnable returns a bool to indicates if quota per project enabled in harbor
func QuotaPerProjectEnable(ctx context.Context) bool {
	return defaultMgr().Get(ctx, common.QuotaPerProjectEnable).GetBool()
}

// QuotaSetting returns the setting of quota.
func QuotaSetting(ctx context.Context) (*cfgModels.QuotaSetting, error) {
	if err := defaultMgr().Load(ctx); err != nil {
		return nil, err
	}
	return &cfgModels.QuotaSetting{
		StoragePerProject: defaultMgr().Get(ctx, common.StoragePerProject).GetInt64(),
	}, nil
}

// RobotPrefix user defined robot name prefix.
func RobotPrefix(ctx context.Context) string {
	return defaultMgr().Get(ctx, common.RobotNamePrefix).GetString()
}

// SplitAndTrim ...
func SplitAndTrim(s, sep string) []string {
	res := make([]string, 0)
	for _, s := range strings.Split(s, sep) {
		if e := strings.TrimSpace(s); len(e) > 0 {
			res = append(res, e)
		}
	}
	return res
}
