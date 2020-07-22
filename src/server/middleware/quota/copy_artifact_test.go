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

package quota

import (
	"net/http"
	"net/http/httptest"
	"testing"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

func Test_parseRepositoryName(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"/api/v2.0/projects", args{"/api/v2.0/projects"}, ""},
		{"/api/v2.0/projects/library/repositories/photon/artifacts", args{"/api/v2.0/projects/library/repositories/photon/artifacts"}, "photon"},
		{"/api/v2.0/projects/library/repositories/photon/artifacts/", args{"/api/v2.0/projects/library/repositories/photon/artifacts/"}, "photon"},
		{"/api/v2.0/projects/library/repositories/amd64/photon/artifacts", args{"/api/v2.0/projects/library/repositories/amd64/photon/artifacts"}, "amd64/photon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRepositoryName(tt.args.p); got != tt.want {
				t.Errorf("parseRepositoryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

type CopyArtifactMiddlewareTestSuite struct {
	RequestMiddlewareTestSuite

	artifact *artifact.Artifact
}

func (suite *CopyArtifactMiddlewareTestSuite) SetupTest() {
	suite.RequestMiddlewareTestSuite.SetupTest()

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)

	suite.artifact = &artifact.Artifact{}

	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.artifactController, "Walk").Return(nil).Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(suite.artifact)
	})

	mock.OnAnything(suite.projectController, "Get").Return(&commonmodels.Project{}, nil)
}

func (suite *CopyArtifactMiddlewareTestSuite) TestResourcesWarning() {
	mock.OnAnything(suite.blobController, "List").Return(nil, nil)
	mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(nil, nil)
	mock.OnAnything(suite.quotaController, "Request").Return(nil).Run(func(args mock.Arguments) {
		f := args.Get(4).(func() error)
		f()
	})

	mock.OnAnything(suite.artifactController, "Count").Return(int64(0), nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	{
		q := &quota.Quota{}
		q.SetHard(types.ResourceList{types.ResourceStorage: 100})
		q.SetUsed(types.ResourceList{types.ResourceStorage: 50})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(q, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0?from=library/photon:2.0.1", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		CopyArtifactMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
		suite.Equal(0, eveCtx.Events.Len())
	}

	{
		q := &quota.Quota{}
		q.SetHard(types.ResourceList{types.ResourceStorage: 100})
		q.SetUsed(types.ResourceList{types.ResourceStorage: 85})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(q, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0?from=library/photon:2.0.1", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		CopyArtifactMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}
}

func TestCopyArtifactMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &CopyArtifactMiddlewareTestSuite{})
}
