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

package distribution

import (
	"testing"

	_ "github.com/distribution/distribution/manifest/manifestlist"
	_ "github.com/distribution/distribution/manifest/ocischema"
	_ "github.com/distribution/distribution/manifest/schema1"
	_ "github.com/distribution/distribution/manifest/schema2"
)

func TestParseSessionID(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"base", args{"/v2"}, ""},
		{"tags", args{"/v2/library/photon/tags/list"}, ""},
		{"manifest", args{"/v2/library/photon/manifests/2.0"}, ""},
		{"blob", args{"/v2/library/photon/blobs/sha256:c52fca2e807cb7807cfd831d6df45a332d5826a97f886f7da0e9c61842f9ce1e"}, ""},
		{"initiate blob upload", args{"/v2/library/photon/blobs/uploads"}, ""},
		{"blob upload", args{"/v2/library/photon/blobs/uploads/aa41e8cb-21b4-423c-b533-9e4b084075c7"}, "aa41e8cb-21b4-423c-b533-9e4b084075c7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSessionID(tt.args.path); got != tt.want {
				t.Errorf("ParseSessionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"base", args{"/v2"}, ""},
		{"tags", args{"/v2/library/photon/tags/list"}, "library/photon"},
		{"manifest", args{"/v2/library/photon/manifests/2.0"}, "library/photon"},
		{"blob", args{"/v2/library/photon/blobs/sha256:c52fca2e807cb7807cfd831d6df45a332d5826a97f886f7da0e9c61842f9ce1e"}, "library/photon"},
		{"initiate blob upload", args{"/v2/library/photon/blobs/uploads"}, "library/photon"},
		{"blob upload", args{"/v2/library/photon/blobs/uploads/aa41e8cb-21b4-423c-b533-9e4b084075c7"}, "library/photon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseName(tt.args.path); got != tt.want {
				t.Errorf("ParseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseProjectName(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"base", args{"/v2"}, ""},
		{"tags", args{"/v2/library/photon/tags/list"}, "library"},
		{"manifest", args{"/v2/library/photon/manifests/2.0"}, "library"},
		{"blob", args{"/v2/library/photon/blobs/sha256:c52fca2e807cb7807cfd831d6df45a332d5826a97f886f7da0e9c61842f9ce1e"}, "library"},
		{"initiate blob upload", args{"/v2/library/photon/blobs/uploads"}, "library"},
		{"blob upload", args{"/v2/library/photon/blobs/uploads/aa41e8cb-21b4-423c-b533-9e4b084075c7"}, "library"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseProjectName(tt.args.path); got != tt.want {
				t.Errorf("ParseProjectName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReference(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tag", args{"/v2/library/photon/manifests/2.0"}, "2.0"},
		{"digest", args{"/v2/library/photon/manifests/sha256:c52fca2e807cb7807cfd831d6df45a332d5826a97f886f7da0e9c61842f9ce1e"}, "sha256:c52fca2e807cb7807cfd831d6df45a332d5826a97f886f7da0e9c61842f9ce1e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseReference(tt.args.path); got != tt.want {
				t.Errorf("ParseReference() = %v, want %v", got, tt.want)
			}
		})
	}
}
