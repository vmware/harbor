package blob

import (
	"fmt"
	beego_orm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/orm"
	pkg_blob "github.com/goharbor/harbor/src/pkg/blob"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HeadBlobUploadMiddlewareTestSuite struct {
	htesting.Suite
}

func (suite *HeadBlobUploadMiddlewareTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"blob"}
}

func (suite *HeadBlobUploadMiddlewareTestSuite) makeRequest(projectName, digest string) *http.Request {
	req := httptest.NewRequest("HEAD", fmt.Sprintf("/v2/%s/blobs/%s", projectName, digest), nil)
	info := lib.ArtifactInfo{
		Repository: fmt.Sprintf("%s/photon", projectName),
		Reference:  "2.0",
		Tag:        "2.0",
		Digest:     digest,
	}
	*req = *(req.WithContext(orm.NewContext(req.Context(), beego_orm.NewOrm())))
	*req = *(req.WithContext(lib.WithArtifactInfo(req.Context(), info)))
	return req
}

func (suite *HeadBlobUploadMiddlewareTestSuite) TestHeadBlobStatusNone() {
	suite.WithProject(func(projectID int64, projectName string) {
		digest := suite.DigestString()

		_, err := blob.Ctl.Ensure(suite.Context(), digest, "application/octet-stream", 512)
		suite.Nil(err)

		req := suite.makeRequest(projectName, digest)
		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusOK, map[string]string{"Docker-Content-Digest": digest})
		HeadBlobMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusOK, res.Code)

		blob, err := blob.Ctl.Get(suite.Context(), digest)
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(blob_models.StatusNone, blob.Status)
	})
}

func (suite *HeadBlobUploadMiddlewareTestSuite) TestHeadBlobStatusDeleting() {
	suite.WithProject(func(projectID int64, projectName string) {
		digest := suite.DigestString()

		id, err := blob.Ctl.Ensure(suite.Context(), digest, "application/octet-stream", 512)
		suite.Nil(err)

		// status-none -> status-delete -> status-deleting
		_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), &blob_models.Blob{ID: id, Status: blob_models.StatusDelete})
		suite.Nil(err)
		_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), &blob_models.Blob{ID: id, Status: blob_models.StatusDeleting, Version: 1})
		suite.Nil(err)

		req := suite.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/blobs/%s", projectName, digest), nil)
		res := httptest.NewRecorder()

		next := suite.NextHandler(http.StatusOK, map[string]string{"Docker-Content-Digest": digest})
		HeadBlobMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusNotFound, res.Code)

		blob, err := blob.Ctl.Get(suite.Context(), digest)
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(blob_models.StatusDeleting, blob.Status)
	})
}

func TestHeadBlobUploadMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &HeadBlobUploadMiddlewareTestSuite{})
}
