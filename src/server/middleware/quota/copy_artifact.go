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
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// CopyArtifactMiddleware middleware to request count and storage resources for copy artifact API
func CopyArtifactMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject:   projectReferenceObject,
		Resources:         copyArtifactResources,
		ResourcesExceeded: copyArtifactResourcesEvent(1),
		ResourcesWarning:  copyArtifactResourcesEvent(2),
	})
}

func parseRepositoryName(p string) string {
	parts := strings.Split(strings.TrimSuffix(path.Clean(p), "/"), "/repositories/")
	if len(parts) != 2 {
		return ""
	}

	return strings.TrimSuffix(parts[1], "/artifacts")
}

func copyArtifactResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	query := r.URL.Query()
	from := query.Get("from")
	if from == "" {
		// miss the from parameter, skip to request the resources
		return nil, nil
	}

	logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

	repository, reference, err := distribution.ParseRef(from)
	if err != nil {
		// bad from parameter, skip to request the resources
		logger.Errorf("parse from parameter failed, error: %v", err)
		return nil, nil
	}

	ctx := r.Context()

	art, err := artifactController.GetByReference(ctx, repository, reference, nil)
	if errors.IsNotFoundErr(err) {
		// artifact not found, discontinue the API request
		return nil, errors.BadRequestError(nil).WithMessage("artifact %s not found", from)
	} else if err != nil {
		logger.Errorf("get artifact %s failed, error: %v", from, err)
		return nil, err
	}

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)
	repositoryName := parseRepositoryName(r.URL.EscapedPath())

	if art.ProjectID == projectID && art.RepositoryName == repositoryName {
		return nil, nil
	}

	var artifactDigests []string
	err = artifactController.Walk(ctx, art, func(a *artifact.Artifact) error {
		artifactDigests = append(artifactDigests, a.Digest)
		return nil
	}, nil)
	if err != nil {
		logger.Errorf("walk the artifact %s failed, error: %v", art.Digest, err)
		return nil, err
	}

	allBlobs, err := blobController.List(ctx, blob.ListParams{ArtifactDigests: artifactDigests})
	if err != nil {
		logger.Errorf("get blobs for artifacts %s failed, error: %v", strings.Join(artifactDigests, ", "), err)
		return nil, err
	}

	blobs, err := blobController.FindMissingAssociationsForProject(ctx, projectID, allBlobs)
	if err != nil {
		logger.Errorf("find missing blobs for project %d failed, error: %v", projectID, err)
		return nil, err
	}

	var size int64
	for _, blob := range blobs {
		if !blob.IsForeignLayer() {
			size += blob.Size
		}
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}

func copyArtifactResourcesEvent(level int) func(*http.Request, string, string, string) event.Metadata {
	return func(r *http.Request, reference, referenceID string, message string) event.Metadata {
		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

		query := r.URL.Query()
		from := query.Get("from")
		if from == "" {
			// this will never be happened
			return nil
		}

		repository, reference, err := distribution.ParseRef(from)
		if err != nil {
			// this will never be happened
			return nil
		}

		art, err := artifactController.GetByReference(ctx, repository, reference, nil)
		if err != nil {
			logger.Errorf("get artifact %s failed, error: %v", from, err)
		}

		projectID, _ := strconv.ParseInt(referenceID, 10, 64)
		project, err := projectController.Get(ctx, projectID)
		if err != nil {
			logger.Errorf("get artifact %s failed, error: %v", from, err)
			return nil
		}

		var tag string
		if distribution.IsDigest(reference) {
			tag = reference
		}

		return &metadata.QuotaMetaData{
			Project:  project,
			Tag:      tag,
			Digest:   art.Digest,
			RepoName: parseRepositoryName(r.URL.EscapedPath()),
			Level:    level,
			Msg:      message,
			OccurAt:  time.Now(),
		}
	}
}
