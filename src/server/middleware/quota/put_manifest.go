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
	"strconv"

	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// PutManifestMiddleware middleware to request count and storage resources for the project
func PutManifestMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject:   projectReferenceObject,
		Resources:         putManifestResources,
		ResourcesExceeded: projectResourcesEvent(1),
		ResourcesWarning:  projectResourcesEvent(2),
	})
}

func putManifestResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)

	manifest, descriptor, err := unmarshalManifest(r)
	if err != nil {
		logger.Errorf("unmarshal manifest failed, error: %v", err)
		return nil, errors.Wrap(err, "unmarshal manifest failed").WithCode(errors.MANIFESTINVALID)
	}

	exist, err := blobController.Exist(r.Context(), descriptor.Digest.String(), blob.IsAssociatedWithProject(projectID))
	if err != nil {
		logger.Errorf("check manifest %s is associated with project failed, error: %v", descriptor.Digest.String(), err)
		return nil, err
	}

	if exist {
		return nil, nil
	}

	size := descriptor.Size

	var blobs []*models.Blob
	for _, reference := range manifest.References() {
		blobs = append(blobs, &models.Blob{
			Digest:      reference.Digest.String(),
			Size:        reference.Size,
			ContentType: reference.MediaType,
		})
	}

	missing, err := blobController.FindMissingAssociationsForProject(r.Context(), projectID, blobs)
	if err != nil {
		return nil, err
	}

	for _, m := range missing {
		if !m.IsForeignLayer() {
			size += m.Size
		}
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}
