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

package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/models"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"time"
)

func init() {
	orm.RegisterModel(&Blob{})
}

// TODO: move ArtifactAndBlob, ProjectBlob to here

// ArtifactAndBlob alias ArtifactAndBlob model
type ArtifactAndBlob = models.ArtifactAndBlob

/*
the status are used for Garbage Collection
StatusNone, the blob is using in Harbor as normal.
StatusDelete, the blob is marked as GC candidate.
StatusDeleting, the blob undergo a GC blob deletion.
StatusDeleteFailed, the blob is failed to delete from the backend storage.

The status transitions
StatusNone -> StatusDelete : Mark the blob as candidate.
StatusDelete -> StatusDeleting : Select the blob and call the API to delete asset in the backend storage.
StatusDeleting -> Trash : Delete success from the backend storage.
StatusDelete -> StatusNone : Client asks the existence of blob, remove it from the candidate.
StatusDelete -> StatusDeleteFailed : The storage driver returns fail when to delete the real data from the configurated file system.
StatusDeleteFailed -> StatusNone : The delete failed blobs can be pushed again, and back to normal.
StatusDeleteFailed -> StatusDelete : The delete failed blobs should be in the candidate.
*/
const (
	StatusNone         = ""
	StatusDelete       = "delete"
	StatusDeleting     = "deleting"
	StatusDeleteFailed = "deletefailed"
)

// StatusMap key is the target status, values are the accept source status. For example, only StatusNone and StatusDeleteFailed can be convert to StatusDelete.
var StatusMap = map[string][]string{
	StatusNone:         {StatusNone, StatusDelete, StatusDeleteFailed},
	StatusDelete:       {StatusNone, StatusDeleteFailed},
	StatusDeleting:     {StatusDelete},
	StatusDeleteFailed: {StatusDeleting},
}

// Blob holds the details of a blob.
type Blob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	ContentType  string    `orm:"column(content_type)" json:"content_type"`
	Size         int64     `orm:"column(size)" json:"size"`
	Status       string    `orm:"column(status)" json:"status"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now_add" json:"update_time"`
	Version      int64     `orm:"column(version)" json:"version"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (b *Blob) TableName() string {
	return "blob"
}

// IsForeignLayer returns true if the blob is foreign layer
func (b *Blob) IsForeignLayer() bool {
	return b.ContentType == schema2.MediaTypeForeignLayer
}

// IsManifest returns true if the blob is manifest layer
func (b *Blob) IsManifest() bool {
	return b.ContentType == schema2.MediaTypeManifest || b.ContentType == schema1.MediaTypeManifest ||
		b.ContentType == v1.MediaTypeImageManifest || b.ContentType == v1.MediaTypeImageIndex || b.ContentType == manifestlist.MediaTypeManifestList
}

// ProjectBlob alias ProjectBlob model
type ProjectBlob = models.ProjectBlob

// ListParams list params
type ListParams struct {
	ArtifactDigest  string    // list blobs which associated with the artifact
	ArtifactDigests []string  // list blobs which associated with these artifacts
	BlobDigests     []string  // list blobs which digest in the digests
	ProjectID       int64     // list blobs which associated with the project
	UpdateTime      time.Time // list blobs which update time less than updatetime
}
