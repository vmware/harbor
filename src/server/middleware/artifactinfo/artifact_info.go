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

package artifactinfo

import (
	"fmt"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/opencontainers/go-digest"
)

const (
	blobMountQuery  = "mount"
	blobFromQuery   = "from"
	blobMountDigest = "blob_mount_digest"
	blobMountRepo   = "blob_mount_repo"
	tag             = "tag"
)

var (
	urlPatterns = map[string]*regexp.Regexp{
		"manifest":    middleware.V2ManifestURLRe,
		"tag_list":    middleware.V2TagListURLRe,
		"blob_upload": middleware.V2BlobUploadURLRe,
		"blob":        middleware.V2BlobURLRe,
	}
)

// Middleware gets the information of artifact via url of the request and inject it into the context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			log.Debugf("In artifact info middleware, url: %s", req.URL.String())
			m, ok := parse(req.URL)
			if !ok {
				next.ServeHTTP(rw, req)
				return
			}
			repo := m[middleware.RepositorySubexp]
			pn, err := projectNameFromRepo(repo)
			if err != nil {
				lib_http.SendError(rw, errors.BadRequestError(err))
				return
			}
			art := lib.ArtifactInfo{
				Repository:  repo,
				ProjectName: pn,
			}
			if d, ok := m[middleware.DigestSubexp]; ok {
				art.Digest = d
			}
			if ref, ok := m[middleware.ReferenceSubexp]; ok {
				art.Reference = ref
			}
			if t, ok := m[tag]; ok {
				art.Tag = t
			}
			if bmr, ok := m[blobMountRepo]; ok {
				// Fail early for now, though in docker registry an invalid may return 202
				// it's not clear in OCI spec how to handle invalid from parm
				bmp, err := projectNameFromRepo(bmr)
				if err != nil {
					lib_http.SendError(rw, errors.BadRequestError(err))
					return
				}
				art.BlobMountDigest = m[blobMountDigest]
				art.BlobMountProjectName = bmp
				art.BlobMountRepository = bmr
			}
			ctx := lib.WithArtifactInfo(req.Context(), art)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func projectNameFromRepo(repo string) (string, error) {
	components := strings.SplitN(repo, "/", 2)
	if len(components) < 2 {
		return "", fmt.Errorf("invalid repository name: %s", repo)
	}
	return components[0], nil
}

func parse(url *url.URL) (map[string]string, bool) {
	path := url.Path
	query := url.Query()
	m := make(map[string]string)
	match := false
	for key, re := range urlPatterns {
		l := re.FindStringSubmatch(path)
		if len(l) > 0 {
			match = true
			for i := 1; i < len(l); i++ {
				m[re.SubexpNames()[i]] = l[i]
			}
			if key == "blob_upload" && len(query.Get(blobFromQuery)) > 0 {
				m[blobMountDigest] = query.Get(blobMountQuery)
				m[blobMountRepo] = query.Get(blobFromQuery)
			}
			break
		}
	}
	if digest.DigestRegexp.MatchString(m[middleware.ReferenceSubexp]) {
		m[middleware.DigestSubexp] = m[middleware.ReferenceSubexp]
	} else if ref, ok := m[middleware.ReferenceSubexp]; ok {
		m[tag] = ref
	}
	return m, match
}
