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

package distribution

import (
	"fmt"
	"regexp"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"

	// docker schema1 manifest
	_ "github.com/docker/distribution/manifest/schema1"
	// docker schema2 manifest
	_ "github.com/docker/distribution/manifest/schema2"
	// manifestlist
	_ "github.com/docker/distribution/manifest/manifestlist"
	// oci schema
	_ "github.com/docker/distribution/manifest/ocischema"
	ref "github.com/docker/distribution/reference"
	"github.com/goharbor/harbor/src/common/utils"
)

// Descriptor alias type of github.com/docker/distribution.Descriptor
type Descriptor = distribution.Descriptor

// Manifest alias type of github.com/docker/distribution.Manifest
type Manifest = distribution.Manifest

var (
	// UnmarshalManifest alias func from `github.com/docker/distribution`
	UnmarshalManifest = distribution.UnmarshalManifest
)

var (
	name      = fmt.Sprintf("(?P<name>%s)", ref.NameRegexp)
	reference = fmt.Sprintf("(?P<reference>((%s)|(%s)))", ref.DigestRegexp, ref.TagRegexp)
	dgt       = fmt.Sprintf("(?P<digest>%s)", ref.DigestRegexp)
	sessionID = "(?P<session_id>[a-zA-Z0-9-_.=]+)"

	// BlobURLRegexp regexp which match blob url
	BlobURLRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/` + dgt)

	// BlobUploadURLRegexp regexp which match blob upload url
	BlobUploadURLRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads/` + sessionID)

	// InitiateBlobUploadRegexp regexp which match initiate blob upload url
	InitiateBlobUploadRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads`)

	// ManifestURLRegexp regexp which match manifest url
	ManifestURLRegexp = regexp.MustCompile(`^/v2/` + name + `/manifests/` + reference)
)

var (
	extractNameRegexp      = regexp.MustCompile(`^/v2/` + name + `/(manifests|blobs|tags)`)
	extractSessionIDRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads/` + sessionID)
)

// ParseName returns name value from distribution API URL path
func ParseName(path string) string {
	m := utils.FindNamedMatches(extractNameRegexp, path)
	if len(m) > 0 {
		return m["name"]
	}

	return ""
}

// ParseReference returns digest or tag from distribution API URL path
func ParseReference(path string) string {
	m := utils.FindNamedMatches(ManifestURLRegexp, path)
	if len(m) > 0 {
		return m["reference"]
	}

	return ""
}

// ParseProjectName returns project name from distribution API URL path
func ParseProjectName(path string) string {
	projectName, _ := utils.ParseRepository(ParseName(path))
	return projectName
}

// ParseSessionID returns session id value from distribution API URL path
func ParseSessionID(path string) string {
	m := utils.FindNamedMatches(extractSessionIDRegexp, path)
	if len(m) > 0 {
		return m["session_id"]
	}

	return ""
}

// ParseRef parse "repository:tag" or "repository@digest" into repository and reference parts
func ParseRef(s string) (string, string, error) {
	matches := ref.ReferenceRegexp.FindStringSubmatch(s)
	if matches == nil {
		return "", "", fmt.Errorf("invalid input: %s", s)
	}

	repository := matches[1]
	reference := matches[2]
	if matches[3] != "" {
		_, err := digest.Parse(matches[3])
		if err != nil {
			return "", "", fmt.Errorf("invalid input: %s", s)
		}
		reference = matches[3]
	}

	return repository, reference, nil
}

// IsDigest returns true when reference is digest
func IsDigest(reference string) bool {
	return ref.DigestRegexp.MatchString(reference)
}
