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

package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/registryctl/auth"
	"github.com/goharbor/harbor/src/registryctl/config"
	"github.com/stretchr/testify/assert"
)

type fakeAuthenticator struct {
	err error
}

func (f *fakeAuthenticator) AuthorizeRequest(req *http.Request) error {
	return f.err
}

type fakeHandler struct {
	responseCode int
}

func (f *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(f.responseCode)
}

func TestNewAuthHandler(t *testing.T) {
	cases := []struct {
		authenticator auth.AuthenticationHandler
		handler       http.Handler
		insecureAPIs  map[string]bool
		responseCode  int
		requestURL    string
	}{
		{nil, nil, nil, http.StatusInternalServerError, "http://localhost/good"},
		{&fakeAuthenticator{err: nil}, nil, nil, http.StatusOK, "http://localhost/hello"},
		{&fakeAuthenticator{err: errors.New("error")}, nil, nil, http.StatusUnauthorized, "http://localhost/hello"},
		{&fakeAuthenticator{err: nil}, &fakeHandler{http.StatusNotFound}, nil, http.StatusNotFound, "http://localhost/notexsit"}, {&fakeAuthenticator{err: nil}, &fakeHandler{http.StatusOK}, map[string]bool{"/api/insecure": true}, http.StatusOK, "http://localhost/api/insecure"},
	}

	for _, c := range cases {
		handler := newAuthHandler(c.authenticator, c.handler, c.insecureAPIs)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", c.requestURL, nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, c.responseCode, w.Code, "unexpected response code")
	}
	cf := config.Configuration{}
	handler := NewHandlerChain(cf)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/api/health", nil)
	handler.ServeHTTP(w, r)

}
