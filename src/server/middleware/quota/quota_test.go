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
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/quota"
	pquota "github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	blobtesting "github.com/goharbor/harbor/src/testing/controller/blob"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	quotatesting "github.com/goharbor/harbor/src/testing/controller/quota"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type RequestMiddlewareTestSuite struct {
	suite.Suite

	originalArtifactController artifact.Controller
	artifactController         *artifacttesting.Controller

	originalBlobController blob.Controller
	blobController         *blobtesting.Controller

	originalProjectController project.Controller
	projectController         *projecttesting.Controller

	originallQuotaController quota.Controller
	quotaController          *quotatesting.Controller
}

func (suite *RequestMiddlewareTestSuite) SetupTest() {
	suite.originalArtifactController = artifactController
	suite.artifactController = &artifacttesting.Controller{}
	artifactController = suite.artifactController

	suite.originalBlobController = blobController
	suite.blobController = &blobtesting.Controller{}
	blobController = suite.blobController

	suite.originalProjectController = projectController
	suite.projectController = &projecttesting.Controller{}
	projectController = suite.projectController

	mock.OnAnything(suite.projectController, "GetByName").Return(&models.Project{ProjectID: 1, Name: "library"}, nil)

	suite.originallQuotaController = quotaController
	suite.quotaController = &quotatesting.Controller{}
	quotaController = suite.quotaController
}

func (suite *RequestMiddlewareTestSuite) TearDownTest() {
	artifactController = suite.originalArtifactController
	blobController = suite.originalBlobController
	projectController = suite.originalProjectController
	quotaController = suite.originallQuotaController
}

func (suite *RequestMiddlewareTestSuite) makeRequestConfig(reference, referenceID string, resources types.ResourceList) RequestConfig {
	return RequestConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return reference, referenceID, nil
		},
		Resources: func(*http.Request, string, string) (types.ResourceList, error) {
			return resources, nil
		},
	}
}

func (suite *RequestMiddlewareTestSuite) TestInvlidConfig() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	RequestMiddleware(RequestConfig{})(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestReferenceNotFound() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	config := RequestConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return "", "", fmt.Errorf("error")
		},
		Resources: func(*http.Request, string, string) (types.ResourceList, error) {
			return nil, nil
		},
	}

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestQuotaDisabled() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"
	config := suite.makeRequestConfig(reference, referenceID, nil)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(false, nil)

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestNoResourcesRequest() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"
	config := suite.makeRequestConfig(reference, referenceID, nil)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestResourcesRequestOK() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"
	resources := types.ResourceList{types.ResourceStorage: 100}
	config := suite.makeRequestConfig(reference, referenceID, resources)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.quotaController, "Request").Return(nil)

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestResourcesRequestFailed() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"
	resources := types.ResourceList{types.ResourceStorage: 100}
	config := suite.makeRequestConfig(reference, referenceID, resources)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.quotaController, "Request").Return(fmt.Errorf("error"))

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *RequestMiddlewareTestSuite) TestResourcesRequestDenied() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"
	resources := types.ResourceList{types.ResourceStorage: 100}
	config := suite.makeRequestConfig(reference, referenceID, resources)

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	var errs pquota.Errors
	errs = errs.Add(fmt.Errorf("Exceed"))
	mock.OnAnything(suite.quotaController, "Request").Return(errs)

	RequestMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusForbidden, rr.Code)
}

func TestRequestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &RequestMiddlewareTestSuite{})
}

type RefreshMiddlewareTestSuite struct {
	suite.Suite
	originallQuotaController quota.Controller
	quotaController          *quotatesting.Controller
}

func (suite *RefreshMiddlewareTestSuite) SetupTest() {
	suite.originallQuotaController = quotaController
	suite.quotaController = &quotatesting.Controller{}
	quotaController = suite.quotaController
}

func (suite *RefreshMiddlewareTestSuite) TestQuotaDisabled() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"

	config := RefreshConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return reference, referenceID, nil
		},
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(false, nil)

	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RefreshMiddlewareTestSuite) TestQuotaIsEnabledFailed() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"

	config := RefreshConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return reference, referenceID, nil
		},
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(false, fmt.Errorf("error"))

	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *RefreshMiddlewareTestSuite) TestInvalidConfig() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	config := RefreshConfig{}
	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func (suite *RefreshMiddlewareTestSuite) TestNotSuccess() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	config := RefreshConfig{}
	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *RefreshMiddlewareTestSuite) TestRefershOK() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"

	config := RefreshConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return reference, referenceID, nil
		},
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.quotaController, "Refresh").Return(nil)

	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RefreshMiddlewareTestSuite) TestRefershFailed() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	rr := httptest.NewRecorder()

	reference, referenceID := "project", "1"

	config := RefreshConfig{
		ReferenceObject: func(*http.Request) (string, string, error) {
			return reference, referenceID, nil
		},
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)
	mock.OnAnything(suite.quotaController, "Refresh").Return(fmt.Errorf("error"))

	RefreshMiddleware(config)(next).ServeHTTP(rr, req)
	suite.Equal(http.StatusInternalServerError, rr.Code)
}

func TestRefreshMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &RefreshMiddlewareTestSuite{})
}
