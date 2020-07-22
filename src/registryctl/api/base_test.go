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

package api

import (
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/goharbor/harbor/src/lib/errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleError(t *testing.T) {
	w := httptest.NewRecorder()
	HandleInternalServerError(w, errors.New("internal"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusInternalServerError)
	}

	w = httptest.NewRecorder()
	HandleBadRequest(w, errors.New("BadRequest"))
	if w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusBadRequest)
	}

	w = httptest.NewRecorder()
	HandleNotMethodAllowed(w)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusMethodNotAllowed)
	}

	w = httptest.NewRecorder()
	HandleError(w, errors.New("handle error"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusInternalServerError)
	}

	w = httptest.NewRecorder()
	HandleError(w, driver.PathNotFoundError{Path: "/blobstore/nonexist"})
	if w.Code != http.StatusNotFound {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusNotFound)
	}

}
