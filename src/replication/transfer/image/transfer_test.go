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

package image

import (
	"bytes"
	"github.com/opencontainers/go-digest"
	"io"
	"io/ioutil"
	"testing"

	"github.com/distribution/distribution"
	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/model"
	trans "github.com/goharbor/harbor/src/replication/transfer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRegistry struct{}

func (f *fakeRegistry) FetchArtifacts([]*model.Filter) ([]*model.Resource, error) {
	return nil, nil
}

func (f *fakeRegistry) ManifestExist(repository, reference string) (bool, *distribution.Descriptor, error) {
	if repository == "destination" && reference == "b1" {
		return true, &distribution.Descriptor{Digest: digest.Digest("sha256:c6b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7")}, nil
	}
	return false, &distribution.Descriptor{Digest: digest.Digest("sha256:c6b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7")}, nil
}
func (f *fakeRegistry) PullManifest(repository, reference string, accepttedMediaTypes ...string) (distribution.Manifest, string, error) {
	manifest := `{
		"schemaVersion": 2,
		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
		"config": {
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size": 7023,
			"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
		},
		"layers": [
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 32654,
				"digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
			},
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 16724,
				"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
			},
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size": 73109,
				"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
			}
		]
	}`
	mediaType := schema2.MediaTypeManifest
	payload := []byte(manifest)
	mani, _, err := distribution.UnmarshalManifest(mediaType, payload)
	if err != nil {
		return nil, "", err
	}
	return mani, "sha256:c6b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7", nil
}
func (f *fakeRegistry) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	return "", nil
}
func (f *fakeRegistry) DeleteManifest(repository, reference string) error {
	return nil
}
func (f *fakeRegistry) BlobExist(repository, digest string) (bool, error) {
	return false, nil
}
func (f *fakeRegistry) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte{'a'}))
	return 1, r, nil
}
func (f *fakeRegistry) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return nil
}
func (f *fakeRegistry) DeleteTag(repository, tag string) error {
	return nil
}

func TestFactory(t *testing.T) {
	tr, err := factory(nil, nil)
	require.Nil(t, err)
	_, ok := tr.(trans.Transfer)
	assert.True(t, ok)
}

func TestShouldStop(t *testing.T) {
	// should stop
	stopFunc := func() bool { return true }
	tr := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
	}
	assert.True(t, tr.shouldStop())

	// should not stop
	stopFunc = func() bool { return false }
	tr = &transfer{
		isStopped: stopFunc,
	}
	assert.False(t, tr.shouldStop())
}

func TestCopy(t *testing.T) {
	stopFunc := func() bool { return false }
	tr := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
		src:       &fakeRegistry{},
		dst:       &fakeRegistry{},
	}

	src := &repository{
		repository: "source",
		tags:       []string{"a1", "a2"},
	}
	dst := &repository{
		repository: "destination",
		tags:       []string{"b1", "b2"},
	}
	override := true
	err := tr.copy(src, dst, override)
	require.Nil(t, err)
}

func TestDelete(t *testing.T) {
	stopFunc := func() bool { return false }
	tr := &transfer{
		logger:    log.DefaultLogger(),
		isStopped: stopFunc,
		dst:       &fakeRegistry{},
	}

	repo := &repository{
		repository: "destination",
		tags:       []string{"b1", "b2"},
	}
	err := tr.delete(repo)
	require.Nil(t, err)
}
