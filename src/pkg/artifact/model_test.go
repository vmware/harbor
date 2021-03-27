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
	"testing"
	"time"

	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type modelTestSuite struct {
	suite.Suite
}

func (m *modelTestSuite) TestArtifactFrom() {
	t := m.T()
	dbArt := &dao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	art := &Artifact{}
	art.From(dbArt)
	assert.Equal(t, dbArt.ID, art.ID)
	assert.Equal(t, dbArt.Type, art.Type)
	assert.Equal(t, dbArt.MediaType, art.MediaType)
	assert.Equal(t, dbArt.ManifestMediaType, art.ManifestMediaType)
	assert.Equal(t, dbArt.ProjectID, art.ProjectID)
	assert.Equal(t, dbArt.RepositoryID, art.RepositoryID)
	assert.Equal(t, dbArt.Digest, art.Digest)
	assert.Equal(t, dbArt.Size, art.Size)
	assert.Equal(t, dbArt.PushTime, art.PushTime)
	assert.Equal(t, dbArt.PullTime, art.PullTime)
	assert.Equal(t, "value1", art.ExtraAttrs["attr1"].(string))
	assert.Equal(t, "value1", art.Annotations["anno1"])
}

func (m *modelTestSuite) TestArtifactTo() {
	t := m.T()
	art := &Artifact{
		ID:                1,
		Type:              "IMAGE",
		ProjectID:         1,
		RepositoryID:      1,
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs: map[string]interface{}{
			"attr1": "value1",
		},
		Annotations: map[string]string{
			"anno1": "value1",
		},
	}
	dbArt := art.To()
	assert.Equal(t, art.ID, dbArt.ID)
	assert.Equal(t, art.Type, dbArt.Type)
	assert.Equal(t, art.MediaType, dbArt.MediaType)
	assert.Equal(t, art.ManifestMediaType, dbArt.ManifestMediaType)
	assert.Equal(t, art.ProjectID, dbArt.ProjectID)
	assert.Equal(t, art.RepositoryID, dbArt.RepositoryID)
	assert.Equal(t, art.Digest, dbArt.Digest)
	assert.Equal(t, art.Size, dbArt.Size)
	assert.Equal(t, art.PushTime, dbArt.PushTime)
	assert.Equal(t, art.PullTime, dbArt.PullTime)
	assert.Equal(t, `{"attr1":"value1"}`, dbArt.ExtraAttrs)
	assert.Equal(t, `{"anno1":"value1"}`, dbArt.Annotations)
}

func (m *modelTestSuite) TestIsImageIndex() {
	art1 := Artifact{ManifestMediaType: v1.MediaTypeImageIndex}
	m.True(art1.IsImageIndex())

	art2 := Artifact{ManifestMediaType: manifestlist.MediaTypeManifestList}
	m.True(art2.IsImageIndex())

	art3 := Artifact{ManifestMediaType: v1.MediaTypeImageManifest}
	m.False(art3.IsImageIndex())
}

func TestModel(t *testing.T) {
	suite.Run(t, &modelTestSuite{})
}
