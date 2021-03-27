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

package dao

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"time"

	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/blob/models"
)

// DAO the dao for Blob, ArtifactAndBlob and ProjectBlob
type DAO interface {
	// CreateArtifactAndBlob create ArtifactAndBlob and ignore conflict on artifact digest and blob digest
	CreateArtifactAndBlob(ctx context.Context, artifactDigest, blobDigest string) (int64, error)

	// GetArtifactAndBlob get ArtifactAndBlob by artifact digest and blob digest
	GetArtifactAndBlob(ctx context.Context, artifactDigest, blobDigest string) (*models.ArtifactAndBlob, error)

	// DeleteArtifactAndBlobByArtifact delete ArtifactAndBlob by artifact digest
	DeleteArtifactAndBlobByArtifact(ctx context.Context, artifactDigest string) error

	// GetAssociatedBlobDigestsForArtifact returns blob digests which associated with the artifact
	GetAssociatedBlobDigestsForArtifact(ctx context.Context, artifact string) ([]string, error)

	// CreateBlob create blob and ignore conflict on digest
	CreateBlob(ctx context.Context, blob *models.Blob) (int64, error)

	// GetBlobByDigest returns blob by digest
	GetBlobByDigest(ctx context.Context, digest string) (*models.Blob, error)

	// UpdateBlob update blob
	UpdateBlob(ctx context.Context, blob *models.Blob) error

	// UpdateBlob update blob status
	UpdateBlobStatus(ctx context.Context, blob *models.Blob) (int64, error)

	// ListBlobs list blobs by query
	ListBlobs(ctx context.Context, query *q.Query) ([]*models.Blob, error)

	// FindBlobsShouldUnassociatedWithProject filter the blobs which should not be associated with the project
	FindBlobsShouldUnassociatedWithProject(ctx context.Context, projectID int64, blobs []*models.Blob) ([]*models.Blob, error)

	// SumBlobsSizeByProject returns sum size of blobs by project, skip foreign blobs when `excludeForeignLayer` is true
	SumBlobsSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error)

	// CreateProjectBlob create ProjectBlob and ignore conflict on project id and blob id
	CreateProjectBlob(ctx context.Context, projectID, blobID int64) (int64, error)

	// DeleteProjectBlob delete project blob
	DeleteProjectBlob(ctx context.Context, projectID int64, blobIDs ...int64) error

	// ExistProjectBlob returns true when ProjectBlob exist
	ExistProjectBlob(ctx context.Context, projectID int64, blobDigest string) (bool, error)

	// DeleteBlob delete blob
	DeleteBlob(ctx context.Context, id int64) (err error)

	// GetBlobsNotRefedByProjectBlob get the blobs that are not referenced by the table project_blob and also not in the reserve window(in hours)
	GetBlobsNotRefedByProjectBlob(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) CreateArtifactAndBlob(ctx context.Context, artifactDigest, blobDigest string) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	md := &models.ArtifactAndBlob{
		DigestAF:     artifactDigest,
		DigestBlob:   blobDigest,
		CreationTime: time.Now(),
	}

	return o.InsertOrUpdate(md, "digest_af, digest_blob")
}

func (d *dao) GetArtifactAndBlob(ctx context.Context, artifactDigest, blobDigest string) (*models.ArtifactAndBlob, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	md := &models.ArtifactAndBlob{
		DigestAF:   artifactDigest,
		DigestBlob: blobDigest,
	}

	if err := o.Read(md, "digest_af", "digest_blob"); err != nil {
		return nil, orm.WrapNotFoundError(err, "not found by artifact digest %s and blob digest %s", artifactDigest, blobDigest)
	}

	return md, nil
}

func (d *dao) DeleteArtifactAndBlobByArtifact(ctx context.Context, artifactDigest string) error {
	qs, err := orm.QuerySetter(ctx, &models.ArtifactAndBlob{}, q.New(q.KeyWords{"digest_af": artifactDigest}))
	if err != nil {
		return err
	}

	_, err = qs.Delete()
	return err
}

func (d *dao) GetAssociatedBlobDigestsForArtifact(ctx context.Context, artifact string) ([]string, error) {
	qs, err := orm.QuerySetter(ctx, &models.ArtifactAndBlob{}, q.New(q.KeyWords{"digest_af": artifact}))
	if err != nil {
		return nil, err
	}

	mds := []*models.ArtifactAndBlob{}
	if _, err = qs.All(&mds); err != nil {
		return nil, err
	}

	var blobDigests []string
	for _, md := range mds {
		blobDigests = append(blobDigests, md.DigestBlob)
	}

	return blobDigests, nil
}

func (d *dao) CreateBlob(ctx context.Context, blob *models.Blob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	blob.CreationTime = time.Now()
	// the default status is none
	blob.Status = models.StatusNone

	return o.InsertOrUpdate(blob, "digest")
}

func (d *dao) GetBlobByDigest(ctx context.Context, digest string) (*models.Blob, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	blob := &models.Blob{Digest: digest}
	if err = o.Read(blob, "digest"); err != nil {
		return nil, orm.WrapNotFoundError(err, "blob %s not found", digest)
	}

	return blob, nil
}

func (d *dao) UpdateBlobStatus(ctx context.Context, blob *models.Blob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return -1, err
	}

	var sql string
	if blob.Status == models.StatusNone {
		sql = `UPDATE blob SET version = version + 1, update_time = ?, status = ? where id = ? AND version >= ? AND status IN (%s) RETURNING version as new_version`
	} else {
		sql = `UPDATE blob SET version = version + 1, update_time = ?, status = ? where id = ? AND version = ? AND status IN (%s) RETURNING version as new_version`
	}

	var newVersion int64
	params := []interface{}{time.Now(), blob.Status, blob.ID, blob.Version}
	stats := models.StatusMap[blob.Status]
	for _, stat := range stats {
		params = append(params, stat)
	}
	if err := o.Raw(fmt.Sprintf(sql, orm.ParamPlaceholderForIn(len(models.StatusMap[blob.Status]))), params...).QueryRow(&newVersion); err != nil {
		if e := orm.AsNotFoundError(err, "no blob is updated"); e != nil {
			log.Warningf("no blob is updated according to query condition, id: %d, status_in, %v, err: %v", blob.ID, models.StatusMap[blob.Status], e)
			return 0, nil
		}
		return -1, err
	}

	blob.Version = newVersion
	return 1, nil
}

// UpdateBlob cannot handle the status change and version increase, for handling blob status change, please call
// for the UpdateBlobStatus.
func (d *dao) UpdateBlob(ctx context.Context, blob *models.Blob) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	blob.UpdateTime = time.Now()
	_, err = o.Update(blob, "size", "content_type", "update_time")
	return err
}

func (d *dao) ListBlobs(ctx context.Context, query *q.Query) ([]*models.Blob, error) {
	qs, err := orm.QuerySetter(ctx, &models.Blob{}, query)
	if err != nil {
		return nil, err
	}
	blobs := []*models.Blob{}
	if _, err = qs.All(&blobs); err != nil {
		return nil, err
	}
	return blobs, nil
}

func (d *dao) FindBlobsShouldUnassociatedWithProject(ctx context.Context, projectID int64, blobs []*models.Blob) ([]*models.Blob, error) {
	if len(blobs) == 0 {
		return nil, nil
	}

	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	sql := `SELECT b.digest_blob FROM artifact a, artifact_blob b WHERE a.digest = b.digest_af AND a.project_id = ? AND b.digest_blob IN (%s)`
	params := []interface{}{projectID}
	for _, blob := range blobs {
		params = append(params, blob.Digest)
	}

	var digests []string
	_, err = o.Raw(fmt.Sprintf(sql, orm.ParamPlaceholderForIn(len(blobs))), params...).QueryRows(&digests)
	if err != nil {
		return nil, err
	}

	shouldAssociated := map[string]bool{}
	for _, digest := range digests {
		shouldAssociated[digest] = true
	}

	var results []*models.Blob
	for _, blob := range blobs {
		if !shouldAssociated[blob.Digest] {
			results = append(results, blob)
		}
	}

	return results, nil
}

func (d *dao) SumBlobsSizeByProject(ctx context.Context, projectID int64, excludeForeignLayer bool) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	params := []interface{}{projectID}
	sql := `SELECT SUM(size) FROM blob JOIN project_blob ON blob.id = project_blob.blob_id AND project_id = ?`
	if excludeForeignLayer {
		foreignLayerTypes := []interface{}{
			schema2.MediaTypeForeignLayer,
		}

		sql = fmt.Sprintf(`%s AND content_type NOT IN (%s)`, sql, orm.ParamPlaceholderForIn(len(foreignLayerTypes)))
		params = append(params, foreignLayerTypes...)
	}

	var totalSize int64
	if err := o.Raw(sql, params...).QueryRow(&totalSize); err != nil {
		return 0, err
	}

	return totalSize, nil
}

func (d *dao) CreateProjectBlob(ctx context.Context, projectID, blobID int64) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	md := &models.ProjectBlob{
		ProjectID:    projectID,
		BlobID:       blobID,
		CreationTime: time.Now(),
	}

	// ignore conflict error on (blob_id, project_id)
	return o.InsertOrUpdate(md, "blob_id, project_id")
}

func (d *dao) ExistProjectBlob(ctx context.Context, projectID int64, blobDigest string) (bool, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false, err
	}

	sql := `SELECT COUNT(*) FROM project_blob JOIN blob ON project_blob.blob_id = blob.id AND project_id = ? AND digest = ?`

	var count int64
	if err := o.Raw(sql, projectID, blobDigest).QueryRow(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (d *dao) DeleteProjectBlob(ctx context.Context, projectID int64, blobIDs ...int64) error {
	if len(blobIDs) == 0 {
		return nil
	}
	ol := &q.OrList{}
	for _, blobID := range blobIDs {
		ol.Values = append(ol.Values, blobID)
	}
	kw := q.KeyWords{"blob_id": ol, "project_id": projectID}
	qs, err := orm.QuerySetter(ctx, &models.ProjectBlob{}, q.New(kw))
	if err != nil {
		return err
	}

	_, err = qs.Delete()
	return err
}

func (d *dao) DeleteBlob(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&models.Blob{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("blob %d not found", id)
	}
	return nil
}

func (d *dao) GetBlobsNotRefedByProjectBlob(ctx context.Context, timeWindowHours int64) ([]*models.Blob, error) {
	var noneRefed []*models.Blob
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return noneRefed, err
	}

	sql := fmt.Sprintf(`SELECT b.id, b.digest, b.content_type, b.status, b.version, b.size FROM blob AS b LEFT JOIN project_blob pb ON b.id = pb.blob_id WHERE pb.id IS NULL AND b.update_time <= now() - interval '%d hours';`, timeWindowHours)
	_, err = ormer.Raw(sql).QueryRows(&noneRefed)
	if err != nil {
		return noneRefed, err
	}

	return noneRefed, nil
}
