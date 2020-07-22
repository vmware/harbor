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
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type localInterfaceMock struct {
	mock.Mock
}

func (l *localInterfaceMock) BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error) {
	args := l.Called(ctx, art)
	return args.Bool(0), args.Error(1)
}

func (l *localInterfaceMock) ManifestExist(ctx context.Context, art lib.ArtifactInfo) bool {
	args := l.Called(ctx, art)
	return args.Bool(0)
}

func (l *localInterfaceMock) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	panic("implement me")
}

func (l *localInterfaceMock) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	panic("implement me")
}

func (l *localInterfaceMock) DeleteManifest(repo, ref string) {
	panic("implement me")
}

type proxyControllerTestSuite struct {
	suite.Suite
	local *localInterfaceMock
	ctr   Controller
}

func (p *proxyControllerTestSuite) SetupTest() {
	p.local = &localInterfaceMock{}
	p.ctr = &controller{
		blobCtl:     blob.Ctl,
		registryMgr: registry.NewDefaultManager(),
		artifactCtl: artifact.Ctl,
		local:       p.local,
	}
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("ManifestExist", mock.Anything, mock.Anything).Return(true, nil)
	result := p.ctr.UseLocalManifest(ctx, art)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("ManifestExist", mock.Anything, mock.Anything).Return(false, nil)
	result := p.ctr.UseLocalManifest(ctx, art)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_False() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	p.local.On("ManifestExist", mock.Anything, mock.Anything).Return(true, nil)
	result := p.ctr.UseLocalManifest(ctx, art)
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
