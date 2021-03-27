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
	"context"
	std_err "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/distribution/distribution/manifest/schema2"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/testing/mock"
	distributiontesting "github.com/goharbor/harbor/src/testing/pkg/distribution"
	"github.com/stretchr/testify/suite"
)

type PutManifestMiddlewareTestSuite struct {
	RequestMiddlewareTestSuite

	unmarshalManifest func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error)
	manifest          distribution.Manifest
}

func (suite *PutManifestMiddlewareTestSuite) SetupTest() {
	suite.RequestMiddlewareTestSuite.SetupTest()

	suite.unmarshalManifest = unmarshalManifest
	suite.manifest = &distributiontesting.Manifest{}

	mock.OnAnything(suite.manifest, "References").Return([]distribution.Descriptor{
		{Digest: "blob1", Size: 10, MediaType: schema2.MediaTypeLayer},
		{Digest: "blob2", Size: 20, MediaType: schema2.MediaTypeLayer},
		{Digest: "blob3", Size: 30, MediaType: schema2.MediaTypeForeignLayer},
		{Digest: "blob4", Size: 40, MediaType: schema2.MediaTypeForeignLayer},
	})

	unmarshalManifest = func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error) {
		return suite.manifest, distribution.Descriptor{Digest: "digest", Size: 100}, nil
	}
}

func (suite *PutManifestMiddlewareTestSuite) TearDownTest() {
	suite.RequestMiddlewareTestSuite.TearDownTest()

	unmarshalManifest = suite.unmarshalManifest
}

func (suite *PutManifestMiddlewareTestSuite) TestMiddleware() {
	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	{
		mock.OnAnything(suite.blobController, "Exist").Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and blobs are already associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()
		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(nil, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 1)
			suite.Equal(resources[types.ResourceStorage], int64(100))

			f := args.Get(4).(func() error)
			f()
		})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(&quota.Quota{}, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and some blobs are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[:1]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 1)
			suite.Equal(resources[types.ResourceStorage], int64(100+10))

			f := args.Get(4).(func() error)
			f()
		})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(&quota.Quota{}, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and some blobs include foreign layers are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[1:]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 1)
			suite.Equal(resources[types.ResourceStorage], int64(100+20))

			f := args.Get(4).(func() error)
			f()
		})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(&quota.Quota{}, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and only foreign layers are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[2:]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 1)
			suite.Equal(resources[types.ResourceStorage], int64(100))

			f := args.Get(4).(func() error)
			f()
		})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(&quota.Quota{}, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}
}

func (suite *PutManifestMiddlewareTestSuite) TestResourcesExceeded() {
	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.blobController, "Exist").Return(false, nil)
	mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(nil, nil)
	mock.OnAnything(suite.projectController, "Get").Return(&commonmodels.Project{}, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	{
		var errs quota.Errors
		errs = errs.Add(quota.NewResourceOverflowError(types.ResourceStorage, 100, 100, 110))
		mock.OnAnything(suite.quotaController, "Request").Return(errs).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.NotEqual(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}

	{
		var errs quota.Errors
		errs = errs.Add(quota.NewResourceOverflowError(types.ResourceStorage, 100, 100, 110))

		err := errors.DeniedError(errs).WithMessage("Quota exceeded when processing the request of %v", errs)
		mock.OnAnything(suite.quotaController, "Request").Return(err).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.NotEqual(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}
}

func (suite *PutManifestMiddlewareTestSuite) TestResourcesWarning() {
	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.blobController, "Exist").Return(false, nil)
	mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(nil, nil)

	mock.OnAnything(suite.quotaController, "Request").Return(nil).Run(func(args mock.Arguments) {
		f := args.Get(4).(func() error)
		f()
	})
	mock.OnAnything(suite.projectController, "Get").Return(&commonmodels.Project{}, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	{
		q := &quota.Quota{}
		q.SetHard(types.ResourceList{types.ResourceStorage: 100})
		q.SetUsed(types.ResourceList{types.ResourceStorage: 50})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(q, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
		suite.Equal(0, eveCtx.Events.Len())
	}

	{
		q := &quota.Quota{}
		q.SetHard(types.ResourceList{types.ResourceStorage: 100})
		q.SetUsed(types.ResourceList{types.ResourceStorage: 85})
		mock.OnAnything(suite.quotaController, "GetByRef").Return(q, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}
}

func (suite *PutManifestMiddlewareTestSuite) TestPutInvalid() {
	unmarshalManifest = func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error) {
		return nil, distribution.Descriptor{}, std_err.New("json: cannot unmarshal string into Go value of type map")
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/sha256:totallywrong", strings.NewReader("YmxhYmxhYmxh"))
	rr := httptest.NewRecorder()

	PutManifestMiddleware()(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)

}

func TestPutManifestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PutManifestMiddlewareTestSuite{})
}
