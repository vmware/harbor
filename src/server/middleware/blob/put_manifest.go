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
	"github.com/goharbor/harbor/src/lib/errors"
	"io/ioutil"
	"net/http"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

// PutManifestMiddleware middleware middleware is to update the manifest status according to the different situation before the request passed into proxy(distribution).
// and it creates Blobs for the foreign layers and associate them with the project, updates the content type of the Blobs which already exist,
// create Blob for the manifest, associate all Blobs with the manifest after PUT /v2/<name>/manifests/<reference> success.
func PutManifestMiddleware() func(http.Handler) http.Handler {
	before := middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()
		logger := log.G(ctx)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		_, descriptor, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return errors.Wrapf(err, "unmarshal manifest failed").WithCode(errors.MANIFESTINVALID)
		}

		return probeBlob(r, descriptor.Digest.String())
	})

	after := middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "blob"})

		p, err := projectController.GetByName(ctx, distribution.ParseProjectName(r.URL.Path))
		if err != nil {
			logger.Errorf("get project failed, error: %v", err)
			return err
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, descriptor, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		// sync blobs
		if err := blobController.Sync(ctx, manifest.References()); err != nil {
			logger.Errorf("sync missing blobs from manifest %s failed, error: %c", descriptor.Digest.String(), err)
			return err
		}

		// NOTE: associate all blobs with project because the already exist associations may cleanup by others
		for _, reference := range manifest.References() {
			if err := blobController.AssociateWithProjectByDigest(ctx, reference.Digest.String(), p.ProjectID); err != nil {
				return err
			}
		}

		// ensure Blob for the manifest
		blobID, err := blobController.Ensure(ctx, descriptor.Digest.String(), contentType, descriptor.Size)
		if err != nil {
			logger.Errorf("ensure blob %s failed, error: %v", descriptor.Digest.String(), err)
			return err
		}

		if err := blobController.AssociateWithProjectByID(ctx, blobID, p.ProjectID); err != nil {
			logger.Errorf("associate manifest with artifact %s failed, error: %v", descriptor.Digest.String(), err)
			return err
		}

		var blobDigests []string
		for _, reference := range manifest.References() {
			blobDigests = append(blobDigests, reference.Digest.String())
		}

		// associate blobs of the manifest with artifact
		if err := blobController.AssociateWithArtifact(ctx, blobDigests, descriptor.Digest.String()); err != nil {
			logger.Errorf("associate blobs with artifact %s failed, error: %v", descriptor.Digest.String(), err)
			return err
		}

		return nil
	})

	return middleware.Chain(before, after)
}
