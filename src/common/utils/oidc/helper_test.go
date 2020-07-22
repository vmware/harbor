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
	"github.com/goharbor/harbor/src/common"
	config2 "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	conf := map[string]interface{}{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	kp := &config2.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}

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
	assert.Equal(t, "test", testP.setting.Load().(models.OIDCSetting).Name)
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
	config.GetCfgManager().UpdateConfig(update)

	t.Log("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)
	assert.Equal(t, "new-secret", testP.setting.Load().(models.OIDCSetting).ClientSecret)
}

func TestAuthCodeURL(t *testing.T) {
	res, err := AuthCodeURL("random")
	assert.Nil(t, err)
	u, err := url.ParseRequestURI(res)
	assert.Nil(t, err)
	q, err := url.ParseQuery(u.RawQuery)
	assert.Nil(t, err)
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

		r, ok := GroupsFromClaims(&fakeClaims{tc.input}, tc.key)
		assert.Equal(t, tc.expect, r)
		assert.Equal(t, tc.ok, ok)
	}
}

func TestUserInfoFromClaims(t *testing.T) {
	s := []struct {
		input      map[string]interface{}
		groupClaim string
		userClaim  string
		expect     *UserInfo
	}{
		{
			input: map[string]interface{}{
				"name":   "Daniel",
				"email":  "daniel@gmail.com",
				"groups": []interface{}{"g1", "g2"},
			},
			groupClaim: "grouplist",
			userClaim:  "",
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
			groupClaim: "groups",
			userClaim:  "",
			expect: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "Daniel",
				Email:         "daniel@gmail.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
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
			groupClaim: "groupclaim",
			userClaim:  "",
			expect: &UserInfo{
				Issuer:        "issuer",
				Subject:       "subject000",
				Username:      "jack",
				Email:         "jack@gmail.com",
				Groups:        []string{},
				hasGroupClaim: true,
			},
		},
		{
			input: map[string]interface{}{
				"name":   "Alvaro",
				"email":  "airadier@gmail.com",
				"groups": []interface{}{"g1", "g2"},
			},
			groupClaim: "grouplist",
			userClaim:  "email",
			expect: &UserInfo{
				Issuer:        "",
				Subject:       "",
				Username:      "airadier@gmail.com",
				Email:         "airadier@gmail.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
		},
	}
	for _, tc := range s {
		out, err := userInfoFromClaims(&fakeClaims{tc.input}, tc.groupClaim, tc.userClaim)
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
	}

	for _, tc := range s {
		m := mergeUserInfo(tc.fromInfo, tc.fromIDToken)
		assert.Equal(t, *tc.expected, *m)
	}
}
