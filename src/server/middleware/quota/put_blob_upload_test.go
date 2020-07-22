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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type PutBlobUploadMiddlewareTestSuite struct {
	RequestMiddlewareTestSuite

	handler http.Handler
}

func (suite *PutBlobUploadMiddlewareTestSuite) SetupTest() {
	suite.RequestMiddlewareTestSuite.SetupTest()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	suite.handler = PutBlobUploadMiddleware()(next)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
}

func (suite *PutBlobUploadMiddlewareTestSuite) makeRequest(contentLength int) *http.Request {
	url := "/v2/library/photon/blobs/uploads/cbabe458-28a1-4e1b-ad15-0cb0229df4e8?digest=sha256:57c2ec3bf82f09c94be2e5c5beb124b86fcbb42e76fb82c99066c054422010e4"
	req := httptest.NewRequest(http.MethodPut, url, nil)
	req.Header.Set("Content-Length", strconv.Itoa(contentLength))

	return req
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestBlobSizeIsZero() {
	mock.OnAnything(suite.blobController, "GetAcceptedBlobSize").Return(int64(0), nil)

	req := suite.makeRequest(0)
	rr := httptest.NewRecorder()

	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestGetAcceptedBlobSizeFailed() {
	mock.OnAnything(suite.blobController, "GetAcceptedBlobSize").Return(int64(0), fmt.Errorf("error"))

	req := suite.makeRequest(0)
	rr := httptest.NewRecorder()

	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestBlobExist() {
	mock.OnAnything(suite.blobController, "Exist").Return(true, nil)

	req := suite.makeRequest(100)
	rr := httptest.NewRecorder()

	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestBlobNotExist() {
	mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()
	mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
		resources := args.Get(3).(types.ResourceList)
		suite.Len(resources, 1)
		suite.Equal(resources[types.ResourceStorage], int64(100))

		f := args.Get(4).(func() error)
		f()
	})
	mock.OnAnything(suite.quotaController, "GetByRef").Return(&quota.Quota{}, nil).Once()

	req := suite.makeRequest(100)
	rr := httptest.NewRecorder()

	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestBlobExistFailed() {
	mock.OnAnything(suite.blobController, "Exist").Return(false, fmt.Errorf("error"))

	req := suite.makeRequest(100)
	rr := httptest.NewRecorder()

	suite.handler.ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestResourcesExceeded() {
	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.blobController, "Exist").Return(false, nil)
	mock.OnAnything(suite.projectController, "Get").Return(&commonmodels.Project{}, nil)

	{
		var errs quota.Errors
		errs = errs.Add(quota.NewResourceOverflowError(types.ResourceStorage, 100, 100, 110))
		mock.OnAnything(suite.quotaController, "Request").Return(errs).Once()

		req := suite.makeRequest(100)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		suite.handler.ServeHTTP(rr, req)
		suite.NotEqual(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}

	{
		var errs quota.Errors
		errs = errs.Add(quota.NewResourceOverflowError(types.ResourceStorage, 100, 100, 110))

		err := errors.DeniedError(errs).WithMessage("Quota exceeded when processing the request of %v", errs)
		mock.OnAnything(suite.quotaController, "Request").Return(err).Once()

		req := suite.makeRequest(100)
		eveCtx := notification.NewEventCtx()
		req = req.WithContext(notification.NewContext(req.Context(), eveCtx))
		rr := httptest.NewRecorder()

		suite.handler.ServeHTTP(rr, req)
		suite.NotEqual(http.StatusOK, rr.Code)
		suite.Equal(1, eveCtx.Events.Len())
	}
}

func TestPutBlobUploadMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PutBlobUploadMiddlewareTestSuite{})
}
