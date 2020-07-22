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
	"net/http"

	"github.com/goharbor/harbor/src/server/middleware/artifactinfo"
	"github.com/goharbor/harbor/src/server/middleware/blob"
	"github.com/goharbor/harbor/src/server/middleware/contenttrust"
	"github.com/goharbor/harbor/src/server/middleware/immutable"
	"github.com/goharbor/harbor/src/server/middleware/quota"
	"github.com/goharbor/harbor/src/server/middleware/repoproxy"
	"github.com/goharbor/harbor/src/server/middleware/v2auth"
	"github.com/goharbor/harbor/src/server/middleware/vulnerable"
	"github.com/goharbor/harbor/src/server/router"
)

// RegisterRoutes for OCI registry APIs
func RegisterRoutes() {
	root := router.NewRoute().
		Path("/v2").
		Middleware(artifactinfo.Middleware()).
		Middleware(v2auth.Middleware())
	// catalog
	root.NewRoute().
		Method(http.MethodGet).
		Path("/_catalog").
		Handler(newRepositoryHandler())
	// list tags
	root.NewRoute().
		Method(http.MethodGet).
		Path("/*/tags/list").
		Handler(newTagHandler())
	// manifest
	root.NewRoute().
		Method(http.MethodGet).
		Path("/*/manifests/:reference").
		Middleware(repoproxy.ManifestGetMiddleware()).
		Middleware(contenttrust.Middleware()).
		Middleware(vulnerable.Middleware()).
		HandlerFunc(getManifest)
	root.NewRoute().
		Method(http.MethodHead).
		Path("/*/manifests/:reference").
		HandlerFunc(getManifest)
	root.NewRoute().
		Method(http.MethodDelete).
		Path("/*/manifests/:reference").
		Middleware(quota.RefreshForProjectMiddleware()).
		HandlerFunc(deleteManifest)
	root.NewRoute().
		Method(http.MethodPut).
		Path("/*/manifests/:reference").
		Middleware(repoproxy.DisableBlobAndManifestUploadMiddleware()).
		Middleware(immutable.Middleware()).
		Middleware(quota.PutManifestMiddleware()).
		Middleware(blob.PutManifestMiddleware()).
		HandlerFunc(putManifest)
	// blob get
	root.NewRoute().
		Method(http.MethodGet).
		Path("/*/blobs/:digest").
		Middleware(repoproxy.BlobGetMiddleware()).
		Handler(proxy)
	// initiate blob upload
	root.NewRoute().
		Method(http.MethodPost).
		Path("/*/blobs/uploads").
		Middleware(quota.PostInitiateBlobUploadMiddleware()).
		Middleware(blob.PostInitiateBlobUploadMiddleware()).
		Handler(proxy)
	// blob upload
	root.NewRoute().
		Method(http.MethodPatch).
		Path("/*/blobs/uploads/:session_id").
		Middleware(repoproxy.DisableBlobAndManifestUploadMiddleware()).
		Middleware(blob.PatchBlobUploadMiddleware()).
		Handler(proxy)
	root.NewRoute().
		Method(http.MethodPut).
		Path("/*/blobs/uploads/:session_id").
		Middleware(repoproxy.DisableBlobAndManifestUploadMiddleware()).
		Middleware(quota.PutBlobUploadMiddleware()).
		Middleware(blob.PutBlobUploadMiddleware()).
		Handler(proxy)
	root.NewRoute().
		Method(http.MethodHead).
		Path("/*/blobs/:digest").
		Middleware(blob.HeadBlobMiddleware()).
		Handler(proxy)
	// others
	root.NewRoute().Path("/*").Handler(proxy)
}
