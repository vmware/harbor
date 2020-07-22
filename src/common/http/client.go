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

package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib"
)

const (
	// InsecureTransport used to get the insecure http Transport
	InsecureTransport = iota
	// SecureTransport used to get the external secure http Transport
	SecureTransport
)

var (
	secureHTTPTransport   *http.Transport
	insecureHTTPTransport *http.Transport
)

func init() {
	secureHTTPTransport = newDefaultTransport()
	insecureHTTPTransport = newDefaultTransport()
	insecureHTTPTransport.TLSClientConfig.InsecureSkipVerify = true

	if InternalTLSEnabled() {
		tlsConfig, err := GetInternalTLSConfig()
		if err != nil {
			panic(err)
		}
		secureHTTPTransport.TLSClientConfig = tlsConfig
	}
}

// Use this instead of Default Transport in library because it sets ForceAttemptHTTP2 to true
// And that options introduced in go 1.13 will cause the https requests hang forever in replication environment
func newDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig:       &tls.Config{},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// Client is a util for common HTTP operations, such Get, Head, Post, Put and Delete.
// Use Do instead if  those methods can not meet your requirement
type Client struct {
	modifiers []modifier.Modifier
	client    *http.Client
}

// GetHTTPTransport returns HttpTransport based on insecure configuration
func GetHTTPTransport(clientType uint) *http.Transport {
	switch clientType {
	case SecureTransport:
		return secureHTTPTransport
	case InsecureTransport:
		return insecureHTTPTransport
	default:
		// default Transport is secure one
		return secureHTTPTransport
	}
}

// GetHTTPTransportByInsecure returns a insecure HttpTransport if insecure is true or it returns secure one
func GetHTTPTransportByInsecure(insecure bool) *http.Transport {
	if insecure {
		return insecureHTTPTransport
	}
	return secureHTTPTransport
}

// NewClient creates an instance of Client.
// Use net/http.Client as the default value if c is nil.
// Modifiers modify the request before sending it.
func NewClient(c *http.Client, modifiers ...modifier.Modifier) *Client {
	client := &Client{
		client: c,
	}
	if client.client == nil {
		client.client = &http.Client{
			Transport: GetHTTPTransport(SecureTransport),
		}
	}
	if len(modifiers) > 0 {
		client.modifiers = modifiers
	}
	return client
}

// Do ...
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for _, modifier := range c.modifiers {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}

	return c.client.Do(req)
}

// Get ...
func (c *Client) Get(url string, v ...interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	data, err := c.do(req)
	if err != nil {
		return err
	}

	if len(v) == 0 {
		return nil
	}

	return json.Unmarshal(data, v[0])
}

// Head ...
func (c *Client) Head(url string) error {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

// Post ...
func (c *Client) Post(url string, v ...interface{}) error {
	var reader io.Reader
	if len(v) > 0 {
		if r, ok := v[0].(io.Reader); ok {
			reader = r
		} else {
			data, err := json.Marshal(v[0])
			if err != nil {
				return err
			}

			reader = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.do(req)
	return err
}

// Put ...
func (c *Client) Put(url string, v ...interface{}) error {
	var reader io.Reader
	if len(v) > 0 {
		data := []byte{}
		data, err := json.Marshal(v[0])
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.do(req)
	return err
}

// Delete ...
func (c *Client) Delete(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}

	return data, nil
}

// GetAndIteratePagination iterates the pagination header and returns all resources
// The parameter "v" must be a pointer to a slice
func (c *Client) GetAndIteratePagination(endpoint string, v interface{}) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v should be a pointer to a slice")
	}
	elemType := rv.Elem().Type()
	if elemType.Kind() != reflect.Slice {
		return errors.New("v should be a pointer to a slice")
	}

	resources := reflect.Indirect(reflect.New(elemType))
	for len(endpoint) > 0 {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return &Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		res := reflect.New(elemType)
		if err = json.Unmarshal(data, res.Interface()); err != nil {
			return err
		}
		resources = reflect.AppendSlice(resources, reflect.Indirect(res))

		endpoint = ""
		links := lib.ParseLinks(resp.Header.Get("Link"))
		for _, link := range links {
			if link.Rel == "next" {
				endpoint = url.Scheme + "://" + url.Host + link.URL
				break
			}
		}
	}
	rv.Elem().Set(resources)
	return nil
}
