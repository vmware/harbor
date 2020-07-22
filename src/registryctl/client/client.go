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

package client

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"io/ioutil"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/api/registry/gc"
)

// const definition
const (
	UserAgent = "harbor-registryctl-client"
)

// Client defines methods that an Registry client should implement
type Client interface {
	// Health tests the connection with registry server
	Health() error
	// StartGC enable the gc of registry server
	StartGC() (*gc.Result, error)
	// DeleteBlob deletes the specified blob. The "reference" should be "digest"
	DeleteBlob(reference string) (err error)
	// DeleteManifest deletes the specified manifest. The "reference" can be "tag" or "digest"
	DeleteManifest(repository, reference string) (err error)
}

type client struct {
	baseURL string
	client  *common_http.Client
}

// Config contains configurations needed for client
type Config struct {
	Secret string
}

// NewClient return an instance of Registry client
func NewClient(baseURL string, cfg *Config) Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	client := &client{
		baseURL: baseURL,
	}
	if cfg != nil {
		authorizer := auth.NewSecretAuthorizer(cfg.Secret)
		client.client = common_http.NewClient(&http.Client{
			Transport: common_http.GetHTTPTransport(common_http.SecureTransport),
		}, authorizer)
	}
	return client
}

// Health ...
func (c *client) Health() error {
	addr := strings.Split(c.baseURL, "://")[1]
	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}
	return utils.TestTCPConn(addr, 60, 2)
}

// StartGC ...
func (c *client) StartGC() (*gc.Result, error) {
	url := c.baseURL + "/api/registry/gc"
	gcr := &gc.Result{}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Failed to start gc: %d", resp.StatusCode)
		return nil, fmt.Errorf("Failed to start GC: %d", resp.StatusCode)
	}
	if err := json.Unmarshal(data, gcr); err != nil {
		return nil, err
	}

	return gcr, nil
}

// DeleteBlob ...
func (c *client) DeleteBlob(reference string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, buildBlobURL(c.baseURL, reference), nil)
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

// DeleteManifest ...
func (c *client) DeleteManifest(repository, reference string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, buildManifestURL(c.baseURL, repository, reference), nil)
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

func (c *client) do(req *http.Request) (*http.Response, error) {
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

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/api/registry/%s/manifests/%s", endpoint, repository, reference)
}

func buildBlobURL(endpoint, reference string) string {
	return fmt.Sprintf("%s/api/registry/blob/%s", endpoint, reference)
}
