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
	"github.com/distribution/distribution"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	_ "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type localInterfaceMock struct {
	mock.Mock
}

func (l *localInterfaceMock) SendPullEvent(ctx context.Context, repo, tag string) {
	panic("implement me")
}

func (l *localInterfaceMock) GetManifest(ctx context.Context, art lib.ArtifactInfo) (*artifact.Artifact, error) {
	args := l.Called(ctx, art)

	var a *artifact.Artifact
	if args.Get(0) != nil {
		a = args.Get(0).(*artifact.Artifact)
	}
	return a, args.Error(1)
}

func (l *localInterfaceMock) SameArtifact(ctx context.Context, repo, tag, dig string) (bool, error) {
	panic("implement me")
}

func (l *localInterfaceMock) BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error) {
	args := l.Called(ctx, art)
	return args.Bool(0), args.Error(1)
}

func (l *localInterfaceMock) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	panic("implement me")
}

func (l *localInterfaceMock) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	args := l.Called(repo, tag, manifest)
	return args.Error(0)
}

func (l *localInterfaceMock) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	panic("implement me")
}

func (l *localInterfaceMock) DeleteManifest(repo, ref string) {
}

type proxyControllerTestSuite struct {
	suite.Suite
	local  *localInterfaceMock
	remote *testproxy.RemoteInterface
	ctr    Controller
	proj   *models.Project
}

func (p *proxyControllerTestSuite) SetupTest() {
	p.local = &localInterfaceMock{}
	p.remote = &testproxy.RemoteInterface{}
	p.proj = &models.Project{RegistryID: 1}
	p.ctr = &controller{
		blobCtl:     blob.Ctl,
		artifactCtl: artifact.Ctl,
		local:       p.local,
	}
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)

	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_False() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	desc := &distribution.Descriptor{}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().True(errors.IsNotFoundErr(err))
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(false, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().False(result)
}

func TestProxyControllerTestSuite(t *testing.T) {
	suite.Run(t, &proxyControllerTestSuite{})
}

func TestProxyCacheRemoteRepo(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal test`,
			in:   lib.ArtifactInfo{ProjectName: "dockerhub_proxy", Repository: "dockerhub_proxy/firstfloor/hello-world"},
			want: "firstfloor/hello-world",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := getRemoteRepo(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
func TestGetRef(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "latest", Digest: "sha256:aabbcc"},
			want: "latest",
		},
		{
			name: `digest_only`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "", Digest: "sha256:aabbcc"},
			want: "sha256:aabbcc",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := getReference(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
