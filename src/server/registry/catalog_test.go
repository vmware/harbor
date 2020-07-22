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

package registry

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	pkg_art "github.com/goharbor/harbor/src/pkg/artifact"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	repotesting "github.com/goharbor/harbor/src/testing/controller/repository"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type catalogTestSuite struct {
	suite.Suite
	originalRepoCtl repository.Controller
	originalArtCtl  artifact.Controller
	repoCtl         *repotesting.FakeController
	artCtl          *artifacttesting.Controller
}

func (c *catalogTestSuite) SetupSuite() {
	c.originalRepoCtl = repository.Ctl
	c.originalArtCtl = artifact.Ctl
}

func (c *catalogTestSuite) SetupTest() {
	c.repoCtl = &repotesting.FakeController{}
	repository.Ctl = c.repoCtl
	c.artCtl = &artifacttesting.Controller{}
	artifact.Ctl = c.artCtl
}

func (c *catalogTestSuite) TearDownTest() {
}

func (c *catalogTestSuite) TearDownSuite() {
	repository.Ctl = c.originalRepoCtl
	artifact.Ctl = c.originalArtCtl
}

func (c *catalogTestSuite) TestCatalog() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkg_art.Artifact{
				ProjectID:    1,
				RepositoryID: 1,
			},
			Tags: []*tag.Tag{
				{
					Tag: model_tag.Tag{
						RepositoryID: 1,
						ArtifactID:   1,
					},
				},
			},
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(2, len(ctlg.Repositories))
}

func (c *catalogTestSuite) TestCatalogPaginationN1() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=1", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkg_art.Artifact{
				ProjectID:    1,
				RepositoryID: 1,
			},
			Tags: []*tag.Tag{
				{
					Tag: model_tag.Tag{
						RepositoryID: 1,
						ArtifactID:   1,
					},
				},
			},
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(1, len(ctlg.Repositories))
	c.Equal("busybox", ctlg.Repositories[0])
}

func (c *catalogTestSuite) TestCatalogPaginationN2() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=3", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkg_art.Artifact{
				ProjectID:    1,
				RepositoryID: 1,
			},
			Tags: []*tag.Tag{
				{
					Tag: model_tag.Tag{
						RepositoryID: 1,
						ArtifactID:   1,
					},
				},
			},
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(2, len(ctlg.Repositories))
	c.Equal("hello-world", ctlg.Repositories[1])
}

func (c *catalogTestSuite) TestCatalogPaginationN3() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?last=busybox&n=1", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkg_art.Artifact{
				ProjectID:    1,
				RepositoryID: 1,
			},
			Tags: []*tag.Tag{
				{
					Tag: model_tag.Tag{
						RepositoryID: 1,
						ArtifactID:   1,
					},
				},
			},
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(1, len(ctlg.Repositories))
	c.Equal("hello-world", ctlg.Repositories[0])
}

func (c *catalogTestSuite) TestCatalogUntaggedArtifact() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	// untagged artifact
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkg_art.Artifact{
				ProjectID:    1,
				RepositoryID: 1,
			},
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(0, len(ctlg.Repositories))
}

func (c *catalogTestSuite) TestCatalogEmptyRepo() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	// empty repository
	mock.OnAnything(c.artCtl, "List").Return([]*artifact.Artifact{
		{},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(0, len(ctlg.Repositories))
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, &catalogTestSuite{})
}
