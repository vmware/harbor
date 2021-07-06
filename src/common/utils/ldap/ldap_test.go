package ldap

import (
	"os"
	"reflect"
	"testing"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	uiConfig "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
)

var ldapTestConfig = map[string]interface{}{
	common.ExtEndpoint:        "host01.com",
	common.AUTHMode:           "ldap_auth",
	common.DatabaseType:       "postgresql",
	common.PostGreSQLHOST:     "127.0.0.1",
	common.PostGreSQLPort:     5432,
	common.PostGreSQLUsername: "postgres",
	common.PostGreSQLPassword: "root123",
	common.PostGreSQLDatabase: "registry",
	// config.SelfRegistration: true,
	common.LDAPURL:              "ldap://127.0.0.1",
	common.LDAPSearchDN:         "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:        "admin",
	common.LDAPBaseDN:           "dc=example,dc=com",
	common.LDAPUID:              "uid",
	common.LDAPFilter:           "",
	common.LDAPScope:            3,
	common.LDAPTimeout:          30,
	common.LDAPPageSize:         0,
	common.AdminInitialPassword: "password",
}

var defaultConfigWithVerifyCert = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.LDAPAuth,
	common.DatabaseType:               "postgresql",
	common.PostGreSQLHOST:             "127.0.0.1",
	common.PostGreSQLPort:             5432,
	common.PostGreSQLUsername:         "postgres",
	common.PostGreSQLPassword:         "root123",
	common.PostGreSQLDatabase:         "registry",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1:389",
	common.LDAPSearchDN:               "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:              "admin",
	common.LDAPBaseDN:                 "dc=example,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.LDAPPageSize:               0,
	common.LDAPVerifyCert:             true,
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.AdminInitialPassword:       "password",
	common.WithNotary:                 false,
	common.WithClair:                  false,
}

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		log.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		log.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	uiConfig.Init()

	uiConfig.Upload(ldapTestConfig)

	os.Exit(m.Run())

}

func TestLoadSystemLdapConfig(t *testing.T) {
	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	if session.ldapConfig.LdapURL != "ldap://127.0.0.1:389" {
		t.Errorf("unexpected LdapURL: %s != %s", session.ldapConfig.LdapURL, "ldap://127.0.0.1:389")
	}

}

func TestConnectTest(t *testing.T) {
	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Errorf("failed to load system ldap config")
	}
	err = session.ConnectionTest()
	if err != nil {
		t.Errorf("Unexpected ldap connect fail: %v", err)
	}

}

func TestCreateWithConfig(t *testing.T) {
	var testConfigs = []struct {
		config        models.LdapConf
		internalValue int
	}{
		{
			models.LdapConf{
				LdapScope: 3,
				LdapURL:   "ldaps://127.0.0.1",
			}, 2},
		{
			models.LdapConf{
				LdapScope: 2,
				LdapURL:   "ldaps://127.0.0.1",
			}, 1},
		{
			models.LdapConf{
				LdapScope: 1,
				LdapURL:   "ldaps://127.0.0.1",
			}, 0},
		{
			models.LdapConf{
				LdapScope: 1,
				LdapURL:   "ldaps://127.0.0.1:abc",
			}, -1},
	}

	for _, val := range testConfigs {
		_, err := CreateWithConfig(val.config)
		if val.internalValue < 0 {
			if err == nil {
				t.Fatalf("Should have error with url :%v", val.config)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Can not create with ui config, err:%v", err)
		}
	}

}

func TestSearchUser(t *testing.T) {

	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Fatalf("Can not load system ldap config")
	}
	err = session.Open()
	if err != nil {
		t.Fatalf("failed to create ldap session %v", err)
	}

	err = session.Bind(session.ldapConfig.LdapSearchDn, session.ldapConfig.LdapSearchPassword)
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}

	defer session.Close()

	result, err := session.SearchUser("test")
	if err != nil || len(result) == 0 {
		t.Fatalf("failed to search user test!")
	}

	result2, err := session.SearchUser("mike")
	if err != nil || len(result2) == 0 {
		t.Fatalf("failed to search user mike!")
	}
	if len(result2[0].GroupDNList) < 1 && result2[0].GroupDNList[0] != "cn=harbor_users,ou=groups,dc=example,dc=com" {
		t.Fatalf("failed to search user mike's memberof")
	}

}

func TestFormatURL(t *testing.T) {

	var invalidURL = "http://localhost:389"
	_, err := formatURL(invalidURL)
	if err == nil {
		t.Fatalf("Should failed on invalid URL %v", invalidURL)
		t.Fail()
	}

	var urls = []struct {
		rawURL  string
		goodURL string
	}{
		{"ldaps://127.0.0.1", "ldaps://127.0.0.1:636"},
		{"ldap://9.123.102.33", "ldap://9.123.102.33:389"},
		{"ldaps://127.0.0.1:389", "ldaps://127.0.0.1:389"},
		{"ldap://127.0.0.1:636", "ldaps://127.0.0.1:636"},
		{"112.122.122.122", "ldap://112.122.122.122:389"},
		{"ldap:\\wrong url", ""},
	}

	for _, u := range urls {
		goodURL, err := formatURL(u.rawURL)
		if u.goodURL == "" {
			if err == nil {
				t.Fatalf("Should failed on wrong url, %v", u.rawURL)
			}
			continue
		}
		if err != nil || goodURL != u.goodURL {
			t.Fatalf("Faild on URL: raw=%v, expected:%v, actual:%v", u.rawURL, u.goodURL, goodURL)
		}
	}

}

func Test_createGroupSearchFilter(t *testing.T) {
	type args struct {
		oldFilter          string
		groupName          string
		groupNameAttribute string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{"Normal Filter", args{oldFilter: "objectclass=groupOfNames", groupName: "harbor_users", groupNameAttribute: "cn"}, "(&(objectclass=groupOfNames)(cn=*harbor_users*))", nil},
		{"Empty Old", args{groupName: "harbor_users", groupNameAttribute: "cn"}, "(cn=*harbor_users*)", nil},
		{"Empty Both", args{groupNameAttribute: "cn"}, "(cn=*)", nil},
		{"Empty name", args{oldFilter: "objectclass=groupOfNames", groupNameAttribute: "cn"}, "(objectclass=groupOfNames)", nil},
		{"Empty name with complex filter", args{oldFilter: "(&(objectClass=groupOfNames)(cn=*sample*))", groupNameAttribute: "cn"}, "(&(objectClass=groupOfNames)(cn=*sample*))", nil},
		{"Empty name with bad filter", args{oldFilter: "(&(objectClass=groupOfNames),cn=*sample*)", groupNameAttribute: "cn"}, "", ErrInvalidFilter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := createGroupSearchFilter(tt.args.oldFilter, tt.args.groupName, tt.args.groupNameAttribute); got != tt.want && err != tt.wantErr {
				t.Errorf("createGroupSearchFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SearchGroup(t *testing.T) {
	type fields struct {
		ldapConfig models.LdapConf
		ldapConn   *goldap.Conn
	}
	type args struct {
		groupDN            string
		filter             string
		groupName          string
		groupNameAttribute string
	}

	ldapConfig := models.LdapConf{
		LdapURL:            ldapTestConfig[common.LDAPURL].(string) + ":389",
		LdapSearchDn:       ldapTestConfig[common.LDAPSearchDN].(string),
		LdapScope:          2,
		LdapSearchPassword: ldapTestConfig[common.LDAPSearchPwd].(string),
		LdapBaseDn:         ldapTestConfig[common.LDAPBaseDN].(string),
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.LdapGroup
		wantErr bool
	}{
		{"normal search",
			fields{ldapConfig: ldapConfig},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com", filter: "objectClass=groupOfNames", groupName: "harbor_users", groupNameAttribute: "cn"},
			[]models.LdapGroup{{GroupName: "harbor_users", GroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ldapConfig: tt.fields.ldapConfig,
				ldapConn:   tt.fields.ldapConn,
			}
			session.Open()
			defer session.Close()
			got, err := session.searchGroup(tt.args.groupDN, tt.args.filter, tt.args.groupName, tt.args.groupNameAttribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.SearchGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.SearchGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SearchGroupByDN(t *testing.T) {
	ldapConfig := models.LdapConf{
		LdapURL:            ldapTestConfig[common.LDAPURL].(string) + ":389",
		LdapSearchDn:       ldapTestConfig[common.LDAPSearchDN].(string),
		LdapScope:          2,
		LdapSearchPassword: ldapTestConfig[common.LDAPSearchPwd].(string),
		LdapBaseDn:         ldapTestConfig[common.LDAPBaseDN].(string),
	}
	ldapGroupConfig := models.LdapGroupConf{
		LdapGroupBaseDN:        "dc=example,dc=com",
		LdapGroupFilter:        "objectclass=groupOfNames",
		LdapGroupNameAttribute: "cn",
		LdapGroupSearchScope:   2,
	}
	ldapGroupConfig2 := models.LdapGroupConf{
		LdapGroupBaseDN:        "dc=example,dc=com",
		LdapGroupFilter:        "objectclass=groupOfNames",
		LdapGroupNameAttribute: "o",
		LdapGroupSearchScope:   2,
	}
	groupConfigWithEmptyBaseDN := models.LdapGroupConf{
		LdapGroupBaseDN:        "",
		LdapGroupFilter:        "(objectclass=groupOfNames)",
		LdapGroupNameAttribute: "cn",
		LdapGroupSearchScope:   2,
	}
	groupConfigWithFilter := models.LdapGroupConf{
		LdapGroupBaseDN:        "dc=example,dc=com",
		LdapGroupFilter:        "(cn=*admin*)",
		LdapGroupNameAttribute: "cn",
		LdapGroupSearchScope:   2,
	}
	groupConfigWithDifferentGroupDN := models.LdapGroupConf{
		LdapGroupBaseDN:        "dc=harbor,dc=example,dc=com",
		LdapGroupFilter:        "(objectclass=groupOfNames)",
		LdapGroupNameAttribute: "cn",
		LdapGroupSearchScope:   2,
	}

	type fields struct {
		ldapConfig      models.LdapConf
		ldapGroupConfig models.LdapGroupConf
		ldapConn        *goldap.Conn
	}
	type args struct {
		groupDN string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.LdapGroup
		wantErr bool
	}{
		{"normal search",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "harbor_users", GroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"}}, false},
		{"search non-exist group",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_non_users,ou=groups,dc=example,dc=com"},
			nil, true},
		{"search invalid group dn",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "random string"},
			nil, true},
		{"search with gid = cn",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "harbor_group", GroupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with gid = o",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: ldapGroupConfig2},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "hgroup", GroupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with empty group base dn",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: groupConfigWithEmptyBaseDN},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "harbor_group", GroupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter success",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: groupConfigWithFilter},
			args{groupDN: "cn=harbor_admin,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "harbor_admin", GroupDN: "cn=harbor_admin,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter fail",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: groupConfigWithFilter},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{}, false},
		{"search with different group base dn success",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupDN: "cn=harbor_root,dc=harbor,dc=example,dc=com"},
			[]models.LdapGroup{{GroupName: "harbor_root", GroupDN: "cn=harbor_root,dc=harbor,dc=example,dc=com"}}, false},
		{"search with different group base dn fail",
			fields{ldapConfig: ldapConfig, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupDN: "cn=harbor_guest,ou=groups,dc=example,dc=com"},
			[]models.LdapGroup{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ldapConfig:      tt.fields.ldapConfig,
				ldapGroupConfig: tt.fields.ldapGroupConfig,
				ldapConn:        tt.fields.ldapConn,
			}
			session.Open()
			defer session.Close()
			got, err := session.SearchGroupByDN(tt.args.groupDN)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.SearchGroupByDN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.SearchGroupByDN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateUserSearchFilter(t *testing.T) {
	type args struct {
		origFilter string
		ldapUID    string
		username   string
	}
	cases := []struct {
		name    string
		in      args
		want    string
		wantErr error
	}{
		{name: `Normal test`, in: args{"(objectclass=inetorgperson)", "cn", "sample"}, want: "(&(objectclass=inetorgperson)(cn=sample)", wantErr: nil},
		{name: `Bad original filter`, in: args{"(objectclass=inetorgperson)ldap*", "cn", "sample"}, want: "", wantErr: ErrInvalidFilter},
		{name: `Complex original filter`, in: args{"(&(objectclass=inetorgperson)(|(memberof=cn=harbor_users,ou=groups,dc=example,dc=com)(memberof=cn=harbor_admin,ou=groups,dc=example,dc=com)(memberof=cn=harbor_guest,ou=groups,dc=example,dc=com)))", "cn", "sample"}, want: "(&(&(objectclass=inetorgperson)(|(memberof=cn=harbor_users,ou=groups,dc=example,dc=com)(memberof=cn=harbor_admin,ou=groups,dc=example,dc=com)(memberof=cn=harbor_guest,ou=groups,dc=example,dc=com)))(cn=sample)", wantErr: nil},
		{name: `Empty original filter`, in: args{"", "cn", "sample"}, want: "(cn=sample)", wantErr: nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := createUserSearchFilter(tt.in.origFilter, tt.in.ldapUID, tt.in.origFilter)
			if got != tt.want && gotErr != tt.wantErr {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}

func TestNormalizeFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal test", args{"(objectclass=user)"}, "(objectclass=user)"},
		{"with space", args{" (objectclass=user) "}, "(objectclass=user)"},
		{"nothing", args{"objectclass=user"}, "(objectclass=user)"},
		{"and condition", args{"&(objectclass=user)(cn=admin)"}, "(&(objectclass=user)(cn=admin))"},
		{"or condition", args{"|(objectclass=user)(cn=admin)"}, "(|(objectclass=user)(cn=admin))"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFilter(tt.args.filter); got != tt.want {
				t.Errorf("normalizeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnderBaseDN(t *testing.T) {
	type args struct {
		baseDN  string
		childDN string
	}
	cases := []struct {
		name      string
		in        args
		wantError bool
		want      bool
	}{
		{
			name:      `normal`,
			in:        args{"dc=example,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      true,
		},
		{
			name:      `false`,
			in:        args{"dc=vmware,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      false,
		},
		{
			name:      `same dn`,
			in:        args{"cn=admin,dc=example,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      true,
		},
		{
			name:      `error format in base`,
			in:        args{"abc", "cn=admin,dc=example,dc=com"},
			wantError: true,
			want:      false,
		},
		{
			name:      `error format in child`,
			in:        args{"dc=vmware,dc=com", "wrong format"},
			wantError: true,
			want:      false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnderBaseDN(tt.in.baseDN, tt.in.childDN)
			if (err != nil) != tt.wantError {
				t.Errorf("UnderBaseDN error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
