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
	distribution2 "github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"testing"

	testregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
	"github.com/stretchr/testify/suite"
)

type mockManifest struct {
	mock.Mock
}

func (m *mockManifest) References() []distribution2.Descriptor {
	args := m.Called()
	desc := make([]distribution2.Descriptor, 0)
	if args[0] != nil {
		desc = args[0].([]distribution2.Descriptor)
	}
	return desc
}

func (m *mockManifest) Payload() (mediaType string, payload []byte, err error) {
	args := m.Called()
	p := make([]byte, 0)
	if args[1] != nil {
		p = args[1].([]byte)
	}
	return args.String(0), p, args.Error(2)
}

type artifactControllerMock struct {
	mock.Mock
}

func (a *artifactControllerMock) GetByReference(ctx context.Context, repository, reference string, option *artifact.Option) (arti *artifact.Artifact, err error) {
	args := a.Called(ctx, repository, reference, option)
	art := &artifact.Artifact{}
	if args[0] != nil {
		art = args[0].(*artifact.Artifact)
	}
	return art, args.Error(1)
}

type localHelperTestSuite struct {
	suite.Suite
	registryClient *testregistry.FakeClient
	local          *localHelper
	artCtl         *artifactControllerMock
}

func (lh *localHelperTestSuite) SetupTest() {
	lh.registryClient = &testregistry.FakeClient{}
	lh.artCtl = &artifactControllerMock{}
	lh.local = &localHelper{registry: lh.registryClient, artifactCtl: lh.artCtl}

}

func (lh *localHelperTestSuite) TestBlobExist_False() {
	repo := "library/hello-world"
	dig := "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}
	ctx := context.Background()
	lh.registryClient.On("BlobExist").Return(false, nil)
	exist, err := lh.local.BlobExist(ctx, art)
	lh.Require().Nil(err)
	lh.Assert().Equal(false, exist)
}
func (lh *localHelperTestSuite) TestBlobExist_True() {
	repo := "library/hello-world"
	dig := "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}
	ctx := context.Background()
	lh.registryClient.On("BlobExist").Return(true, nil)
	exist, err := lh.local.BlobExist(ctx, art)
	lh.Require().Nil(err)
	lh.Assert().Equal(true, exist)
}

func (lh *localHelperTestSuite) TestPushManifest() {
	dig := "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	lh.registryClient.On("PushManifest").Return(dig, nil)
	manifest := &mockManifest{}
	var ct string
	manifest.Mock.On("Payload").Return(ct, []byte("example"), nil)
	ct = schema2.MediaTypeManifest
	err := lh.local.PushManifest("library/hello-world", "", manifest)
	lh.Require().Nil(err)
}

func (lh *localHelperTestSuite) TestUpdateManifestList() {
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
	ar := &artifact.Artifact{}
	var emptyArtifact *artifact.Artifact
	var opt *artifact.Option
	lh.artCtl.On("GetByReference", ctx, "library/hello-world", amdDig, opt).Return(ar, nil)
	lh.artCtl.On("GetByReference", ctx, "library/hello-world", armDig, opt).Return(emptyArtifact, nil)
	newMan, err := lh.local.updateManifestList(ctx, "library/hello-world", manList)
	lh.Require().Nil(err)
	lh.Assert().Equal(len(newMan.References()), 1)
}

func (lh *localHelperTestSuite) TestCheckDependencies_Fail() {
	ctx := context.Background()
	manifest := &mockManifest{}
	refs := []distribution2.Descriptor{
		{Digest: "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"},
		{Digest: "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"},
	}
	manifest.On("References").Return(refs)
	lh.registryClient.On("BlobExist").Return(false, nil)
	ret := lh.local.CheckDependencies(ctx, "library/hello-world", manifest)
	lh.Assert().Equal(len(ret), 2)
}

func (lh *localHelperTestSuite) TestCheckDependencies_Suc() {
	ctx := context.Background()
	manifest := &mockManifest{}
	refs := []distribution2.Descriptor{
		{Digest: "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"},
		{Digest: "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"},
	}
	manifest.On("References").Return(refs)
	lh.registryClient.On("BlobExist").Return(true, nil)
	ret := lh.local.CheckDependencies(ctx, "library/hello-world", manifest)
	lh.Assert().Equal(len(ret), 0)
}

func (lh *localHelperTestSuite) TestManifestExist() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	ar := &artifact.Artifact{}
	var opt *artifact.Option
	lh.artCtl.On("GetByReference", ctx, "library/hello-world", dig, opt).Return(ar, nil)
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	exist := lh.local.ManifestExist(ctx, art)
	lh.Assert().True(exist)
}

func TestLocalHelperTestSuite(t *testing.T) {
	suite.Run(t, &localHelperTestSuite{})
}
