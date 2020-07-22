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
package config

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

// test functions under package core/config
func TestConfig(t *testing.T) {
	test.InitDatabaseFromEnv()
	dao.PrepareTestData([]string{"delete from properties where k='scan_all_policy'"}, []string{})
	defaultCACertPath = path.Join(currPath(), "test", "ca.crt")
	c := map[string]interface{}{
		common.WithClair:       false,
		common.WithChartMuseum: false,
		common.WithNotary:      false,
	}
	Init()

	Upload(c)

	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)
	assert := assert.New(t)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}
	oriKeyPath := os.Getenv("TOKEN_PRIVATE_KEY_PATH")
	if err := os.Setenv("TOKEN_PRIVATE_KEY_PATH", ""); err != nil {
		t.Fatalf("failed to set env %s: %v", "TOKEN_PRIVATE_KEY_PATH", err)
	}
	defer os.Setenv("TOKEN_PRIVATE_KEY_PATH", oriKeyPath)

	os.Setenv("JOBSERVICE_URL", "http://myjob:8888")
	os.Setenv("GC_TIME_WINDOW_HOURS", "0")

	Init()

	if err := Load(); err != nil {
		t.Fatalf("failed to load configurations: %v", err)
	}

	if err := Upload(map[string]interface{}{}); err != nil {
		t.Fatalf("failed to upload configurations: %v", err)
	}

	if _, err := GetSystemCfg(); err != nil {
		t.Fatalf("failed to get system configurations: %v", err)
	}

	mode, err := AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}

	if _, err := LDAPConf(); err != nil {
		t.Fatalf("failed to get ldap settings: %v", err)
	}

	if _, err := LDAPGroupConf(); err != nil {
		t.Fatalf("failed to get ldap group settings: %v", err)
	}

	if _, err := TokenExpiration(); err != nil {
		t.Fatalf("failed to get token expiration: %v", err)
	}

	tkExp := RobotTokenDuration()
	assert.Equal(tkExp, 43200)

	if _, err := ExtEndpoint(); err != nil {
		t.Fatalf("failed to get domain name: %v", err)
	}

	if _, err := SecretKey(); err != nil {
		t.Fatalf("failed to get secret key: %v", err)
	}

	if _, err := SelfRegistration(); err != nil {
		t.Fatalf("failed to get self registration: %v", err)
	}

	if _, err := RegistryURL(); err != nil {
		t.Fatalf("failed to get registry URL: %v", err)
	}

	if len(InternalJobServiceURL()) == 0 {
		t.Error("the internal job service url is null")
	}

	if len(InternalTokenServiceEndpoint()) == 0 {
		t.Error("the internal token service endpoint is null")
	}

	if _, err := InitialAdminPassword(); err != nil {
		t.Fatalf("failed to get initial admin password: %v", err)
	}

	if _, err := OnlyAdminCreateProject(); err != nil {
		t.Fatalf("failed to get onldy admin create project: %v", err)
	}

	if _, err := Email(); err != nil {
		t.Fatalf("failed to get email settings: %v", err)
	}

	if _, err := Database(); err != nil {
		t.Fatalf("failed to get database: %v", err)
	}

	defaultConfig := test.GetDefaultConfigMap()
	Upload(defaultConfig)

	if InternalNotaryEndpoint() != "http://notary-server:4443" {
		t.Errorf("Unexpected notary endpoint: %s", InternalNotaryEndpoint())
	}
	if WithNotary() {
		t.Errorf("Withnotary should be false")
	}
	if WithClair() {
		t.Errorf("WithClair should be false")
	}
	if ReadOnly() {
		t.Errorf("ReadOnly should be false")
	}

	extURL, err := ExtURL()
	if err != nil {
		t.Errorf("Unexpected error getting external URL: %v", err)
	}
	if extURL != "host01.com" {
		t.Errorf(`extURL should be "host01.com".`)
	}

	mode, err = AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}

	if tokenKeyPath := TokenPrivateKeyPath(); tokenKeyPath != "/etc/core/private_key.pem" {
		t.Errorf("Unexpected token private key path: %s, expected: %s", tokenKeyPath, "/etc/core/private_key.pem")
	}

	us, err := UAASettings()
	if err != nil {
		t.Fatalf("failed to get UAA setting, error: %v", err)
	}

	if us.ClientID != "testid" || us.ClientSecret != "testsecret" || us.Endpoint != "10.192.168.5" || us.VerifyCert {
		t.Errorf("Unexpected UAA setting: %+v", *us)
	}
	assert.Equal("http://myjob:8888", InternalJobServiceURL())
	assert.Equal("http://myui:8888/service/token", InternalTokenServiceEndpoint())

	localCoreURL := LocalCoreURL()
	assert.Equal("http://127.0.0.1:8080", localCoreURL)

	assert.True(NotificationEnable())
	assert.Equal(int64(0), GetGCTimeWindow())

}

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

func TestHTTPAuthProxySetting(t *testing.T) {
	certificate := `-----BEGIN CERTIFICATE-----
MIIFXzCCA0egAwIBAgIUY7f2ECRISPMeb1iVNvV5iQsIErUwDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UEBwwIQmVpIEpp
bmcxDzANBgNVBAoMBlZNd2FyZTERMA8GA1UEAwwISGFyYm9yQ0EwHhcNMTkxMTE2
MjI1NjQ0WhcNMjAxMTE1MjI1NjQ0WjBdMQswCQYDVQQGEwJDTjEMMAoGA1UECAwD
UEVLMREwDwYDVQQHDAhCZWkgSmluZzEPMA0GA1UECgwGVk13YXJlMRwwGgYDVQQD
DBNqdC1kZXYuaGFyYm9yLmxvY2FsMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIIC
CgKCAgEAwQ/TQTHwnHwEC/KyHP8Tyv/v35GwXRGW6s1MYoqVnyQMPud0scLHAA2u
PZv2F5jy7PtnhcR0ZHGf05L/igY1utV7/4F2aFgOq0ExYMKxvzilitdcvsxmfTLI
m2pwS8+kH/1s1xR9/7ZPlPSdHuxcHgjtMqorljJykRyq0RBLvXCG+fmAY91kdLil
XWiuIU73lNpZHuXEDl4m2XUzb9cuhwvaHYs7aT6BhwqAJZUjwURUqMe1PIOo7vkQ
cKUHe3u3Fg/vbxfecEr3AHcKfIqm5fwI9vdzj5BP3lGT9hrxduAwI6SgehxGGWP4
aN/cKGIKt/2kzgFoQi/d5p3RBkLVNP/sEyAt9dLJj12ovkQwJzdKDVOy50t3ws9g
Mf3rUUb/wdZADK26lxolep9EXVe4kuWpOo1RvdI+lJJvWc3QaJIoVbr9LM8QN3e7
Iyk3pYRyaQj9EKZ4k0RgWVbIZfRLy1LkGMqmCcIqunHVdGDDjbO/ri8z0sKocMGl
qrqcBTPYmsau7PEcfzJY8k/HUDYdhZgIv2C1iLBl6eoTVDRbrGFcu8LzleWx2/19
OA1Cx7S8WyzN+9mjygqwEYc6qMtoeutAkOA5J8JkxBp0yqjUEnAB6E7R07xQP8AY
IKq5oVpkbD8WRI3w7l/X0AAkDtnijbgYWTfPVGihXHhRtkr/b9cCAwEAAaMiMCAw
HgYDVR0RBBcwFYITanQtZGV2LmhhcmJvci5sb2NhbDANBgkqhkiG9w0BAQsFAAOC
AgEAqwU10WwhI5W8w62vOpT+PKSXRVjHKhm3ltaIzAN7S772hiGl6L81cP9dXZ5y
FN0tFUtUVK01JRJHJaduXNwx24HlwRPNp7mLa4IPpeeVfG14/QCoOd8vxHtKG+V6
zE7Jx2FBVfUJ7P4SngEv4QfvZPt+lCXGK3V1RRTpkLD2knhBfu85rjPi+VW56Z7b
Jb2IEmVRlfR7Z0oYm8z3Obt2XuLIC1/8NtfxthggKr6DeSwZSJXrdGVbyvdiAmk2
iHQch0+UTkRDinL0je6WWbxBoAPXsWA9Hc69o8kmjcXUa99/i8FrC2QxPDUoxxMn
1zWk0jct2Tsr3VZ5HnaI5e8ifG7RUcE5Vr6w7MI5P44Q88zhboP1ShYQ/s513cu2
heELKvO3+mqv96lERtkUUwe8tm1zoPKzQI6ecGuqaTcMbXAGax+ud5XnUlz4xzTI
cByAsQ9DNhYIcOftnfz349zkHeWmMum4uiQwfp/+OrqX+O8U0eJYhlfu9vqCU05T
3mE8Hw5veNdLaZx+mzUVIDzrOB3fh/O62J9CsaZKtxwgLlGiT2ltuC1xUqn3DL8s
pkgODrJUf0p5dhcnLyA2nZolRV1rtwlgJstnEV4JpG1MwtmAZYZUilLvnfpVxTtA
y1bQusZMygQezfCuEzsewF+OpANFovCTUEs6s5vyoVNP8lk=
-----END CERTIFICATE-----
`
	m := map[string]interface{}{
		common.HTTPAuthProxySkipSearch:        "true",
		common.HTTPAuthProxyVerifyCert:        "true",
		common.HTTPAuthProxyEndpoint:          "https://auth.proxy/suffix",
		common.HTTPAuthProxyServerCertificate: certificate,
	}
	InitWithSettings(m)
	v, e := HTTPAuthProxySetting()
	assert.Nil(t, e)
	assert.Equal(t, *v, models.HTTPAuthProxy{
		Endpoint:          "https://auth.proxy/suffix",
		SkipSearch:        true,
		VerifyCert:        true,
		ServerCertificate: certificate,
	})
}

func TestOIDCSetting(t *testing.T) {
	m := map[string]interface{}{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://oidc.test",
		common.OIDCVerifyCert:   "true",
		common.OIDCAutoOnboard:  "false",
		common.OIDCScope:        "openid, profile",
		common.OIDCGroupsClaim:  "my_group",
		common.OIDCUserClaim:    "username",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	InitWithSettings(m)
	v, e := OIDCSetting()
	assert.Nil(t, e)
	assert.Equal(t, "test", v.Name)
	assert.Equal(t, "https://oidc.test", v.Endpoint)
	assert.True(t, v.VerifyCert)
	assert.Equal(t, "my_group", v.GroupsClaim)
	assert.False(t, v.AutoOnboard)
	assert.Equal(t, "client", v.ClientID)
	assert.Equal(t, "secret", v.ClientSecret)
	assert.Equal(t, "https://harbor.test/c/oidc/callback", v.RedirectURL)
	assert.ElementsMatch(t, []string{"openid", "profile"}, v.Scope)
	assert.Equal(t, "username", v.UserClaim)
}
