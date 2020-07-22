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

package bearer

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	cacheCapacity = 100
)

// NewAuthorizer return a bearer token authorizer
// The parameter "a" is an authorizer used to fetch the token
func NewAuthorizer(realm, service string, a lib.Authorizer, transport ...*http.Transport) lib.Authorizer {
	authorizer := &authorizer{
		realm:      realm,
		service:    service,
		authorizer: a,
		cache:      newCache(cacheCapacity),
	}
	tp := http.DefaultTransport
	if len(transport) > 0 && transport[0] != nil {
		tp = transport[0]
	}
	authorizer.client = &http.Client{Transport: tp}
	return authorizer
}

type authorizer struct {
	realm      string
	service    string
	authorizer lib.Authorizer
	cache      *cache
	client     *http.Client
}

func (a *authorizer) Modify(req *http.Request) error {
	// parse scopes from request
	scopes := parseScopes(req)

	// get token
	token, err := a.getToken(scopes)
	if err != nil {
		return err
	}

	// set authorization header
	if token != nil && len(token.Token) > 0 {
		req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token.Token))
	}
	return nil
}

func (a *authorizer) getToken(scopes []*scope) (*token, error) {
	// get token from cache first
	token := a.cache.get(scopes)
	if token != nil {
		return token, nil
	}

	// get no token from cache, fetch it from the token service
	token, err := a.fetchToken(scopes)
	if err != nil {
		return nil, err
	}

	// set the token into the cache
	a.cache.set(scopes, token)
	return token, nil
}

type token struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"` // the token returned by azure container registry is called "access_token"
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

func (a *authorizer) fetchToken(scopes []*scope) (*token, error) {
	url, err := url.Parse(a.realm)
	if err != nil {
		return nil, err
	}
	query := url.Query()
	query.Add("service", a.service)
	for _, scope := range scopes {
		query.Add("scope", scope.String())
	}
	url.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	if a.authorizer != nil {
		if err = a.authorizer.Modify(req); err != nil {
			return nil, err
		}
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		message := fmt.Sprintf("http status code: %d, body: %s", resp.StatusCode, string(body))
		code := errors.GeneralCode
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			code = errors.UnAuthorizedCode
		case http.StatusForbidden:
			code = errors.ForbiddenCode
		}
		return nil, errors.New(nil).WithCode(code).
			WithMessage(message)
	}
	token := &token{}
	if err = json.Unmarshal(body, token); err != nil {
		return nil, err
	}
	if len(token.Token) == 0 && len(token.AccessToken) > 0 {
		token.Token = token.AccessToken
	}
	return token, nil
}
