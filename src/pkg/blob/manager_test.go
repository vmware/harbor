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
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/blob/models"
)

type ManagerTestSuite struct {
	htesting.Suite
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"artifact_blob", "project_blob", "blob"}
}

func (suite *ManagerTestSuite) isAssociatedWithArtifact(ctx context.Context, blobDigest, artifactDigest string) (bool, error) {
	blobs, err := Mgr.List(ctx, ListParams{BlobDigests: []string{blobDigest}, ArtifactDigest: artifactDigest})
	if err != nil {
		return false, err
	}

	return len(blobs) > 0, nil
}

func (suite *ManagerTestSuite) isAssociatedWithProject(ctx context.Context, blobDigest string, projectID int64) (bool, error) {
	blobs, err := Mgr.List(ctx, ListParams{BlobDigests: []string{blobDigest}, ProjectID: projectID})
	if err != nil {
		return false, err
	}

	return len(blobs) > 0, nil
}

func (suite *ManagerTestSuite) TestAssociateWithProject() {
	ctx := suite.Context()

	digest := suite.DigestString()

	blobID, err := Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	projectID := int64(1)

	_, err = Mgr.AssociateWithProject(ctx, blobID, projectID)
	suite.Nil(err)

	associated, err := suite.isAssociatedWithProject(ctx, digest, projectID)
	suite.Nil(err)
	suite.True(associated)
}

func (suite *ManagerTestSuite) TestCleanupAssociationsForArtifact() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blob1Digest := suite.DigestString()
	blob2Digest := suite.DigestString()

	for _, digest := range []string{blob1Digest, blob2Digest} {
		_, err := Mgr.Create(ctx, digest, "media type", 100)
		suite.Nil(err)

		_, err = Mgr.AssociateWithArtifact(ctx, digest, artifactDigest)
		suite.Nil(err)

		associated, err := suite.isAssociatedWithArtifact(ctx, digest, artifactDigest)
		suite.Nil(err)
		suite.True(associated)
	}

	suite.Nil(Mgr.CleanupAssociationsForArtifact(ctx, artifactDigest))

	for _, digest := range []string{blob1Digest, blob2Digest} {
		associated, err := suite.isAssociatedWithArtifact(ctx, digest, artifactDigest)
		suite.Nil(err)
		suite.False(associated)
	}
}

func (suite *ManagerTestSuite) TestCleanupAssociationsForProject() {
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

		ctx := suite.Context()

		blobDigests := []string{digest1, digest2, digest3, digest4, digest5}
		for _, digest := range blobDigests {
			blobID, err := Mgr.Create(ctx, digest, "media type", 100)
			if suite.Nil(err) {
				Mgr.AssociateWithProject(ctx, blobID, projectID)
			}
		}

		blobs, err := Mgr.List(ctx, ListParams{BlobDigests: blobDigests})
		suite.Nil(err)
		suite.Len(blobs, 5)

		for _, digest := range []string{digest1, digest2, digest3} {
			Mgr.AssociateWithArtifact(ctx, digest, artifact1)
		}

		for _, digest := range blobDigests {
			Mgr.AssociateWithArtifact(ctx, digest, artifact2)
		}

		{
			suite.Nil(Mgr.CleanupAssociationsForProject(ctx, projectID, blobs))
			for _, digest := range blobDigests {
				associated, err := suite.isAssociatedWithProject(ctx, digest, projectID)
				suite.Nil(err)
				suite.True(associated)
			}
		}

		suite.ExecSQL(`DELETE FROM artifact WHERE digest = ?`, artifact2)

		{
			suite.Nil(Mgr.CleanupAssociationsForProject(ctx, projectID, blobs))
			for _, digest := range []string{digest1, digest2, digest3} {
				associated, err := suite.isAssociatedWithProject(ctx, digest, projectID)
				suite.Nil(err)
				suite.True(associated)
			}

			for _, digest := range []string{digest4, digest5} {
				associated, err := suite.isAssociatedWithProject(ctx, digest, projectID)
				suite.Nil(err)
				suite.False(associated)
			}
		}
	})
}

func (suite *ManagerTestSuite) TestGet() {
	ctx := suite.Context()

	digest := suite.DigestString()

	blob, err := Mgr.Get(ctx, digest)
	suite.IsNotFoundErr(err)
	suite.Nil(blob)

	_, err = Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	blob, err = Mgr.Get(ctx, digest)
	if suite.Nil(err) {
		suite.Equal(digest, blob.Digest)
		suite.Equal("media type", blob.ContentType)
		suite.Equal(int64(100), blob.Size)
	}
}

func (suite *ManagerTestSuite) TestUpdate() {
	ctx := suite.Context()

	digest := suite.DigestString()
	_, err := Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	blob, err := Mgr.Get(ctx, digest)
	if suite.Nil(err) {
		blob.Size = 1000
		suite.Nil(Mgr.Update(ctx, blob))

		{
			blob, err := Mgr.Get(ctx, digest)
			suite.Nil(err)
			suite.Equal(digest, blob.Digest)
			suite.Equal("media type", blob.ContentType)
			suite.Equal(int64(1000), blob.Size)
			suite.Equal(models.StatusNone, blob.Status)
		}
	}
}

func (suite *ManagerTestSuite) TestList() {
	ctx := suite.Context()

	digest1 := suite.DigestString()
	digest2 := suite.DigestString()

	blobs, err := Mgr.List(ctx, ListParams{BlobDigests: []string{digest1, digest2}})
	suite.Nil(err)
	suite.Len(blobs, 0)

	Mgr.Create(ctx, digest1, "media type", 100)
	Mgr.Create(ctx, digest2, "media type", 100)

	blobs, err = Mgr.List(ctx, ListParams{BlobDigests: []string{digest1, digest2}})
	suite.Nil(err)
	suite.Len(blobs, 2)

	blobs, err = Mgr.List(ctx, models.ListParams{UpdateTime: time.Now().Add(-time.Hour)})
	if suite.Nil(err) {
		suite.Len(blobs, 0)
	}
}

func (suite *ManagerTestSuite) TestListByArtifact() {
	ctx := suite.Context()

	artifact1 := suite.DigestString()
	artifact2 := suite.DigestString()

	digest1 := suite.DigestString()
	digest2 := suite.DigestString()
	digest3 := suite.DigestString()
	digest4 := suite.DigestString()
	digest5 := suite.DigestString()

	blobDigests := []string{digest1, digest2, digest3, digest4, digest5}
	for _, digest := range blobDigests {
		Mgr.Create(ctx, digest, "media type", 100)
	}

	for i, digest := range blobDigests {
		Mgr.AssociateWithArtifact(ctx, digest, artifact1)

		if i < 3 {
			Mgr.AssociateWithArtifact(ctx, digest, artifact2)
		}
	}

	blobs, err := Mgr.List(ctx, ListParams{ArtifactDigest: artifact1})
	suite.Nil(err)
	suite.Len(blobs, 5)

	blobs, err = Mgr.List(ctx, ListParams{ArtifactDigest: artifact2})
	suite.Nil(err)
	suite.Len(blobs, 3)
}

func (suite *ManagerTestSuite) TestDelete() {
	ctx := suite.Context()
	digest := suite.DigestString()
	blobID, err := Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	err = Mgr.Delete(ctx, blobID)
	suite.Nil(err)
}

func (suite *ManagerTestSuite) TestUpdateStatus() {
	ctx := suite.Context()

	digest := suite.DigestString()
	_, err := Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	blob, err := Mgr.Get(ctx, digest)
	if suite.Nil(err) {

		blob.Status = "unknown"
		count, err := Mgr.UpdateBlobStatus(ctx, blob)
		suite.NotNil(err)
		suite.Equal(int64(-1), count)

		// StatusNone cannot be updated to StatusDeleting
		blob.Status = models.StatusDeleting
		count, err = Mgr.UpdateBlobStatus(ctx, blob)
		suite.Nil(err)
		suite.Equal(int64(0), count)

		blob.Status = models.StatusDelete
		count, err = Mgr.UpdateBlobStatus(ctx, blob)
		suite.Nil(err)
		suite.Equal(int64(1), count)

		{
			blob, err := Mgr.Get(ctx, digest)
			suite.Nil(err)
			suite.Equal(digest, blob.Digest)
			suite.Equal(models.StatusDelete, blob.Status)
		}
	}
}

func (suite *ManagerTestSuite) TestUselessBlobs() {
	ctx := suite.Context()

	blobs, err := Mgr.UselessBlobs(ctx, 0)
	suite.Require().Nil(err)
	beforeAdd := len(blobs)

	Mgr.Create(ctx, suite.DigestString(), "media type", 100)
	Mgr.Create(ctx, suite.DigestString(), "media type", 100)
	digest := suite.DigestString()
	blobID, err := Mgr.Create(ctx, digest, "media type", 100)
	suite.Nil(err)

	projectID := int64(1)
	_, err = Mgr.AssociateWithProject(ctx, blobID, projectID)
	suite.Nil(err)

	blobs, err = Mgr.UselessBlobs(ctx, 0)
	suite.Require().Nil(err)
	suite.Require().Equal(2+beforeAdd, len(blobs))

	blobs, err = Mgr.UselessBlobs(ctx, 2)
	suite.Require().Nil(err)
	suite.Require().Equal(0, len(blobs))
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
