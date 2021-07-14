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

package oidc

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/orm"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	conf := map[string]interface{}{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}

	config.InitWithSettings(conf, kp)

	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}
func TestHelperLoadConf(t *testing.T) {
	testP := &providerHelper{}
	assert.Nil(t, testP.setting.Load())
	err := testP.reloadSetting()
	assert.Nil(t, err)
	assert.Equal(t, "test", testP.setting.Load().(cfgModels.OIDCSetting).Name)
}

func TestHelperCreate(t *testing.T) {
	testP := &providerHelper{}
	err := testP.reloadSetting()
	assert.Nil(t, err)
	assert.Nil(t, testP.instance.Load())
	err = testP.create()
	assert.Nil(t, err)
	assert.NotNil(t, testP.instance.Load())
	assert.True(t, time.Now().Sub(testP.creationTime) < 2*time.Second)
}

func TestHelperGet(t *testing.T) {
	testP := &providerHelper{}
	p, err := testP.get()
	assert.Nil(t, err)
	assert.Equal(t, "https://oauth2.googleapis.com/token", p.Endpoint().TokenURL)

	update := map[string]interface{}{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "new-secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	ctx := orm.Context()
	config.GetCfgManager(ctx).UpdateConfig(ctx, update)

	t.Log("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)
	assert.Equal(t, "new-secret", testP.setting.Load().(cfgModels.OIDCSetting).ClientSecret)
}

func TestAuthCodeURL(t *testing.T) {
	conf := map[string]interface{}{
		common.OIDCName:               "test",
		common.OIDCEndpoint:           "https://accounts.google.com",
		common.OIDCVerifyCert:         "true",
		common.OIDCScope:              "openid, profile, offline_access",
		common.OIDCCLientID:           "client",
		common.OIDCClientSecret:       "secret",
		common.ExtEndpoint:            "https://harbor.test",
		common.OIDCExtraRedirectParms: `{"test_key":"test_value"}`,
	}
	ctx := orm.Context()
	config.GetCfgManager(ctx).UpdateConfig(ctx, conf)
	res, err := AuthCodeURL("random")
	assert.Nil(t, err)
	u, err := url.ParseRequestURI(res)
	assert.Nil(t, err)
	q, err := url.ParseQuery(u.RawQuery)
	assert.Nil(t, err)
	assert.Equal(t, "test_value", q.Get("test_key"))
	assert.Equal(t, "offline", q.Get("access_type"))
	assert.False(t, strings.Contains(q.Get("scope"), "offline_access"))
}

func TestTestEndpoint(t *testing.T) {
	c1 := Conn{
		URL:        googleEndpoint,
		VerifyCert: true,
	}
	c2 := Conn{
		URL:        "https://www.baidu.com",
		VerifyCert: false,
	}
	assert.Nil(t, TestEndpoint(c1))
	assert.NotNil(t, TestEndpoint(c2))
}

type fakeClaims struct {
	claims map[string]interface{}
}

func (fc *fakeClaims) Claims(n interface{}) error {
	b, err := json.Marshal(fc.claims)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, n)
}

func TestGroupsFromClaim(t *testing.T) {
	in := map[string]interface{}{
		"user":     "user1",
		"groups":   []interface{}{"group1", "group2"},
		"groups_2": []interface{}{"group1", "group2", 2},
	}

	m := []struct {
		input  map[string]interface{}
		key    string
		expect []string
		ok     bool
	}{
		{
			in,
			"user",
			[]string{},
			false,
		},
		{
			in,
			"prg",
			[]string{},
			false,
		},
		{
			in,
			"groups",
			[]string{"group1", "group2"},
			true,
		},
		{
			in,
			"groups_2",
			[]string{"group1", "group2"},
			true,
		},
	}

	for _, tc := range m {

		r, ok := groupsFromClaims(&fakeClaims{tc.input}, tc.key)
		assert.Equal(t, tc.expect, r)
		assert.Equal(t, tc.ok, ok)
	}
}

func TestUserInfoFromClaims(t *testing.T) {
	s := []struct {
		input   map[string]interface{}
		setting cfgModels.OIDCSetting
		expect  *UserInfo
	}{
		{
			input: map[string]interface{}{
				"name":   "Daniel",
				"email":  "daniel@gmail.com",
				"groups": []interface{}{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t1",
				GroupsClaim: "grouplist",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "Daniel",
				Email:         "daniel@gmail.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
		},
		{
			input: map[string]interface{}{
				"name":   "Daniel",
				"email":  "daniel@gmail.com",
				"groups": []interface{}{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t2",
				GroupsClaim: "groups",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:           "",
				Subject:          "",
				Username:         "Daniel",
				Email:            "daniel@gmail.com",
				Groups:           []string{"g1", "g2"},
				AdminGroupMember: true,
				hasGroupClaim:    true,
			},
		},
		{
			input: map[string]interface{}{
				"iss":        "issuer",
				"sub":        "subject000",
				"name":       "jack",
				"email":      "jack@gmail.com",
				"groupclaim": []interface{}{},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t3",
				GroupsClaim: "groupclaim",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:           "issuer",
				Subject:          "subject000",
				Username:         "jack",
				Email:            "jack@gmail.com",
				Groups:           []string{},
				hasGroupClaim:    true,
				AdminGroupMember: false,
			},
		},
		{
			input: map[string]interface{}{
				"name":   "Alvaro",
				"email":  "airadier@gmail.com",
				"groups": []interface{}{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t4",
				GroupsClaim: "grouplist",
				UserClaim:   "email",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:           "",
				Subject:          "",
				Username:         "airadier@gmail.com",
				Email:            "airadier@gmail.com",
				Groups:           []string{},
				hasGroupClaim:    false,
				AdminGroupMember: false,
			},
		},
	}
	for _, tc := range s {
		out, err := userInfoFromClaims(&fakeClaims{tc.input}, tc.setting)
		assert.Nil(t, err)
		assert.Equal(t, *tc.expect, *out)
	}
}

func TestMergeUserInfo(t *testing.T) {
	s := []struct {
		fromInfo    *UserInfo
		fromIDToken *UserInfo
		expected    *UserInfo
	}{
		{
			fromInfo: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "daniel",
				Email:         "daniel@gmail.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
			fromIDToken: &UserInfo{
				Issuer:        "issuer-google",
				Subject:       "subject-daniel",
				Username:      "daniel",
				Email:         "daniel@yahoo.com",
				Groups:        []string{"developers", "everyone"},
				hasGroupClaim: true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-google",
				Subject:       "subject-daniel",
				Username:      "daniel",
				Email:         "daniel@gmail.com",
				Groups:        []string{"developers", "everyone"},
				hasGroupClaim: true,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "tom",
				Email:         "tom@gmail.com",
				Groups:        nil,
				hasGroupClaim: false,
			},
			fromIDToken: &UserInfo{
				Issuer:        "issuer-okta",
				Subject:       "subject-jiangtan",
				Username:      "tom",
				Email:         "tom@okta.com",
				Groups:        []string{"nouse"},
				hasGroupClaim: false,
			},
			expected: &UserInfo{
				Issuer:        "issuer-okta",
				Subject:       "subject-jiangtan",
				Username:      "tom",
				Email:         "tom@gmail.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "jim",
				Email:         "jim@gmail.com",
				Groups:        []string{},
				hasGroupClaim: true,
			},
			fromIDToken: &UserInfo{
				Issuer:        "issuer-yahoo",
				Subject:       "subject-jim",
				Username:      "jim",
				Email:         "jim@yaoo.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-yahoo",
				Subject:       "subject-jim",
				Username:      "jim",
				Email:         "jim@gmail.com",
				Groups:        []string{},
				hasGroupClaim: true,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "",
				Email:         "kevin@whatever.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
			fromIDToken: &UserInfo{
				Issuer:        "issuer-whatever",
				Subject:       "subject-kevin",
				Username:      "kevin",
				Email:         "kevin@whatever.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-whatever",
				Subject:       "subject-kevin",
				Username:      "kevin",
				Email:         "kevin@whatever.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
			},
		},
	}

	for _, tc := range s {
		m := mergeUserInfo(tc.fromInfo, tc.fromIDToken)
		assert.Equal(t, *tc.expected, *m)
	}
}

func TestInjectGroupsToUser(t *testing.T) {
	cases := []struct {
		userInfo *UserInfo
		old      *models.User
		new      *models.User
	}{
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{},
				hasGroupClaim:    true,
				AdminGroupMember: false,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
		},
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{"1", "abc"},
				hasGroupClaim:    true,
				AdminGroupMember: true,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: true,
			},
		},
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{"1", "2"},
				hasGroupClaim:    true,
				AdminGroupMember: true,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{1, 2},
				AdminRoleInAuth: true,
			},
		},
	}
	for _, c := range cases {
		u := c.old
		InjectGroupsToUser(c.userInfo, u, mockPopulateGroups)
		assert.Equal(t, *c.new, *u)
	}
}
