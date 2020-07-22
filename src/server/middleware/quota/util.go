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
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

func projectReferenceObject(r *http.Request) (string, string, error) {
	projectName := util.ParseProjectName(r)

	if projectName == "" {
		return "", "", fmt.Errorf("request %s not match any project", r.URL.Path)
	}

	project, err := projectController.GetByName(r.Context(), projectName)
	if err != nil {
		return "", "", err
	}

	return quota.ProjectReference, quota.ReferenceID(project.ProjectID), nil
}

var (
	unmarshalManifest = func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error) {
		lib.NopCloseRequest(r)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, distribution.Descriptor{}, err
		}

		contentType := r.Header.Get("Content-Type")
		return distribution.UnmarshalManifest(contentType, body)
	}
)

func projectResourcesEvent(level int) func(*http.Request, string, string, string) event.Metadata {
	return func(r *http.Request, reference, referenceID string, message string) event.Metadata {
		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

		path := r.URL.EscapedPath()

		var (
			digest string
			tag    string
		)
		if distribution.ManifestURLRegexp.MatchString(path) {
			_, descriptor, err := unmarshalManifest(r)
			if err != nil {
				logger.Errorf("unmarshal manifest failed, error: %v", err)
				return nil
			}

			digest = descriptor.Digest.String()
			if ref := distribution.ParseReference(path); !distribution.IsDigest(ref) {
				tag = ref
			}
		}

		projectID, _ := strconv.ParseInt(referenceID, 10, 64)
		project, err := projectController.Get(ctx, projectID)
		if err != nil {
			logger.Errorf("get project %d failed, error: %v", projectID, err)

			return nil
		}

		return &metadata.QuotaMetaData{
			Project:  project,
			Tag:      tag,
			Digest:   digest,
			RepoName: distribution.ParseName(path),
			Level:    level,
			Msg:      message,
			OccurAt:  time.Now(),
		}
	}
}
