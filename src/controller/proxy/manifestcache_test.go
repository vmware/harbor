//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"context"
	"fmt"
	"github.com/distribution/distribution"
	"github.com/distribution/distribution/manifest"
	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CacheTestSuite struct {
	suite.Suite
	mHandler *ManifestListCache
	local    localInterfaceMock
}

func (suite *CacheTestSuite) SetupSuite() {
	suite.local = localInterfaceMock{}
	suite.mHandler = &ManifestListCache{local: &suite.local}
}

func (suite *CacheTestSuite) TearDownSuite() {
}
func (suite *CacheTestSuite) TestUpdateManifestList() {
	ctx := context.Background()
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestList := manifestlist.ManifestList{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     manifestlist.MediaTypeManifestList,
		},
		Manifests: []manifestlist.ManifestDescriptor{
			{
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(amdDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "amd64",
					OS:           "linux",
				},
			}, {
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(armDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "arm",
					OS:           "linux",
				},
			},
		},
	}
	manList := &manifestlist.DeserializedManifestList{
		ManifestList: manifestList,
	}
	artInfo1 := lib.ArtifactInfo{
		Repository: "library/hello-world",
		Digest:     amdDig,
		Tag:        "",
	}
	ar := &artifact.Artifact{}
	suite.local.On("GetManifest", ctx, artInfo1).Return(ar, nil)
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	newMan, err := suite.mHandler.updateManifestList(ctx, "library/hello-world", manList)
	suite.Require().Nil(err)
	suite.Assert().Equal(len(newMan.References()), 1)
}

func (suite *CacheTestSuite) TestPushManifestList() {
	ctx := context.Background()
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestList := manifestlist.ManifestList{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     manifestlist.MediaTypeManifestList,
		},
		Manifests: []manifestlist.ManifestDescriptor{
			{
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(amdDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "amd64",
					OS:           "linux",
				},
			}, {
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(armDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "arm",
					OS:           "linux",
				},
			},
		},
	}
	manList := &manifestlist.DeserializedManifestList{
		ManifestList: manifestList,
	}
	repo := "library/hello-world"
	artInfo1 := lib.ArtifactInfo{
		Repository: repo,
		Digest:     amdDig,
		Tag:        "",
	}
	ar := &artifact.Artifact{}
	_, payload, err := manList.Payload()
	suite.Nil(err)
	originDigest := digest.FromBytes(payload)

	suite.local.On("GetManifest", ctx, artInfo1).Return(ar, nil)
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	suite.local.On("PushManifest", repo, originDigest, mock.Anything).Return(fmt.Errorf("wrong digest"))
	suite.local.On("PushManifest", repo, mock.Anything, mock.Anything).Return(nil)

	err = suite.mHandler.push(ctx, "library/hello-world", string(originDigest), manList)
	suite.Require().Nil(err)
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, &CacheTestSuite{})
}
