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
	"encoding/json"
	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

type clientTestSuite struct {
	suite.Suite
	client Client
}

func (c *clientTestSuite) TestPing() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: test.Handler(nil),
		})
	defer server.Close()

	err := NewClient(server.URL, "", "", true).Ping()
	c.Require().Nil(err)
}

func (c *clientTestSuite) TestCatalog() {
	type repositories struct {
		Repositories []string `json:"repositories"`
	}

	isFirstRequest := true

	handler := func(w http.ResponseWriter, r *http.Request) {
		if isFirstRequest {
			isFirstRequest = false

			repos := &repositories{
				Repositories: []string{"library/alpine"},
			}
			link := lib.Link{
				URL: `/v2/_catalog?last=library/alpine`,
				Rel: "next",
			}
			w.Header().Set(http.CanonicalHeaderKey("link"), link.String())
			encoder := json.NewEncoder(w)
			err := encoder.Encode(repos)
			c.Require().Nil(err)
			return
		}

		if r.URL.String() != "/v2/_catalog?last=library/alpine" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		repos := &repositories{
			Repositories: []string{"library/hello-world"},
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(repos)
		c.Require().Nil(err)
		return
	}

	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/_catalog",
			Handler: handler,
		})
	defer server.Close()

	repos, err := NewClient(server.URL, "", "", true).Catalog()
	c.Require().Nil(err)
	c.Len(repos, 2)
	c.EqualValues([]string{"library/alpine", "library/hello-world"}, repos)
}

func (c *clientTestSuite) TestListTags() {
	type tags struct {
		Tags []string `json:"tags"`
	}

	isFirstRequest := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		if isFirstRequest {
			isFirstRequest = false

			tgs := &tags{
				Tags: []string{"1.0"},
			}
			link := lib.Link{
				URL: `/v2/library/hello-world/tags/list?last=1.0`,
				Rel: "next",
			}
			w.Header().Set(http.CanonicalHeaderKey("link"), link.String())
			encoder := json.NewEncoder(w)
			err := encoder.Encode(tgs)
			c.Require().Nil(err)
			return
		}

		if r.URL.String() != "/v2/library/hello-world/tags/list?last=1.0" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tgs := &tags{
			Tags: []string{"2.0"},
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(tgs)
		c.Require().Nil(err)
		return
	}

	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/library/hello-world/tags/list",
			Handler: handler,
		})
	defer server.Close()

	repos, err := NewClient(server.URL, "", "", true).ListTags("library/hello-world")
	c.Require().Nil(err)
	c.Len(repos, 2)
	c.EqualValues([]string{"1.0", "2.0"}, repos)
}

func (c *clientTestSuite) TestManifestExist() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "HEAD",
			Pattern: "/v2/library/alpine/manifests/latest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusNotFound,
			}),
		},
		&test.RequestHandlerMapping{
			Method:  "HEAD",
			Pattern: "/v2/library/hello-world/manifests/latest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusOK,
				Headers: map[string]string{
					"Docker-Content-Digest": "digest",
				},
			}),
		},
	)
	defer server.Close()

	client := NewClient(server.URL, "", "", true)
	// doesn't exist
	exist, desc, err := client.ManifestExist("library/alpine", "latest")
	c.Require().Nil(err)
	c.False(exist)

	// exist
	exist, desc, err = client.ManifestExist("library/hello-world", "latest")
	c.Require().Nil(err)
	c.True(exist)
	c.Equal("digest", string(desc.Digest))
}

func (c *clientTestSuite) TestPullManifest() {
	data := `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1510,
      "digest": "sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 977,
         "digest": "sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced"
      }
   ]
}`
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/library/hello-world/manifests/latest",
			Handler: test.Handler(&test.Response{
				Headers: map[string]string{
					"Docker-Content-Digest": "digest",
					"Content-Type":          schema2.MediaTypeManifest,
				},
				Body: []byte(data),
			}),
		})
	defer server.Close()

	manifest, digest, err := NewClient(server.URL, "", "", true).PullManifest("library/hello-world", "latest")
	c.Require().Nil(err)
	c.Equal("digest", digest)

	mediaType, payload, err := manifest.Payload()
	c.Require().Nil(err)
	c.Equal(schema2.MediaTypeManifest, mediaType)
	c.Equal(data, string(payload))
}

func (c *clientTestSuite) TestPushManifest() {

	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "PUT",
			Pattern: "/v2/library/hello-world/manifests/latest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusCreated,
				Headers: map[string]string{
					"Docker-Content-Digest": "digest",
				},
			}),
		})
	defer server.Close()

	digest, err := NewClient(server.URL, "", "", true).PushManifest("library/hello-world", "latest", "", nil)
	c.Require().Nil(err)
	c.Equal("digest", digest)
}

func (c *clientTestSuite) TestDeleteManifest() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "HEAD",
			Pattern: "/v2/library/hello-world/manifests/latest",
			Handler: test.Handler(&test.Response{
				Headers: map[string]string{
					"Docker-Content-Digest": "digest",
				},
			}),
		},
		&test.RequestHandlerMapping{
			Method:  "DELETE",
			Pattern: "/v2/library/hello-world/manifests/digest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusAccepted,
			}),
		})
	defer server.Close()

	err := NewClient(server.URL, "", "", true).DeleteManifest("library/hello-world", "latest")
	c.Require().Nil(err)
}

func (c *clientTestSuite) TestBlobExist() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "HEAD",
			Pattern: "/v2/library/hello-world/blobs/digest1",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusNotFound,
			}),
		},
		&test.RequestHandlerMapping{
			Method:  "HEAD",
			Pattern: "/v2/library/hello-world/blobs/digest2",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusOK,
			}),
		},
	)
	defer server.Close()

	// doesn't exist
	client := NewClient(server.URL, "", "", true)
	exist, err := client.BlobExist("library/hello-world", "digest1")
	c.Require().Nil(err)
	c.False(exist)

	// exist
	exist, err = client.BlobExist("library/hello-world", "digest2")
	c.Require().Nil(err)
	c.True(exist)
}

func (c *clientTestSuite) TestPullBlob() {
	data := []byte{'a'}
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/library/hello-world/blobs/digest",
			Handler: test.Handler(&test.Response{
				Headers: map[string]string{
					"Content-Length": strconv.Itoa(len(data)),
				},
				Body: data,
			}),
		})
	defer server.Close()

	size, blob, err := NewClient(server.URL, "", "", true).PullBlob("library/hello-world", "digest")
	c.Require().Nil(err)
	c.Equal(int64(len(data)), size)

	b, err := ioutil.ReadAll(blob)
	c.Require().Nil(err)
	c.EqualValues(data, b)
}

func (c *clientTestSuite) TestPushBlob() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "POST",
			Pattern: "/v2/library/hello-world/blobs/uploads/",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusAccepted,
				Headers: map[string]string{
					"Location": "/v2/library/hello-world/blobs/uploads/uuid",
				},
			}),
		},
		&test.RequestHandlerMapping{
			Method:  "PUT",
			Pattern: "/v2/library/hello-world/blobs/uploads/uuid",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusCreated,
			}),
		})
	defer server.Close()

	err := NewClient(server.URL, "", "", true).PushBlob("library/hello-world", "digest", 0, nil)
	c.Require().Nil(err)
}

func (c *clientTestSuite) TestDeleteBlob() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "DELETE",
			Pattern: "/v2/library/hello-world/blobs/digest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusAccepted,
			}),
		})
	defer server.Close()

	err := NewClient(server.URL, "", "", true).DeleteBlob("library/hello-world", "digest")
	c.Require().Nil(err)
}

func (c *clientTestSuite) TestMountBlob() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "POST",
			Pattern: "/v2/library/hello-world/blobs/uploads/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				mount := r.URL.Query().Get("mount")
				from := r.URL.Query().Get("from")
				if mount != "digest" || from != "library/alpine" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusAccepted)
			},
		})
	defer server.Close()

	err := NewClient(server.URL, "", "", true).MountBlob("library/alpine", "digest", "library/hello-world")
	c.Require().Nil(err)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, &clientTestSuite{})
}
