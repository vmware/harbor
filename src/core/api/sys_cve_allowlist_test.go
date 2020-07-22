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
package api

import (
	"github.com/goharbor/harbor/src/common/models"
	"net/http"
	"testing"
)

func TestSysCVEAllowlistAPIGet(t *testing.T) {
	url := "/api/system/CVEAllowlist"
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    url,
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        url,
				credential: nonSysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestSysCVEAllowlistAPIPut(t *testing.T) {
	url := "/api/system/CVEAllowlist"
	s := int64(1573254000)
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    url,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 400
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        url,
				bodyJSON:   []string{"CVE-1234-1234"},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    url,
				bodyJSON: models.CVEAllowlist{
					ExpiresAt: &s,
					Items: []models.CVEAllowlistItem{
						{CVEID: "CVE-2019-12310"},
					},
					ProjectID: 2,
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 400
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    url,
				bodyJSON: models.CVEAllowlist{
					ExpiresAt: &s,
					Items: []models.CVEAllowlistItem{
						{CVEID: "CVE-2019-12310"},
						{CVEID: "CVE-2019-12310"},
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		// 200
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    url,
				bodyJSON: models.CVEAllowlist{
					ExpiresAt: &s,
					Items: []models.CVEAllowlistItem{
						{CVEID: "CVE-2019-12310"},
						{CVEID: "RHSA-2019:2237"},
					},
				},
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
