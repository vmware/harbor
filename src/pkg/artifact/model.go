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

package artifact

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Artifact is the abstract object managed by Harbor. It hides the
// underlying concrete detail and provides an unified artifact view
// for all users.
type Artifact struct {
	ID                int64                  `json:"id"`
	Type              string                 `json:"type"`                // image, chart, etc
	MediaType         string                 `json:"media_type"`          // the media type of artifact. Mostly, it's the value of `manifest.config.mediatype`
	ManifestMediaType string                 `json:"manifest_media_type"` // the media type of manifest/index
	ProjectID         int64                  `json:"project_id"`
	RepositoryID      int64                  `json:"repository_id"`
	RepositoryName    string                 `json:"repository_name"`
	Digest            string                 `json:"digest"`
	Size              int64                  `json:"size"`
	Icon              string                 `json:"icon"`
	PushTime          time.Time              `json:"push_time"`
	PullTime          time.Time              `json:"pull_time"`
	ExtraAttrs        map[string]interface{} `json:"extra_attrs"` // only contains the simple attributes specific for the different artifact type, most of them should come from the config layer
	Annotations       map[string]string      `json:"annotations"`
	References        []*Reference           `json:"references"` // child artifacts referenced by the parent artifact if the artifact is an index
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s@%s", a.RepositoryName, a.Digest)
}

// IsImageIndex returns true when artifact is image index
func (a *Artifact) IsImageIndex() bool {
	return a.ManifestMediaType == v1.MediaTypeImageIndex ||
		a.ManifestMediaType == manifestlist.MediaTypeManifestList
}

// From converts the database level artifact to the business level object
func (a *Artifact) From(art *dao.Artifact) {
	a.ID = art.ID
	a.Type = art.Type
	a.MediaType = art.MediaType
	a.ManifestMediaType = art.ManifestMediaType
	a.ProjectID = art.ProjectID
	a.RepositoryID = art.RepositoryID
	a.RepositoryName = art.RepositoryName
	a.Digest = art.Digest
	a.Size = art.Size
	a.Icon = art.Icon
	a.PushTime = art.PushTime
	a.PullTime = art.PullTime
	a.ExtraAttrs = map[string]interface{}{}
	a.Annotations = map[string]string{}
	if len(art.ExtraAttrs) > 0 {
		if err := json.Unmarshal([]byte(art.ExtraAttrs), &a.ExtraAttrs); err != nil {
			log.Errorf("failed to unmarshal the extra attrs of artifact %d: %v", art.ID, err)
		}
	}
	if len(art.Annotations) > 0 {
		if err := json.Unmarshal([]byte(art.Annotations), &a.Annotations); err != nil {
			log.Errorf("failed to unmarshal the annotations of artifact %d: %v", art.ID, err)
		}
	}
}

// To converts the artifact to the database level object
func (a *Artifact) To() *dao.Artifact {
	art := &dao.Artifact{
		ID:                a.ID,
		Type:              a.Type,
		MediaType:         a.MediaType,
		ManifestMediaType: a.ManifestMediaType,
		ProjectID:         a.ProjectID,
		RepositoryID:      a.RepositoryID,
		RepositoryName:    a.RepositoryName,
		Digest:            a.Digest,
		Size:              a.Size,
		Icon:              a.Icon,
		PushTime:          a.PushTime,
		PullTime:          a.PullTime,
	}
	if len(a.ExtraAttrs) > 0 {
		attrs, err := json.Marshal(a.ExtraAttrs)
		if err != nil {
			log.Errorf("failed to marshal the extra attrs of artifact %d: %v", a.ID, err)
		}
		art.ExtraAttrs = string(attrs)
	}
	if len(a.Annotations) > 0 {
		annotations, err := json.Marshal(a.Annotations)
		if err != nil {
			log.Errorf("failed to marshal the annotations of artifact %d: %v", a.ID, err)
		}
		art.Annotations = string(annotations)
	}
	return art
}

// Reference records the child artifact referenced by parent artifact
type Reference struct {
	ID          int64  `json:"id"`
	ParentID    int64  `json:"parent_id"`
	ChildID     int64  `json:"child_id"`
	ChildDigest string `json:"child_digest"`
	Platform    *v1.Platform
	URLs        []string          `json:"urls"`
	Annotations map[string]string `json:"annotations"`
}

// From converts the data level reference to business level
func (r *Reference) From(ref *dao.ArtifactReference) {
	r.ID = ref.ID
	r.ParentID = ref.ParentID
	r.ChildID = ref.ChildID
	r.ChildDigest = ref.ChildDigest
	if len(ref.Platform) > 0 {
		r.Platform = &v1.Platform{}
		if err := json.Unmarshal([]byte(ref.Platform), r.Platform); err != nil {
			log.Errorf("failed to unmarshal the platform of reference: %v", err)
		}
	}
	if len(ref.URLs) > 0 {
		r.URLs = []string{}
		if err := json.Unmarshal([]byte(ref.URLs), &r.URLs); err != nil {
			log.Errorf("failed to unmarshal the URLs of reference: %v", err)
		}
	}
	if len(ref.Annotations) > 0 {
		r.Annotations = map[string]string{}
		if err := json.Unmarshal([]byte(ref.Annotations), &r.Annotations); err != nil {
			log.Errorf("failed to unmarshal the annotations of reference: %v", err)
		}
	}
}

// To converts the reference to data level object
func (r *Reference) To() *dao.ArtifactReference {
	ref := &dao.ArtifactReference{
		ID:          r.ID,
		ParentID:    r.ParentID,
		ChildID:     r.ChildID,
		ChildDigest: r.ChildDigest,
	}
	if r.Platform != nil {
		platform, err := json.Marshal(r.Platform)
		if err != nil {
			log.Errorf("failed to marshal the platform of reference: %v", err)
		}
		ref.Platform = string(platform)
	}
	if len(r.URLs) > 0 {
		urls, err := json.Marshal(r.URLs)
		if err != nil {
			log.Errorf("failed to marshal the URLs of reference: %v", err)
		}
		ref.URLs = string(urls)
	}
	if len(r.Annotations) > 0 {
		annotations, err := json.Marshal(r.Annotations)
		if err != nil {
			log.Errorf("failed to marshal the annotations of reference: %v", err)
		}
		ref.Annotations = string(annotations)
	}
	return ref
}
