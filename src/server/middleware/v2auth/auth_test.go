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

package v2auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
)

type mockPM struct{}

func (mockPM) Get(projectIDOrName interface{}) (*models.Project, error) {
	name := projectIDOrName.(string)
	id, _ := strconv.Atoi(strings.TrimPrefix(name, "project_"))
	if id == 0 {
		return nil, nil
	}
	return &models.Project{
		ProjectID: int64(id),
		Name:      name,
	}, nil
}

func (mockPM) Create(*models.Project) (int64, error) {
	panic("implement me")
}

func (mockPM) Delete(projectIDOrName interface{}) error {
	panic("implement me")
}

func (mockPM) Update(projectIDOrName interface{}, project *models.Project) error {
	panic("implement me")
}

func (mockPM) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	panic("implement me")
}

func (mockPM) IsPublic(projectIDOrName interface{}) (bool, error) {
	return false, nil
}

func (mockPM) Exists(projectIDOrName interface{}) (bool, error) {
	panic("implement me")
}

func (mockPM) GetPublic() ([]*models.Project, error) {
	panic("implement me")
}

func (mockPM) GetAuthorized(user *models.User) ([]*models.Project, error) {
	return nil, nil
}

func (mockPM) GetMetadataManager() metamgr.ProjectMetadataManager {
	panic("implement me")
}

func TestMain(m *testing.M) {
	checker = reqChecker{
		pm: mockPM{},
	}
	conf := map[string]interface{}{
		common.ExtEndpoint: "https://harbor.test",
		common.CoreURL:     "https://harbor.core:8443",
	}
	config.InitWithSettings(conf)
	if rc := m.Run(); rc != 0 {
		os.Exit(rc)
	}
}

func TestMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)
	mock.OnAnything(sc, "Can").Return(func(action types.Action, resource types.Resource) bool {
		perms := map[string]map[rbac.Action]struct{}{
			"/project/1/repository": {
				rbac.ActionPull: {},
				rbac.ActionPush: {},
			},
			"/project/2/repository": {
				rbac.ActionPull: {},
			},
		}
		m, ok := perms[resource.String()]
		if !ok {
			return false
		}
		_, ok = m[action]
		return ok
	})

	baseCtx := security.NewContext(context.Background(), sc)
	ar1 := lib.ArtifactInfo{
		Repository:  "project_1/hello-world",
		Reference:   "v1",
		ProjectName: "project_1",
	}
	ar2 := lib.ArtifactInfo{
		Repository:  "library/ubuntu",
		Reference:   "14.04",
		ProjectName: "library",
	}
	ar3 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_2/ubuntu",
		BlobMountProjectName: "project_2",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}
	ar4 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_3/ubuntu",
		BlobMountProjectName: "project_3",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}
	ar5 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_0/ubuntu",
		BlobMountProjectName: "project_0",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}

	ctx1 := lib.WithArtifactInfo(baseCtx, ar1)
	ctx2 := lib.WithArtifactInfo(baseCtx, ar2)
	ctx3 := lib.WithArtifactInfo(baseCtx, ar3)
	ctx4 := lib.WithArtifactInfo(baseCtx, ar4)
	ctx5 := lib.WithArtifactInfo(baseCtx, ar5)
	req1a, _ := http.NewRequest(http.MethodGet, "/v2/project_1/hello-world/manifest/v1", nil)
	req1b, _ := http.NewRequest(http.MethodDelete, "/v2/project_1/hello-world/manifest/v1", nil)
	req1c, _ := http.NewRequest(http.MethodHead, "/v2/project_1/hello-world/manifest/v1", nil)
	req2, _ := http.NewRequest(http.MethodGet, "/v2/library/ubuntu/manifest/14.04", nil)
	req3, _ := http.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	req4, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_2/ubuntu", nil)
	req5, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_3/ubuntu", nil)
	req6, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_0/ubuntu", nil)

	cases := []struct {
		input  *http.Request
		status int
	}{
		{
			input:  req1a.WithContext(ctx1),
			status: http.StatusOK,
		},
		{
			input:  req1b.WithContext(ctx1),
			status: http.StatusUnauthorized,
		},
		{
			input:  req1c.WithContext(ctx1),
			status: http.StatusUnauthorized,
		},
		{
			input:  req2.WithContext(ctx2),
			status: http.StatusUnauthorized,
		},
		{
			input:  req3.WithContext(baseCtx),
			status: http.StatusUnauthorized,
		},
		{
			input:  req4.WithContext(ctx3),
			status: http.StatusOK,
		},
		{
			input:  req5.WithContext(ctx4),
			status: http.StatusUnauthorized,
		},
		{
			input:  req6.WithContext(ctx5),
			status: http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		t.Logf("req : %s, %s", c.input.Method, c.input.URL)
		Middleware()(next).ServeHTTP(rec, c.input)
		assert.Equal(t, c.status, rec.Result().StatusCode)
	}
}

func TestGetChallenge(t *testing.T) {
	req1, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/", nil)
	req1x := req1.Clone(req1.Context())
	req1x.SetBasicAuth("u", "p")
	req2, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/_catalog", nil)
	req2x := req2.Clone(req2.Context())
	req2x.Header.Set("Authorization", "Bearer xx")
	req3, _ := http.NewRequest(http.MethodPost, "https://registry.test/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_2/ubuntu", nil)
	req3 = req3.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_2/ubuntu",
		BlobMountProjectName: "project_2",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}))
	req3x := req3.Clone(req3.Context())
	req3x.SetBasicAuth("", "")
	req3x.Host = "harbor.test"
	req4, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/project_1/hello-world/manifests/v1", nil)
	req4 = req4.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
		Repository:  "project_1/hello-world",
		Reference:   "v1",
		ProjectName: "project_1",
	}))
	req4.Host = "harbor.core:8443"

	cases := []struct {
		request   *http.Request
		challenge string
	}{
		{
			request:   req1,
			challenge: `Bearer realm="https://harbor.test/service/token",service="harbor-registry"`,
		},
		{
			request:   req1x,
			challenge: `Basic realm="harbor"`,
		},
		{
			request:   req2,
			challenge: `Bearer realm="https://harbor.test/service/token",service="harbor-registry"`,
		},
		{
			request:   req2x,
			challenge: `Basic realm="harbor"`,
		},
		{
			request:   req3,
			challenge: `Bearer realm="https://harbor.test/service/token",service="harbor-registry",scope="repository:project_1/ubuntu:pull,push repository:project_2/ubuntu:pull"`,
		},
		{
			request:   req3x,
			challenge: `Basic realm="harbor"`,
		},
		{
			request:   req4,
			challenge: `Bearer realm="https://harbor.core:8443/service/token",service="harbor-registry",scope="repository:project_1/hello-world:pull"`,
		},
	}
	for _, c := range cases {
		acs := accessList(c.request)
		assert.Equal(t, c.challenge, getChallenge(c.request, acs))
	}

}
