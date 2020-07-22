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
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// PutBlobUploadMiddleware middleware to request storage resource for the project
func PutBlobUploadMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject:   projectReferenceObject,
		Resources:         putBlobUploadResources,
		ResourcesExceeded: projectResourcesEvent(1),
		ResourcesWarning:  projectResourcesEvent(2),
	})
}

func putBlobUploadResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil || size == 0 {
		size, err = blobController.GetAcceptedBlobSize(distribution.ParseSessionID(r.URL.Path))
	}
	if err != nil {
		logger.Errorf("get blob size failed, error: %v", err)
		return nil, err
	}

	if size == 0 {
		logger.Debug("blob size is 0")
		return nil, nil
	}

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)

	digest := r.URL.Query().Get("digest")
	exist, err := blobController.Exist(r.Context(), digest, blob.IsAssociatedWithProject(projectID))
	if err != nil {
		logger.Errorf("checking blob %s is associated with project %d failed, error: %v", digest, projectID, err)
		return nil, err
	}

	if exist {
		return nil, nil
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}
