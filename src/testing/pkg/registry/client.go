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

package registry

import (
	"github.com/distribution/distribution"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

// FakeClient is a fake registry client that implement src/pkg/registry.Client interface
type FakeClient struct {
	mock.Mock
}

// Ping ...
func (f *FakeClient) Ping() (err error) {
	args := f.Called()
	return args.Error(0)
}

// Catalog ...
func (f *FakeClient) Catalog() ([]string, error) {
	args := f.Called()
	var repositories []string
	if args[0] != nil {
		repositories = args[0].([]string)
	}
	return repositories, args.Error(1)
}

// ListTags ...
func (f *FakeClient) ListTags(repository string) ([]string, error) {
	args := f.Called()
	var tags []string
	if args[0] != nil {
		tags = args[0].([]string)
	}
	return tags, args.Error(1)
}

// ManifestExist ...
func (f *FakeClient) ManifestExist(repository, reference string) (bool, *distribution.Descriptor, error) {
	args := f.Called()
	var desc *distribution.Descriptor
	if args[0] != nil {
		desc = args[0].(*distribution.Descriptor)
	}
	return args.Bool(0), desc, args.Error(2)
}

// PullManifest ...
func (f *FakeClient) PullManifest(repository, reference string, acceptedMediaTypes ...string) (distribution.Manifest, string, error) {
	args := f.Called()
	var manifest distribution.Manifest
	if args[0] != nil {
		manifest = args[0].(distribution.Manifest)
	}
	return manifest, args.String(1), args.Error(2)
}

// PushManifest ...
func (f *FakeClient) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}

// DeleteManifest ...
func (f *FakeClient) DeleteManifest(repository, reference string) error {
	args := f.Called()
	return args.Error(0)
}

// BlobExist ...
func (f *FakeClient) BlobExist(repository, digest string) (bool, error) {
	args := f.Called()
	return args.Bool(0), args.Error(1)
}

// PullBlob ...
func (f *FakeClient) PullBlob(repository, digest string) (int64, io.ReadCloser, error) {
	args := f.Called()
	var blob io.ReadCloser
	if args[1] != nil {
		blob = args[1].(io.ReadCloser)
	}
	return int64(args.Int(0)), blob, args.Error(2)
}

// PushBlob ...
func (f *FakeClient) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	args := f.Called()
	return args.Error(0)
}

// MountBlob ...
func (f *FakeClient) MountBlob(srcRepository, digest, dstRepository string) (err error) {
	args := f.Called()
	return args.Error(0)
}

// DeleteBlob ...
func (f *FakeClient) DeleteBlob(repository, digest string) (err error) {
	args := f.Called()
	return args.Error(0)
}

// Copy ...
func (f *FakeClient) Copy(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	args := f.Called()
	return args.Error(0)
}

func (f *FakeClient) Do(req *http.Request) (*http.Response, error) {
	args := f.Called()
	var resp *http.Response
	if args[0] != nil {
		resp = args[0].(*http.Response)
	}
	return resp, args.Error(1)
}
