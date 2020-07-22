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

package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	authorizationHeader = "Authorization"
)

// AuthHandlerTestSuite is test suite for testing auth handler
type AuthHandlerTestSuite struct {
	suite.Suite
}

// TestAuthHandler is the entry method of running AuthHandlerTestSuite
func TestAuthHandler(t *testing.T) {
	suite.Run(t, &AuthHandlerTestSuite{})
}

// TestNoneHandler test none handler
func (suite *AuthHandlerTestSuite) TestNoneHandler() {
	none := &NoneAuthHandler{}
	suite.Equal(AuthModeNone, none.Mode(), "auth mode None")
	r, err := http.NewRequest(http.MethodGet, "https://p2p.none.com", nil)
	require.NoError(suite.T(), err, "new HTTP request")
	err = none.Authorize(r, nil)
	require.NoError(suite.T(), err, "authorize HTTP request")
	suite.Equal(0, len(r.Header.Get(authorizationHeader)), "check authorization header")
}

// TestBasicHandler test basic auth handler
func (suite *AuthHandlerTestSuite) TestBasicHandler() {
	basic := &BasicAuthHandler{}
	suite.Equal(AuthModeBasic, basic.Mode(), "auth mode basic")
	r, err := http.NewRequest(http.MethodGet, "https://p2p.basic.com", nil)
	require.NoError(suite.T(), err, "new HTTP request")
	cred := &Credential{
		Mode: AuthModeBasic,
		Data: map[string]string{
			"username": "password",
		},
	}
	err = basic.Authorize(r, cred)
	require.NoError(suite.T(), err, "authorize HTTP request")
	encodedStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "username", "password")))
	suite.Equal(fmt.Sprintf("%s %s", "Basic", encodedStr), r.Header.Get(authorizationHeader), "check basic authorization header")
}

// TestTokenHandler test token auth handler
func (suite *AuthHandlerTestSuite) TestTokenHandler() {
	token := &TokenAuthHandler{}
	suite.Equal(AuthModeOAuth, token.Mode(), "auth mode token")
	r, err := http.NewRequest(http.MethodGet, "https://p2p.token.com", nil)
	require.NoError(suite.T(), err, "new HTTP request")
	cred := &Credential{
		Mode: AuthModeOAuth,
		Data: map[string]string{
			"token": "my-token",
		},
	}
	err = token.Authorize(r, cred)
	require.NoError(suite.T(), err, "authorize HTTP request")
	suite.Equal("Bearer my-token", r.Header.Get(authorizationHeader), "check token authorization header")
}

// TestCustomHandler test custom auth handler
func (suite *AuthHandlerTestSuite) TestCustomHandler() {
	custom := &CustomAuthHandler{}
	suite.Equal(AuthModeCustom, custom.Mode(), "auth mode custom")
	r, err := http.NewRequest(http.MethodGet, "https://p2p.custom.com", nil)
	require.NoError(suite.T(), err, "new HTTP request")
	cred := &Credential{
		Mode: AuthModeCustom,
		Data: map[string]string{
			"api-key": "my-api-key",
		},
	}
	err = custom.Authorize(r, cred)
	require.NoError(suite.T(), err, "authorize HTTP request")
	suite.Equal("my-api-key", r.Header.Get("api-key"), "check custom authorization header")
}
