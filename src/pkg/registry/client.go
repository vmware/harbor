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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	_ "github.com/docker/distribution/manifest/ocischema" // register oci manifest unmarshal function
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry/auth"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// Cli is the global registry client instance, it targets to the backend docker registry
	Cli = func() Client {
		url, _ := config.RegistryURL()
		username, password := config.RegistryCredential()
		return NewClient(url, username, password, false)
	}()

	accepts = []string{
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageManifest,
		schema2.MediaTypeManifest,
		schema1.MediaTypeSignedManifest,
		schema1.MediaTypeManifest,
	}
)

// const definition
const (
	UserAgent = "harbor-registry-client"
)

// Client defines the methods that a registry client should implements
type Client interface {
	// Ping the base API endpoint "/v2/"
	Ping() (err error)
	// Catalog the repositories
	Catalog() (repositories []string, err error)
	// ListTags lists the tags under the specified repository
	ListTags(repository string) (tags []string, err error)
	// ManifestExist checks the existence of the manifest
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	// PullManifest pulls the specified manifest
	PullManifest(repository, reference string, acceptedMediaTypes ...string) (manifest distribution.Manifest, digest string, err error)
	// PushManifest pushes the specified manifest
	PushManifest(repository, reference, mediaType string, payload []byte) (digest string, err error)
	// DeleteManifest deletes the specified manifest. The "reference" can be "tag" or "digest"
	DeleteManifest(repository, reference string) (err error)
	// BlobExist checks the existence of the specified blob
	BlobExist(repository, digest string) (exist bool, err error)
	// PullBlob pulls the specified blob. The caller must close the returned "blob"
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	// PushBlob pushes the specified blob
	PushBlob(repository, digest string, size int64, blob io.Reader) error
	// MountBlob mounts the blob from the source repository
	MountBlob(srcRepository, digest, dstRepository string) (err error)
	// DeleteBlob deletes the specified blob
	DeleteBlob(repository, digest string) (err error)
	// Copy the artifact from source repository to the destination. The "override"
	// is used to specify whether the destination artifact will be overridden if
	// its name is same with source but digest isn't
	Copy(srcRepository, srcReference, dstRepository, dstReference string, override bool) (err error)
}

// NewClient creates a registry client with the default authorizer which determines the auth scheme
// of the registry automatically and calls the corresponding underlying authorizers(basic/bearer) to
// do the auth work. If a customized authorizer is needed, use "NewClientWithAuthorizer" instead
func NewClient(url, username, password string, insecure bool) Client {
	return &client{
		url:        url,
		authorizer: auth.NewAuthorizer(username, password, insecure),
		client: &http.Client{
			Transport: commonhttp.GetHTTPTransportByInsecure(insecure),
			Timeout:   30 * time.Minute,
		},
	}
}

// NewClientWithAuthorizer creates a registry client with the provided authorizer
func NewClientWithAuthorizer(url string, authorizer lib.Authorizer, insecure bool) Client {
	var transportType uint
	if insecure {
		transportType = commonhttp.InsecureTransport
	} else {
		transportType = commonhttp.SecureTransport
	}

	return &client{
		url:        url,
		authorizer: authorizer,
		client: &http.Client{
			Transport: commonhttp.GetHTTPTransport(transportType),
		},
	}
}

type client struct {
	url        string
	authorizer lib.Authorizer
	client     *http.Client
}

func (c *client) Ping() error {
	req, err := http.NewRequest(http.MethodGet, buildPingURL(c.url), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) Catalog() ([]string, error) {
	var repositories []string
	url := buildCatalogURL(c.url)
	for {
		repos, next, err := c.catalog(url)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repos...)

		url = next
		// no next page, end the loop
		if len(url) == 0 {
			break
		}
		// relative URL
		if !strings.Contains(url, "://") {
			url = c.url + url
		}
	}
	return repositories, nil
}

func (c *client) catalog(url string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	repositories := struct {
		Repositories []string `json:"repositories"`
	}{}
	if err := json.Unmarshal(body, &repositories); err != nil {
		return nil, "", err
	}
	return repositories.Repositories, next(resp.Header.Get("Link")), nil
}

func (c *client) ListTags(repository string) ([]string, error) {
	var tags []string
	url := buildTagListURL(c.url, repository)
	for {
		tgs, next, err := c.listTags(url)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tgs...)

		url = next
		// no next page, end the loop
		if len(url) == 0 {
			break
		}
		// relative URL
		if !strings.Contains(url, "://") {
			url = c.url + url
		}
	}
	return tags, nil
}

func (c *client) listTags(url string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	tgs := struct {
		Tags []string `json:"tags"`
	}{}
	if err := json.Unmarshal(body, &tgs); err != nil {
		return nil, "", err
	}
	return tgs.Tags, next(resp.Header.Get("Link")), nil
}

func (c *client) ManifestExist(repository, reference string) (bool, string, error) {
	req, err := http.NewRequest(http.MethodHead, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return false, "", err
	}
	for _, mediaType := range accepts {
		req.Header.Add(http.CanonicalHeaderKey("Accept"), mediaType)
	}
	resp, err := c.do(req)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return false, "", nil
		}
		return false, "", err
	}
	defer resp.Body.Close()
	return true, resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest")), nil
}

func (c *client) PullManifest(repository, reference string, acceptedMediaTypes ...string) (
	distribution.Manifest, string, error) {
	req, err := http.NewRequest(http.MethodGet, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return nil, "", err
	}
	if len(acceptedMediaTypes) == 0 {
		acceptedMediaTypes = accepts
	}
	for _, mediaType := range acceptedMediaTypes {
		req.Header.Add(http.CanonicalHeaderKey("Accept"), mediaType)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mediaType := resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
	manifest, _, err := distribution.UnmarshalManifest(mediaType, payload)
	if err != nil {
		return nil, "", err
	}
	digest := resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
	return manifest, digest, nil
}

func (c *client) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	req, err := http.NewRequest(http.MethodPut, buildManifestURL(c.url, repository, reference),
		bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set(http.CanonicalHeaderKey("Content-Type"), mediaType)
	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest")), nil
}

func (c *client) DeleteManifest(repository, reference string) error {
	_, err := digest.Parse(reference)
	if err != nil {
		// the reference is tag, get the digest first
		exist, digest, err := c.ManifestExist(repository, reference)
		if err != nil {
			return err
		}
		if !exist {
			return errors.New(nil).WithCode(errors.NotFoundCode).
				WithMessage("%s:%s not found", repository, reference)
		}
		reference = digest
	}
	req, err := http.NewRequest(http.MethodDelete, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) BlobExist(repository, digest string) (bool, error) {
	req, err := http.NewRequest(http.MethodHead, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return false, err
	}
	resp, err := c.do(req)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	return true, nil
}

func (c *client) PullBlob(repository, digest string) (int64, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Add(http.CanonicalHeaderKey("Accept-Encoding"), "identity")
	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}

	var size int64
	n := resp.Header.Get(http.CanonicalHeaderKey("Content-Length"))
	// no content-length is acceptable, which can taken from manifests
	if len(n) > 0 {
		size, err = strconv.ParseInt(n, 10, 64)
		if err != nil {
			defer resp.Body.Close()
			return 0, nil, err
		}
	}

	return size, resp.Body, nil
}

func (c *client) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	location, _, err := c.initiateBlobUpload(repository)
	if err != nil {
		return err
	}
	return c.monolithicBlobUpload(location, digest, size, blob)
}

func (c *client) initiateBlobUpload(repository string) (string, string, error) {
	req, err := http.NewRequest(http.MethodPost, buildInitiateBlobUploadURL(c.url, repository), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set(http.CanonicalHeaderKey("Content-Length"), "0")
	resp, err := c.do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	return resp.Header.Get(http.CanonicalHeaderKey("Location")),
		resp.Header.Get(http.CanonicalHeaderKey("Docker-Upload-UUID")), nil
}

func (c *client) monolithicBlobUpload(location, digest string, size int64, data io.Reader) error {
	url, err := buildMonolithicBlobUploadURL(c.url, location, digest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return err
	}
	req.ContentLength = size
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) MountBlob(srcRepository, digest, dstRepository string) error {
	req, err := http.NewRequest(http.MethodPost, buildMountBlobURL(c.url, dstRepository, digest, srcRepository), nil)
	if err != nil {
		return err
	}
	req.Header.Set(http.CanonicalHeaderKey("Content-Length"), "0")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) DeleteBlob(repository, digest string) error {
	req, err := http.NewRequest(http.MethodDelete, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) Copy(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	// pull the manifest from the source repository
	manifest, srcDgt, err := c.PullManifest(srcRepo, srcRef)
	if err != nil {
		return err
	}

	// check the existence of the artifact on the destination repository
	exist, dstDgt, err := c.ManifestExist(dstRepo, dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same artifact already exists
		if srcDgt == dstDgt {
			return nil
		}
		// the same name artifact exists, but not allowed to override
		if !override {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessage("the same name but different digest artifact exists, but the override is set to false")
		}
	}

	for _, descriptor := range manifest.References() {
		digest := descriptor.Digest.String()
		switch descriptor.MediaType {
		// skip foreign layer
		case schema2.MediaTypeForeignLayer:
			continue
		// manifest or index
		case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
			v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
			schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest:
			if err = c.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
				return err
			}
		// common layer
		default:
			exist, err := c.BlobExist(dstRepo, digest)
			if err != nil {
				return err
			}
			// the layer already exist, skip
			if exist {
				continue
			}
			// when the copy happens inside the same registry, use mount
			if err = c.MountBlob(srcRepo, digest, dstRepo); err != nil {
				return err
			}
		}
	}

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}
	// push manifest to the destination repository
	if _, err = c.PushManifest(dstRepo, dstRef, mediaType, payload); err != nil {
		return err
	}

	return nil
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	if c.authorizer != nil {
		if err := c.authorizer.Modify(req); err != nil {
			return nil, err
		}
	}
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), UserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		message := fmt.Sprintf("http status code: %d, body: %s", resp.StatusCode, string(body))
		code := errors.GeneralCode
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			code = errors.UnAuthorizedCode
		case http.StatusForbidden:
			code = errors.ForbiddenCode
		case http.StatusNotFound:
			code = errors.NotFoundCode
		}
		return nil, errors.New(nil).WithCode(code).
			WithMessage(message)
	}
	return resp, nil
}

// parse the next page link from the link header
func next(link string) string {
	links := lib.ParseLinks(link)
	for _, lk := range links {
		if lk.Rel == "next" {
			return lk.URL
		}
	}
	return ""
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

func buildCatalogURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/_catalog?n=1000", endpoint)
}

func buildTagListURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repository)
}

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", endpoint, repository, reference)
}

func buildBlobURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, reference)
}

func buildMountBlobURL(endpoint, repository, digest, from string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/?mount=%s&from=%s", endpoint, repository, digest, from)
}

func buildInitiateBlobUploadURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/", endpoint, repository)
}

func buildMonolithicBlobUploadURL(endpoint, location, digest string) (string, error) {
	url, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	q := url.Query()
	q.Set("digest", digest)
	url.RawQuery = q.Encode()
	if url.IsAbs() {
		return url.String(), nil
	}
	// the "relativeurls" is enabled in registry
	return endpoint + url.String(), nil
}
