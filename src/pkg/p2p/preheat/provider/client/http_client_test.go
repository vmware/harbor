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

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/stretchr/testify/suite"
)

// HTTPClientTestSuite is a test suite for testing the HTTP client.
type HTTPClientTestSuite struct {
	suite.Suite

	ts *httptest.Server
}

// TestHTTPClient is the entry of running HttpClientTestSuite.
func TestHTTPClient(t *testing.T) {
	suite.Run(t, &HTTPClientTestSuite{})
}

// SetupSuite prepares the env for the test suite.
func (suite *HTTPClientTestSuite) SetupSuite() {
	suite.ts = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Authorization")
		if len(a) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Add("Content-type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))

	suite.ts.StartTLS()
}

// TearDownSuite clears the env for the test suite.
func (suite *HTTPClientTestSuite) TearDownSuite() {
	suite.ts.Close()
}

// TestClientGet tests the client getter method.
func (suite *HTTPClientTestSuite) TestClientGet() {
	c := GetHTTPClient(true)
	suite.NotNil(c, "get insecure HTTP client")
	t := c.internalClient.Transport.(*http.Transport)
	suite.Equal(true, t.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify=true")

	c2 := GetHTTPClient(false)
	suite.NotNil(c2, "get secure HTTP client")
	t2 := c2.internalClient.Transport.(*http.Transport)
	suite.Equal(false, t2.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify=false")
}

// TestGet test the Get method
func (suite *HTTPClientTestSuite) TestGet() {
	c := GetHTTPClient(true)
	suite.NotNil(c, "get insecure HTTP client")

	_, err := c.Get(suite.ts.URL, nil, nil, nil)
	suite.Error(err, "unauthorized error", err)

	cred := &auth.Credential{
		Mode: auth.AuthModeBasic,
		Data: map[string]string{"username": "password"},
	}
	data, err := c.Get(suite.ts.URL, cred, map[string]string{"name": "TestGet"}, map[string]string{"Accept": "application/json"})
	suite.NoError(err, "get data")
	suite.Equal("{}", string(data), "get json data")
}

// TestPost test the Post method
func (suite *HTTPClientTestSuite) TestPost() {
	c := GetHTTPClient(true)
	suite.NotNil(c, "get insecure HTTP client")

	cred := &auth.Credential{
		Mode: auth.AuthModeBasic,
		Data: map[string]string{"username": "password"},
	}
	data, err := c.Post(suite.ts.URL, cred, []byte("{}"), map[string]string{"Accept": "application/json"})
	suite.NoError(err, "post data")
	suite.Equal("{}", string(data), "post json data")
}
