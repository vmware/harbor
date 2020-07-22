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

package blob

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	pkg_blob "github.com/goharbor/harbor/src/pkg/blob"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/pkg/distribution"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PutManifestMiddlewareTestSuite struct {
	htesting.Suite
}

func (suite *PutManifestMiddlewareTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"project_blob", "blob", "artifact_blob"}
}

func (suite *PutManifestMiddlewareTestSuite) pushBlob(name string, digest string, size int64) {
	req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, uuid.New().String()), nil)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	res := httptest.NewRecorder()

	next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": digest})
	PutBlobUploadMiddleware()(next).ServeHTTP(res, req)
	suite.Equal(res.Code, http.StatusCreated)
}

func (suite *PutManifestMiddlewareTestSuite) prepare(name string) (distribution.Manifest, distribution.Descriptor, *http.Request) {
	body := fmt.Sprintf(`
	{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
		"mediaType": "application/vnd.docker.container.image.v1+json",
		"size": 6868,
		"digest": "%s"
		},
		"layers": [
		{
			"mediaType": "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
			"size": 27092274,
			"digest": "sha256:8ec398bc03560e0fa56440e96da307cdf0b1ad153f459b52bca53ae7ddb8236d"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 1730,
			"digest": "sha256:da01136793fac089b2ff13c2bf3c9d5d5550420fbd9981e08198fd251a0ab7b4"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 1357602,
			"digest": "sha256:cf1486a2c0b86ddb45238e86c6bf9666c20113f7878e4cd4fa175fd74ac5d5b7"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 7344202,
			"digest": "sha256:a44f7da98d9e65b723ee913a9e6758db120a43fcce564b3dcf61cb9eb2823dad"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 97,
			"digest": "sha256:c677fde73875fc4c1e38ccdc791fe06380be0468fac220358f38c910e336266e"
		},
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 409,
			"digest": "sha256:727f8da63ac248054cb7dda635ee16da76e553ec99be565a54180c83d04025a8"
		}
		]
	}`, suite.DigestString())

	manifest, descriptor, err := distribution.UnmarshalManifest("application/vnd.docker.distribution.manifest.v2+json", []byte(body))
	suite.Nil(err)

	req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", name, descriptor.Digest.String()), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	info := lib.ArtifactInfo{
		Repository: name,
		Reference:  "latest",
		Tag:        "latest",
		Digest:     descriptor.Digest.String(),
	}
	return manifest, descriptor, req.WithContext(lib.WithArtifactInfo(req.Context(), info))
}

func (suite *PutManifestMiddlewareTestSuite) TestMiddleware() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/redis", projectName)

		manifest, descriptor, req := suite.prepare(name)
		res := httptest.NewRecorder()

		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		PutManifestMiddleware()(next).ServeHTTP(res, req)

		suite.Equal(http.StatusCreated, res.Code)

		for _, reference := range manifest.References() {
			opts := []blob.Option{
				blob.IsAssociatedWithArtifact(descriptor.Digest.String()),
				blob.IsAssociatedWithProject(projectID),
			}

			b, err := blob.Ctl.Get(suite.Context(), reference.Digest.String(), opts...)
			if suite.Nil(err) {
				suite.Equal(reference.MediaType, b.ContentType)
				suite.Equal(reference.Size, b.Size)
			}
		}

		{
			opts := []blob.Option{
				blob.IsAssociatedWithArtifact(descriptor.Digest.String()),
				blob.IsAssociatedWithProject(projectID),
			}
			b, err := blob.Ctl.Get(suite.Context(), descriptor.Digest.String(), opts...)
			if suite.Nil(err) {
				suite.Equal(descriptor.MediaType, b.ContentType)
				suite.Equal(descriptor.Size, b.Size)
			}
		}
	})
}

func (suite *PutManifestMiddlewareTestSuite) TestMFInDeleting() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/photon", projectName)
		_, descriptor, req := suite.prepare(name)
		res := httptest.NewRecorder()

		id, err := blob.Ctl.Ensure(suite.Context(), descriptor.Digest.String(), "application/vnd.docker.distribution.manifest.v2+json", 512)
		suite.Nil(err)

		// status-none -> status-delete -> status-deleting
		_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), &blob_models.Blob{ID: id, Status: blob_models.StatusDelete})
		suite.Nil(err)
		_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), &blob_models.Blob{ID: id, Status: blob_models.StatusDeleting, Version: 1})
		suite.Nil(err)

		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		PutManifestMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusNotFound, res.Code)
	})
}

func (suite *PutManifestMiddlewareTestSuite) TestMFInDelete() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/photon", projectName)
		manifest, descriptor, req := suite.prepare(name)
		res := httptest.NewRecorder()

		id, err := blob.Ctl.Ensure(suite.Context(), descriptor.Digest.String(), "application/vnd.docker.distribution.manifest.v2+json", 512)
		suite.Nil(err)

		// status-none -> status-delete -> status-deleting
		_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), &blob_models.Blob{ID: id, Status: blob_models.StatusDelete})
		suite.Nil(err)

		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		PutManifestMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		for _, reference := range manifest.References() {
			opts := []blob.Option{
				blob.IsAssociatedWithArtifact(descriptor.Digest.String()),
				blob.IsAssociatedWithProject(projectID),
			}

			b, err := blob.Ctl.Get(suite.Context(), reference.Digest.String(), opts...)
			if suite.Nil(err) {
				suite.Equal(reference.MediaType, b.ContentType)
				suite.Equal(reference.Size, b.Size)
				suite.Equal(blob_models.StatusNone, b.Status)
			}
		}
	})
}

func TestPutManifestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PutManifestMiddlewareTestSuite{})
}
