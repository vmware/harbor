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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"blob", "artifact_blob", "project_blob"}
	suite.dao = New()
}

func (suite *DaoTestSuite) TestCreateArtifactAndBlob() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blobDigest := suite.DigestString()

	_, err := suite.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest)
	suite.Nil(err)

	_, err = suite.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestGetArtifactAndBlob() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blobDigest := suite.DigestString()

	md, err := suite.dao.GetArtifactAndBlob(ctx, artifactDigest, blobDigest)
	suite.IsNotFoundErr(err)
	suite.Nil(md)

	_, err = suite.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest)
	suite.Nil(err)

	md, err = suite.dao.GetArtifactAndBlob(ctx, artifactDigest, blobDigest)
	if suite.Nil(err) {
		suite.Equal(artifactDigest, md.DigestAF)
		suite.Equal(blobDigest, md.DigestBlob)
	}
}

func (suite *DaoTestSuite) TestDeleteArtifactAndBlobByArtifact() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blobDigest1 := suite.DigestString()
	blobDigest2 := suite.DigestString()

	_, err := suite.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest1)
	suite.Nil(err)

	_, err = suite.dao.CreateArtifactAndBlob(ctx, artifactDigest, blobDigest2)
	suite.Nil(err)

	digests, err := suite.dao.GetAssociatedBlobDigestsForArtifact(ctx, artifactDigest)
	suite.Nil(err)
	suite.Len(digests, 2)

	suite.Nil(suite.dao.DeleteArtifactAndBlobByArtifact(ctx, artifactDigest))

	digests, err = suite.dao.GetAssociatedBlobDigestsForArtifact(ctx, artifactDigest)
	suite.Nil(err)
	suite.Len(digests, 0)
}

func (suite *DaoTestSuite) TestGetAssociatedBlobDigestsForArtifact() {

}

func (suite *DaoTestSuite) TestCreateBlob() {
	ctx := suite.Context()

	digest := suite.DigestString()

	_, err := suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	suite.Nil(err)

	_, err = suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestGetBlobByDigest() {
	ctx := suite.Context()

	digest := suite.DigestString()

	blob, err := suite.dao.GetBlobByDigest(ctx, digest)
	suite.IsNotFoundErr(err)
	suite.Nil(blob)

	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})

	blob, err = suite.dao.GetBlobByDigest(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(digest, blob.Digest)
	}
}

func (suite *DaoTestSuite) TestUpdateBlob() {
	ctx := suite.Context()

	digest := suite.DigestString()

	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	blob, err := suite.dao.GetBlobByDigest(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(int64(0), blob.Size)
	}

	blob.Size = 100
	if suite.Nil(suite.dao.UpdateBlob(ctx, blob)) {
		blob, err := suite.dao.GetBlobByDigest(ctx, digest)
		if suite.Nil(err) {
			suite.Equal(int64(100), blob.Size)
			suite.Equal(int64(0), blob.Version)
		}
	}

	blob.Status = "deleting"
	suite.Nil(suite.dao.UpdateBlob(ctx, blob), "cannot be updated.")
	blob, err = suite.dao.GetBlobByDigest(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(int64(0), blob.Version)
		suite.Equal(models.StatusNone, blob.Status)
	}
}

func (suite *DaoTestSuite) TestUpdateBlobStatus() {
	ctx := suite.Context()

	digest := suite.DigestString()

	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	blob, err := suite.dao.GetBlobByDigest(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(int64(0), blob.Size)
	}

	// StatusNone cannot be updated to StatusDeleting directly
	blob.Status = models.StatusDeleting
	count, err := suite.dao.UpdateBlobStatus(ctx, blob)
	suite.Nil(err)
	suite.Equal(int64(0), count)

	blob.Status = models.StatusDelete
	count, err = suite.dao.UpdateBlobStatus(ctx, blob)
	suite.Nil(err)
	suite.Equal(int64(1), count)

	blob.Status = models.StatusDeleting
	count, err = suite.dao.UpdateBlobStatus(ctx, blob)
	suite.Nil(err)
	suite.Equal(int64(1), count)

	blob.Status = models.StatusDeleteFailed
	count, err = suite.dao.UpdateBlobStatus(ctx, blob)
	suite.Nil(err)
	suite.Equal(int64(1), count)

	blob, err = suite.dao.GetBlobByDigest(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(int64(3), blob.Version)
		suite.Equal(models.StatusDeleteFailed, blob.Status)
	}
}

func (suite *DaoTestSuite) TestListBlobs() {
	ctx := suite.Context()

	digest1 := suite.DigestString()
	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest1})

	digest2 := suite.DigestString()
	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest2})

	blobs, err := suite.dao.ListBlobs(ctx, models.ListParams{BlobDigests: []string{digest1}})
	if suite.Nil(err) {
		suite.Len(blobs, 1)
	}

	blobs, err = suite.dao.ListBlobs(ctx, models.ListParams{BlobDigests: []string{digest1, digest2}})
	if suite.Nil(err) {
		suite.Len(blobs, 2)
	}

	blobs, err = suite.dao.ListBlobs(ctx, models.ListParams{UpdateTime: time.Now().Add(-time.Hour)})
	if suite.Nil(err) {
		suite.Len(blobs, 0)
	}

	digest3 := suite.DigestString()
	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest3, UpdateTime: time.Now().Add(-time.Hour * 2)})
	blobs, err = suite.dao.ListBlobs(ctx, models.ListParams{UpdateTime: time.Now().Add(-time.Hour)})
	if suite.Nil(err) {
		suite.Len(blobs, 1)
	}
}

func (suite *DaoTestSuite) TestListBlobsAssociatedWithArtifact() {

}

func (suite *DaoTestSuite) TestFindBlobsShouldUnassociatedWithProject() {
	ctx := suite.Context()

	suite.WithProject(func(projectID int64, projectName string) {
		artifact1 := suite.DigestString()
		artifact2 := suite.DigestString()

		sql := `INSERT INTO artifact ("type", media_type, manifest_media_type, digest, project_id, repository_id, repository_name) VALUES ('image', 'media_type', 'manifest_media_type', ?, ?, ?, 'library/hello-world')`
		suite.ExecSQL(sql, artifact1, projectID, 10)
		suite.ExecSQL(sql, artifact2, projectID, 10)

		defer suite.ExecSQL(`DELETE FROM artifact WHERE project_id = ?`, projectID)

		digest1 := suite.DigestString()
		digest2 := suite.DigestString()
		digest3 := suite.DigestString()
		digest4 := suite.DigestString()
		digest5 := suite.DigestString()

		blobDigests := []string{digest1, digest2, digest3, digest4, digest5}
		for _, digest := range blobDigests {
			blobID, err := suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
			if suite.Nil(err) {
				suite.dao.CreateProjectBlob(ctx, projectID, blobID)
			}
		}

		blobs, err := suite.dao.ListBlobs(ctx, models.ListParams{BlobDigests: blobDigests})
		suite.Nil(err)
		suite.Len(blobs, 5)

		for _, digest := range []string{digest1, digest2, digest3} {
			suite.dao.CreateArtifactAndBlob(ctx, artifact1, digest)
		}

		for _, digest := range blobDigests {
			suite.dao.CreateArtifactAndBlob(ctx, artifact2, digest)
		}

		{
			results, err := suite.dao.FindBlobsShouldUnassociatedWithProject(ctx, projectID, blobs)
			suite.Nil(err)
			suite.Len(results, 0)
		}

		suite.ExecSQL(`DELETE FROM artifact WHERE digest = ?`, artifact2)

		{
			results, err := suite.dao.FindBlobsShouldUnassociatedWithProject(ctx, projectID, blobs)
			suite.Nil(err)
			if suite.Len(results, 2) {
				suite.Contains([]string{results[0].Digest, results[1].Digest}, digest4)
				suite.Contains([]string{results[0].Digest, results[1].Digest}, digest5)
			}

		}
	})

}

func (suite *DaoTestSuite) TestCreateProjectBlob() {
	ctx := suite.Context()

	projectID := int64(1)
	blobID := int64(1000)

	_, err := suite.dao.CreateProjectBlob(ctx, projectID, blobID)
	suite.Nil(err)

	_, err = suite.dao.CreateProjectBlob(ctx, projectID, blobID)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestExistProjectBlob() {
	ctx := suite.Context()

	digest := suite.DigestString()

	projectID := int64(1)

	exist, err := suite.dao.ExistProjectBlob(ctx, projectID, digest)
	suite.Nil(err)
	suite.False(exist)

	blobID, err := suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	suite.Nil(err)

	_, err = suite.dao.CreateProjectBlob(ctx, projectID, blobID)
	suite.Nil(err)

	exist, err = suite.dao.ExistProjectBlob(ctx, projectID, digest)
	suite.Nil(err)
	suite.True(exist)
}

func (suite *DaoTestSuite) TestDeleteProjectBlob() {
	ctx := suite.Context()

	digest := suite.DigestString()
	blobID, err := suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	suite.Nil(err)

	projectID1 := int64(1)
	projectID2 := int64(2)
	projectID3 := int64(3)

	_, err = suite.dao.CreateProjectBlob(ctx, projectID1, blobID)
	suite.Nil(err)

	_, err = suite.dao.CreateProjectBlob(ctx, projectID2, blobID)
	suite.Nil(err)

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID1, digest)
		suite.Nil(err)
		suite.True(exist)
	}

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID2, digest)
		suite.Nil(err)
		suite.True(exist)
	}

	suite.Nil(suite.dao.DeleteProjectBlob(ctx, projectID3, blobID))

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID1, digest)
		suite.Nil(err)
		suite.True(exist)
	}

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID2, digest)
		suite.Nil(err)
		suite.True(exist)
	}

	suite.Nil(suite.dao.DeleteProjectBlob(ctx, projectID1, blobID))

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID1, digest)
		suite.Nil(err)
		suite.False(exist)
	}

	{
		exist, err := suite.dao.ExistProjectBlob(ctx, projectID2, digest)
		suite.Nil(err)
		suite.True(exist)
	}
}

func (suite *DaoTestSuite) TestDelete() {
	ctx := suite.Context()

	err := suite.dao.DeleteBlob(ctx, 100021)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	digest := suite.DigestString()
	id, err := suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})
	suite.Nil(err)
	err = suite.dao.DeleteBlob(ctx, id)
	suite.Require().Nil(err)
}

func (suite *DaoTestSuite) TestGetBlobsNotRefedByProjectBlob() {
	ctx := suite.Context()

	blobs, err := suite.dao.GetBlobsNotRefedByProjectBlob(ctx, 0)
	suite.Require().Nil(err)
	beforeAdd := len(blobs)

	suite.dao.CreateBlob(ctx, &models.Blob{Digest: suite.DigestString()})
	suite.dao.CreateBlob(ctx, &models.Blob{Digest: suite.DigestString()})
	digest := suite.DigestString()
	suite.dao.CreateBlob(ctx, &models.Blob{Digest: digest})

	blob, err := suite.dao.GetBlobByDigest(ctx, digest)
	suite.Nil(err)

	projectID := int64(1)
	_, err = suite.dao.CreateProjectBlob(ctx, projectID, blob.ID)
	suite.Nil(err)

	blobs, err = suite.dao.GetBlobsNotRefedByProjectBlob(ctx, 0)
	suite.Require().Nil(err)
	suite.Require().Equal(2+beforeAdd, len(blobs))

	blobs, err = suite.dao.GetBlobsNotRefedByProjectBlob(ctx, 2)
	suite.Require().Nil(err)
	suite.Require().Equal(0, len(blobs))
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
