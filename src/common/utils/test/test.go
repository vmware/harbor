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
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"fmt"
	"os"
	"sort"

	"github.com/goharbor/harbor/src/common"
	"github.com/gorilla/mux"
)

// RequestHandlerMapping is a mapping between request and its handler
type RequestHandlerMapping struct {
	// Method is the method the request used
	Method string
	// Pattern is the pattern the request must match
	Pattern string
	// Handler is the handler which handles the request
	Handler func(http.ResponseWriter, *http.Request)
}

// ServeHTTP ...
func (rhm *RequestHandlerMapping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(rhm.Method) != 0 && r.Method != strings.ToUpper(rhm.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rhm.Handler(w, r)
}

// Response is a response used for unit test
type Response struct {
	// StatusCode is the status code of the response
	StatusCode int
	// Headers are the headers of the response
	Headers map[string]string
	// Boby is the body of the response
	Body []byte
}

// Handler returns a handler function which handle requst according to
// the response provided
func Handler(resp *Response) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if resp == nil {
			return
		}

		for k, v := range resp.Headers {
			w.Header().Add(http.CanonicalHeaderKey(k), v)
		}

		if resp.StatusCode == 0 {
			resp.StatusCode = http.StatusOK
		}
		w.WriteHeader(resp.StatusCode)

		if len(resp.Body) != 0 {
			io.Copy(w, bytes.NewReader(resp.Body))
		}
	}
}

// NewServer creates a HTTP server for unit test
func NewServer(mappings ...*RequestHandlerMapping) *httptest.Server {
	r := mux.NewRouter()

	for _, mapping := range mappings {
		r.PathPrefix(mapping.Pattern).Handler(mapping).Methods(mapping.Method)
	}

	return httptest.NewServer(r)
}

// GetUnitTestConfig ...
func GetUnitTestConfig() map[string]interface{} {
	ipAddress := os.Getenv("IP")
	return map[string]interface{}{
		common.ExtEndpoint:            fmt.Sprintf("https://%s", ipAddress),
		common.AUTHMode:               "db_auth",
		common.DatabaseType:           "postgresql",
		common.PostGreSQLHOST:         ipAddress,
		common.PostGreSQLPort:         5432,
		common.PostGreSQLUsername:     "postgres",
		common.PostGreSQLPassword:     "root123",
		common.PostGreSQLDatabase:     "registry",
		common.LDAPURL:                "ldap://ldap.vmware.com",
		common.LDAPSearchDN:           "cn=admin,dc=example,dc=com",
		common.LDAPSearchPwd:          "admin",
		common.LDAPBaseDN:             "dc=example,dc=com",
		common.LDAPUID:                "uid",
		common.LDAPFilter:             "",
		common.LDAPScope:              2,
		common.LDAPTimeout:            30,
		common.LDAPVerifyCert:         true,
		common.UAAVerifyCert:          true,
		common.AdminInitialPassword:   "Harbor12345",
		common.LDAPGroupSearchFilter:  "objectclass=groupOfNames",
		common.LDAPGroupBaseDN:        "dc=example,dc=com",
		common.LDAPGroupAttributeName: "cn",
		common.LDAPGroupSearchScope:   2,
		common.LDAPGroupAdminDn:       "cn=harbor_users,ou=groups,dc=example,dc=com",
		common.WithNotary:             "false",
		common.WithChartMuseum:        "false",
		common.SelfRegistration:       "true",
		common.WithClair:              "true",
		common.TokenServiceURL:        "http://core:8080/service/token",
		common.RegistryURL:            fmt.Sprintf("http://%s:5000", ipAddress),
		common.ReadOnly:               false,
	}
}

// TraceCfgMap ...
func TraceCfgMap(cfgs map[string]interface{}) {
	var keys []string
	for k := range cfgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%v=%v\n", k, cfgs[k])
	}
}

// CheckSetsEqual - check int set if they are equals
func CheckSetsEqual(setA, setB []int) bool {
	if len(setA) != len(setB) {
		return false
	}
	type void struct{}
	var exist void
	setAll := make(map[int]void)
	for _, r := range setA {
		setAll[r] = exist
	}
	for _, r := range setB {
		if _, ok := setAll[r]; !ok {
			return false
		}
	}

	setAll = make(map[int]void)
	for _, r := range setB {
		setAll[r] = exist
	}
	for _, r := range setA {
		if _, ok := setAll[r]; !ok {
			return false
		}
	}
	return true

}
