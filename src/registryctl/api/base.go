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
	"encoding/json"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
)

// HandleInternalServerError ...
func HandleInternalServerError(w http.ResponseWriter, err error) {
	HandleError(w, errors.UnknownError(err))
}

// HandleNotMethodAllowed ...
func HandleNotMethodAllowed(w http.ResponseWriter) {
	HandleError(w, errors.MethodNotAllowedError(nil))
}

// HandleBadRequest ...
func HandleBadRequest(w http.ResponseWriter, err error) {
	HandleError(w, errors.BadRequestError(err))
}

// HandleError ...
func HandleError(w http.ResponseWriter, err error) {
	if _, ok := err.(driver.PathNotFoundError); ok {
		err = errors.New(nil).WithCode(errors.NotFoundCode).WithMessage(err.Error())
	}
	lib_http.SendError(w, err)
}

// WriteJSON response status code will be written automatically if there is an error
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		HandleInternalServerError(w, err)
		return err
	}

	if _, err = w.Write(b); err != nil {
		return err
	}
	return nil
}
