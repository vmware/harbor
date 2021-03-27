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
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/blob/models"

	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/distribution"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
	blobtesting "github.com/goharbor/harbor/src/testing/pkg/blob"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	pkg_blob "github.com/goharbor/harbor/src/pkg/blob"
)

type ControllerTestSuite struct {
	htesting.Suite
}

func (suite *ControllerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"blob", "artifact_blob", "project_blob"}
}

func (suite *ControllerTestSuite) prepareBlob() string {

	ctx := suite.Context()
	digest := suite.DigestString()

	_, err := Ctl.Ensure(ctx, digest, "application/octet-stream", 100)
	suite.Nil(err)

	return digest
}

func (suite *ControllerTestSuite) TestAttachToArtifact() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blobDigests := []string{
		suite.prepareBlob(),
		suite.prepareBlob(),
		suite.prepareBlob(),
	}

	suite.Nil(Ctl.AssociateWithArtifact(ctx, blobDigests, artifactDigest))

	for _, digest := range blobDigests {
		exist, err := Ctl.Exist(ctx, digest, IsAssociatedWithArtifact(artifactDigest))
		suite.Nil(err)
		suite.True(exist)
	}

	suite.Nil(Ctl.AssociateWithArtifact(ctx, blobDigests, artifactDigest))
}

func (suite *ControllerTestSuite) TestAttachToProjectByDigest() {
	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		digest := suite.prepareBlob()
		suite.Nil(Ctl.AssociateWithProjectByDigest(ctx, digest, projectID))

		exist, err := Ctl.Exist(ctx, digest, IsAssociatedWithProject(projectID))
		suite.Nil(err)
		suite.True(exist)
	})
}

func (suite *ControllerTestSuite) TestCalculateTotalSizeByProject() {
	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		id1, _ := Ctl.Ensure(ctx, suite.DigestString(), schema2.MediaTypeForeignLayer, 100)
		Ctl.AssociateWithProjectByID(ctx, id1, projectID)
		id2, _ := Ctl.Ensure(ctx, suite.DigestString(), schema2.MediaTypeLayer, 100)
		Ctl.AssociateWithProjectByID(ctx, id2, projectID)

		{
			size, err := Ctl.CalculateTotalSizeByProject(ctx, projectID, true)
			suite.Nil(err)
			suite.Equal(int64(100), size)
		}

		{
			size, err := Ctl.CalculateTotalSizeByProject(ctx, projectID, false)
			suite.Nil(err)
			suite.Equal(int64(200), size)
		}
	})
}

func (suite *ControllerTestSuite) TestEnsure() {
	ctx := suite.Context()

	digest := suite.DigestString()

	_, err := Ctl.Ensure(ctx, digest, "application/octet-stream", 100)
	suite.Nil(err)

	exist, err := Ctl.Exist(ctx, digest)
	suite.Nil(err)
	suite.True(exist)

	_, err = Ctl.Ensure(ctx, digest, "application/octet-stream", 100)
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestExist() {
	ctx := suite.Context()

	exist, err := Ctl.Exist(ctx, suite.DigestString())
	suite.Nil(err)
	suite.False(exist)
}

func (suite *ControllerTestSuite) TestFindMissingAssociationsForProjectByArtifact() {
	blobMgr := &blobtesting.Manager{}

	ctl := &controller{blobMgr: blobMgr}

	ctx := context.TODO()
	projectID := int64(1)

	{
		blobs, err := ctl.FindMissingAssociationsForProject(ctx, projectID, nil)
		suite.Nil(err)
		suite.Len(blobs, 0)
	}

	blobs := []*blob.Blob{{Digest: "1"}, {Digest: "2"}, {Digest: "3"}}

	{
		mock.OnAnything(blobMgr, "List").Return(nil, nil).Once()
		missing, err := ctl.FindMissingAssociationsForProject(ctx, projectID, blobs)
		suite.Nil(err)
		suite.Len(missing, len(blobs))
	}

	{
		mock.OnAnything(blobMgr, "List").Return(blobs, nil).Once()
		missing, err := ctl.FindMissingAssociationsForProject(ctx, projectID, blobs)
		suite.Nil(err)
		suite.Len(missing, 0)
	}

	{
		associated := []*blob.Blob{{Digest: "1"}}
		mock.OnAnything(blobMgr, "List").Return(associated, nil).Once()
		missing, err := ctl.FindMissingAssociationsForProject(ctx, projectID, blobs)
		suite.Nil(err)
		suite.Len(missing, len(blobs)-len(associated))
	}
}

func (suite *ControllerTestSuite) TestGet() {
	ctx := suite.Context()

	{
		digest := suite.prepareBlob()
		blob, err := Ctl.Get(ctx, digest)
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(int64(100), blob.Size)
		suite.Equal("application/octet-stream", blob.ContentType)
	}

	{
		digest := suite.prepareBlob()
		artifactDigest := suite.DigestString()

		_, err := Ctl.Get(ctx, digest, IsAssociatedWithArtifact(artifactDigest))
		suite.NotNil(err)

		Ctl.AssociateWithArtifact(ctx, []string{digest}, artifactDigest)

		blob, err := Ctl.Get(ctx, digest, IsAssociatedWithArtifact(artifactDigest))
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(int64(100), blob.Size)
		suite.Equal("application/octet-stream", blob.ContentType)
	}

	{
		digest := suite.prepareBlob()

		suite.WithProject(func(projectID int64, projectName string) {
			_, err := Ctl.Get(ctx, digest, IsAssociatedWithProject(projectID))
			suite.NotNil(err)

			Ctl.AssociateWithProjectByDigest(ctx, digest, projectID)

			blob, err := Ctl.Get(ctx, digest, IsAssociatedWithProject(projectID))
			suite.Nil(err)
			suite.Equal(digest, blob.Digest)
			suite.Equal(int64(100), blob.Size)
			suite.Equal("application/octet-stream", blob.ContentType)
		})
	}
}

func (suite *ControllerTestSuite) TestSync() {
	var references []distribution.Descriptor
	for i := 0; i < 5; i++ {
		references = append(references, distribution.Descriptor{
			MediaType: fmt.Sprintf("media type %d", i),
			Digest:    suite.Digest(),
			Size:      int64(100 + i),
		})
	}

	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		{
			suite.Nil(Ctl.Sync(ctx, references))
			for _, reference := range references {
				blob, err := Ctl.Get(ctx, reference.Digest.String())
				suite.Nil(err)
				suite.Equal(reference.MediaType, blob.ContentType)
				suite.Equal(reference.Digest.String(), blob.Digest)
				suite.Equal(reference.Size, blob.Size)
			}
		}

		{
			references[0].MediaType = "media type"

			references = append(references, distribution.Descriptor{
				MediaType: "media type",
				Digest:    suite.Digest(),
				Size:      int64(100),
			})

			suite.Nil(Ctl.Sync(ctx, references))
		}
	})
}

func (suite *ControllerTestSuite) TestGetSetAcceptedBlobSize() {
	{
		sessionID := uuid.New().String()

		size, err := Ctl.GetAcceptedBlobSize(sessionID)
		suite.Nil(err)
		suite.Equal(int64(0), size)

		suite.Nil(Ctl.SetAcceptedBlobSize(sessionID, 100))

		size, err = Ctl.GetAcceptedBlobSize(sessionID)
		suite.Nil(err)
		suite.Equal(int64(100), size)
	}

	{
		// test blob size expired
		ctl := &controller{blobSizeExpiration: time.Second * 5}

		sessionID := uuid.New().String()

		size, err := ctl.GetAcceptedBlobSize(sessionID)
		suite.Nil(err)
		suite.Equal(int64(0), size)

		suite.Nil(ctl.SetAcceptedBlobSize(sessionID, 100))

		size, err = ctl.GetAcceptedBlobSize(sessionID)
		suite.Nil(err)
		suite.Equal(int64(100), size)

		time.Sleep(time.Second * 10)

		size, err = ctl.GetAcceptedBlobSize(sessionID)
		suite.Nil(err)
		suite.Equal(int64(0), size)
	}
}

func (suite *ControllerTestSuite) TestTouch() {
	ctx := suite.Context()

	err := Ctl.Touch(ctx, &blob.Blob{
		Status: models.StatusNone,
	})
	suite.NotNil(err)
	suite.True(errors.IsNotFoundErr(err))

	digest := suite.prepareBlob()
	blob, err := Ctl.Get(ctx, digest)
	suite.Nil(err)
	blob.Status = models.StatusDelete
	_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), blob)
	suite.Nil(err)

	err = Ctl.Touch(ctx, blob)
	suite.Nil(err)
	suite.Equal(blob.Status, models.StatusNone)
}

func (suite *ControllerTestSuite) TestFail() {
	ctx := suite.Context()

	err := Ctl.Fail(ctx, &blob.Blob{
		Status: models.StatusNone,
	})
	suite.NotNil(err)
	suite.True(errors.IsNotFoundErr(err))

	digest := suite.prepareBlob()
	blob, err := Ctl.Get(ctx, digest)
	suite.Nil(err)

	blob.Status = models.StatusDelete
	_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), blob)
	suite.Nil(err)

	// StatusDelete cannot be marked as StatusDeleteFailed
	err = Ctl.Fail(ctx, blob)
	suite.NotNil(err)
	suite.True(errors.IsNotFoundErr(err))

	blob.Status = models.StatusDeleting
	_, err = pkg_blob.Mgr.UpdateBlobStatus(suite.Context(), blob)
	suite.Nil(err)

	err = Ctl.Fail(ctx, blob)
	suite.Nil(err)

	blobAfter, err := Ctl.Get(ctx, digest)
	suite.Nil(err)
	suite.Equal(models.StatusDeleteFailed, blobAfter.Status)
}

func (suite *ControllerTestSuite) TestDelete() {
	ctx := suite.Context()

	digest := suite.DigestString()
	_, err := Ctl.Ensure(ctx, digest, "application/octet-stream", 100)
	suite.Nil(err)

	blob, err := Ctl.Get(ctx, digest)
	suite.Nil(err)
	suite.Equal(digest, blob.Digest)

	err = Ctl.Delete(ctx, blob.ID)
	suite.Nil(err)

	exist, err := Ctl.Exist(ctx, digest)
	suite.Nil(err)
	suite.False(exist)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
