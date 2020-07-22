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
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/blob/dao"
	"github.com/goharbor/harbor/src/pkg/blob/models"
)

// Blob alias `models.Blob` to make it natural to use the Manager
type Blob = models.Blob

// ListParams alias `models.ListParams` to make it natural to use the Manager
type ListParams = models.ListParams

var (
	// Mgr default blob manager
	Mgr = NewManager()
)

// Manager interface provide the management functions for blobs
type Manager interface {
	// AssociateWithArtifact associate blob with artifact
	AssociateWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error)

	// AssociateWithProject associate blob with project
	AssociateWithProject(ctx context.Context, blobID, projectID int64) (int64, error)

	// CalculateTotalSizeByProject returns total blob size by project, skip foreign blobs when `excludeForeignLayer` is true
	CalculateTotalSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error)

	// Create create blob
	Create(ctx context.Context, digest string, contentType string, size int64) (int64, error)

	// CleanupAssociationsForArtifact remove all associations between blob and artifact by artifact digest
	CleanupAssociationsForArtifact(ctx context.Context, artifactDigest string) error

	// CleanupAssociationsForProject remove unneeded associations between blobs and project
	CleanupAssociationsForProject(ctx context.Context, projectID int64, blobs []*Blob) error

	// Get get blob by digest
	Get(ctx context.Context, digest string) (*Blob, error)

	// Update the blob
	Update(ctx context.Context, blob *Blob) error

	// Update the blob status
	UpdateBlobStatus(ctx context.Context, blob *models.Blob) (int64, error)

	// List returns blobs by params
	List(ctx context.Context, params ListParams) ([]*Blob, error)

	// DeleteBlob delete blob
	Delete(ctx context.Context, id int64) (err error)

	// UselessBlobs useless blob is the blob that is not used in any of projects.
	UselessBlobs(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error)
}

type manager struct {
	dao dao.DAO
}

func (m *manager) AssociateWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error) {
	return m.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest)
}

func (m *manager) AssociateWithProject(ctx context.Context, blobID, projectID int64) (int64, error) {
	return m.dao.CreateProjectBlob(ctx, projectID, blobID)
}

func (m *manager) CalculateTotalSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error) {
	return m.dao.SumBlobsSizeByProject(ctx, projectID, excludeForeignLayer)
}

func (m *manager) Create(ctx context.Context, digest string, contentType string, size int64) (int64, error) {
	return m.dao.CreateBlob(ctx, &Blob{Digest: digest, ContentType: contentType, Size: size})
}

func (m *manager) CleanupAssociationsForArtifact(ctx context.Context, artifactDigest string) error {
	return m.dao.DeleteArtifactAndBlobByArtifact(ctx, artifactDigest)
}

func (m *manager) CleanupAssociationsForProject(ctx context.Context, projectID int64, blobs []*Blob) error {
	if len(blobs) == 0 {
		return nil
	}

	shouldUnassociatedBlobs, err := m.dao.FindBlobsShouldUnassociatedWithProject(ctx, projectID, blobs)
	if err != nil {
		return err
	}

	var blobIDs []int64
	for _, blob := range shouldUnassociatedBlobs {
		blobIDs = append(blobIDs, blob.ID)
	}

	return m.dao.DeleteProjectBlob(ctx, projectID, blobIDs...)
}

func (m *manager) Get(ctx context.Context, digest string) (*Blob, error) {
	return m.dao.GetBlobByDigest(ctx, digest)
}

func (m *manager) Update(ctx context.Context, blob *Blob) error {
	return m.dao.UpdateBlob(ctx, blob)
}

func (m *manager) UpdateBlobStatus(ctx context.Context, blob *models.Blob) (int64, error) {
	_, exist := models.StatusMap[blob.Status]
	if !exist {
		return -1, errors.New(nil).WithMessage("cannot update blob status, as the status is unknown. digest: %s, status: %s", blob.Digest, blob.Status)
	}
	return m.dao.UpdateBlobStatus(ctx, blob)
}

func (m *manager) List(ctx context.Context, params ListParams) ([]*Blob, error) {
	return m.dao.ListBlobs(ctx, params)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.DeleteBlob(ctx, id)
}

func (m *manager) UselessBlobs(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error) {
	return m.dao.GetBlobsNotRefedByProjectBlob(ctx, timeWindowHours)
}

// NewManager returns blob manager
func NewManager() Manager {
	return &manager{dao: dao.New()}
}
